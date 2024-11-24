package healthcheck

import (
	"context"
	"fmt"
	"log/slog"
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
	done   chan struct{}
}

func NewHealthController(config ControllerConfig) *HealthController {
	return &HealthController{Config: config, State: NewState(config.ListOfNodes)}
}

func (h *HealthController) Done() <-chan struct{} {
	return h.done
}

func (h *HealthController) Run(ctx context.Context) error {
	defer func() {
		h.done <- struct{}{}
	}()

	var wg sync.WaitGroup
	reviveCh := make(chan string)

	go h.RunReviver(ctx, reviveCh)

	for {
		select {
		case <-time.After(time.Second * time.Duration(h.Config.HealthCheckInterval)):
			for node, info := range h.State.GetAllNodes() {
				wg.Add(1)
				go h.CheckNode(node, info, reviveCh, &wg)
			}
			wg.Wait()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (h *HealthController) RunReviver(ctx context.Context, reviveCh <-chan string) {
	for {
		select {
		case node := <-reviveCh:
			slog.Info("reviving node", "node", node)
			if err := RestartNode(node); err != nil {
				slog.Error("Error restarting node", "node", node, "error", err.Error())
			} else {
				h.State.ResetRetries(node)
			}

			h.State.ToggleRecoveryMode(node, false)
		case <-ctx.Done():
			return
		}
	}
}

func (h *HealthController) CheckNode(node string, nodeInfo NodeInfo, reviveCh chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Check if limit was obtained or the service is down and being revived
	if nodeInfo.Count >= h.Config.MaxRetries && !nodeInfo.RecoveryMode {
		slog.Info("node reached a limit, sending it to be revived", "node", node)
		h.State.ToggleRecoveryMode(node, true)

		// Send node to be revived
		reviveCh <- node
		return
	} else if nodeInfo.RecoveryMode {
		slog.Info("node is temporarily down and should be restarted soon", "node", node)
		return
	}

	// Creates Dial for UDP messaging
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", node, h.Config.NodesPort))
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		// Check if the service is available
		if netErr, ok := err.(*net.OpError); ok && strings.Contains(netErr.Err.Error(), "missing address") {
			slog.Info("couldn't connect to node, the container could be down", "node", node)
			h.State.IncreaseRetries(node)
			return

		} else {
			slog.Error("failed to connect to node", "node", node, "error", err.Error())
			return
		}
	}
	defer conn.Close()

	// Sends check alive message
	err = SendCheckMessage(conn)
	if err != nil {
		slog.Error("failed to send check message", "node", node, "error", err.Error())
		return
	}

	// Set Timeout for recv the node response and wait to recv
	response, _, err := ReadOkMSG(conn, time.Duration(h.Config.MaxTimeout)*time.Second)
	if err != nil {
		//
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			slog.Info("node didn't respond within timeout period", "node", node, "error", err.Error())
			h.State.IncreaseRetries(node)
		} else {
			slog.Error("failed to read from node", "node", node, "error", err.Error())
			return
		}
	} else {
		slog.Debug("received response from node", "node", node, "response", response)
		h.State.ResetRetries(node)
	}
}
