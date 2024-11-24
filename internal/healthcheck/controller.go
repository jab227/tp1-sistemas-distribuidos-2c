package healthcheck

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type ControllerConfig struct {
	Port        int
	NodesPort   int
	ListOfNodes []string

	HealthCheckInterval int

	MaxRetries int
	MaxTimeout int
}

type NodeInfo struct {
	Count        int
	RecoveryMode bool
}

type State struct {
	mu    sync.Mutex
	nodes map[string]NodeInfo
}

func NewState(nodes []string) *State {
	s := &State{mu: sync.Mutex{}, nodes: make(map[string]NodeInfo)}

	for _, n := range nodes {
		s.nodes[n] = NodeInfo{}
	}

	return s
}

func (s *State) GetAllNodes() map[string]NodeInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	copiedNodes := make(map[string]NodeInfo, len(s.nodes))
	for k, v := range s.nodes {
		copiedNodes[k] = v
	}

	return copiedNodes
}

func (s *State) IncreaseRetries(node string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nodeInfo := s.nodes[node]
	nodeInfo.Count++
	s.nodes[node] = nodeInfo
}

func (s *State) ResetRetries(node string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nodeInfo := s.nodes[node]
	nodeInfo.Count = 0
	s.nodes[node] = nodeInfo
}

func (s *State) ToggleRecoveryMode(node string, mode bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nodeInfo := s.nodes[node]
	nodeInfo.RecoveryMode = mode
	s.nodes[node] = nodeInfo
}

type HealthController struct {
	Config ControllerConfig
	State  *State
}

func NewHealthController(config ControllerConfig) *HealthController {
	return &HealthController{Config: config, State: NewState(config.ListOfNodes)}
}

func (h *HealthController) Run() error {
	var wg sync.WaitGroup
	reviveCh := make(chan string)

	go h.RunReviver(reviveCh)

	for {
		for node, info := range h.State.GetAllNodes() {
			// TODO - Handle errors
			wg.Add(1)
			go h.CheckNode(node, info, reviveCh, &wg)
		}
		wg.Wait()
		time.Sleep(time.Duration(h.Config.HealthCheckInterval) * time.Second)
	}
}

func (h *HealthController) RunReviver(reviveCh <-chan string) {
	for {
		select {
		case node := <-reviveCh:
			fmt.Printf("Reviving node %s\n", node)
			if err := RestartNode(node); err != nil {
				fmt.Printf("Error restarting node %s: %s (This will be retried later)\n", node, err)
			} else {
				h.State.ResetRetries(node)
			}

			h.State.ToggleRecoveryMode(node, false)
		}
	}
}

// TODO - Use logs
func (h *HealthController) CheckNode(node string, nodeInfo NodeInfo, reviveCh chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Checking if node %s is alive\n", node)

	// Check if limit was obtained or the service is down and being revived
	if nodeInfo.Count >= h.Config.MaxRetries && !nodeInfo.RecoveryMode {
		fmt.Printf("Node %s reached a limit - Sending for revive\n", node)
		h.State.ToggleRecoveryMode(node, true)

		// Send node to be revived
		reviveCh <- node
		return
	} else if nodeInfo.RecoveryMode {
		fmt.Printf("Node %s is set as down temporarly\n", node)
		return
	}

	// Creates Dial for UDP messaging
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", node, h.Config.NodesPort))
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		// Check if the service is available
		if netErr, ok := err.(*net.OpError); ok && strings.Contains(netErr.Err.Error(), "missing address") {
			fmt.Printf("Node %s is down\n", node)
			h.State.IncreaseRetries(node)
			return

		} else {
			fmt.Printf("failed to connect to %s: %s\n", node, err)
			return
		}
	}
	defer conn.Close()

	// Sends check alive message
	err = SendCheckMessage(conn)
	if err != nil {
		fmt.Printf("failed to write to %s: %s\n", node, err)
		return
	}

	// Set Timeout for recv the node response
	if err := conn.SetReadDeadline(time.Now().Add(time.Duration(h.Config.MaxTimeout) * time.Second)); err != nil {
		fmt.Printf("failed to set read deadline: %s", err)
		return
	}
	response, addr, err := ReadResponseMessage(conn)
	if err != nil {
		//
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Printf("Node %s didn't respond\n", node)
			h.State.IncreaseRetries(node)
		} else {
			fmt.Printf("failed to read from %s: %s\n", node, err)
			return
		}
	} else {
		fmt.Printf("Received data from %s: %s\n", addr.String(), response)
		h.State.ResetRetries(node)
	}
}
