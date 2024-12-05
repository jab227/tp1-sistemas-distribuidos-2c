package healthcheck

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/healthcheck/leader"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/healthcheck/leader/middleware"
)

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
	Config *ControllerConfig
	State  *State
	done   chan struct{}
}

func NewHealthController(config *ControllerConfig) *HealthController {
	return &HealthController{Config: config, State: NewState(config.ListOfNodes)}
}

func (h *HealthController) Done() <-chan struct{} {
	return h.done
}

type electionState struct {
	NeighbourIDs   []int
	NodeID         int
	LeaderID       int
	MaxRetries     int
	Port           int
	CheckInterval  time.Duration
	RequestTimeout time.Duration
	State          leader.State
}

func newElectionState(
	nodeID int,
	neighbourIDs []int,
	maxRetries,
	port int,
	checkInterval,
	requestTimeout time.Duration,
) electionState {
	return electionState{
		NeighbourIDs:   neighbourIDs,
		NodeID:         nodeID,
		LeaderID:       -1,
		MaxRetries:     maxRetries,
		Port:           port,
		CheckInterval:  checkInterval,
		RequestTimeout: requestTimeout,
		State:          leader.StateElecting,
	}
}

func (e electionState) isLeader() bool {
	isLeader := true
	for _, neighbourID := range e.NeighbourIDs {
		if e.NodeID < neighbourID {
			isLeader = false
			break
		}
	}
	return isLeader
}

func (e *electionState) handleElection(ctx context.Context, m *middleware.LeaderMiddleware, reader <-chan middleware.Message) error {
	slog.Debug("state electing")
	if e.isLeader() {
		slog.Debug("I have the biggest ID so im the leader")
		if err := sendCoordinatorToAll(ctx, m, e.NodeID, e.NeighbourIDs); err != nil {
			return err
		}
		// Become the leader
		e.State = leader.StateLeading
		e.LeaderID = e.NodeID
	} else {
		slog.Debug("starting election")
		if err := startElection(ctx, m, e.NodeID, e.NeighbourIDs); err != nil {
			return err
		}
		var receivedAlive bool
	Exit:
		for {
			select {
			case <-time.After(8 * time.Second):
				if receivedAlive {
					// received alive, so im expected to wait for a coordinator
					// reset variable, so the next time it triggers I will assume that im the leader
					receivedAlive = false
					continue
				}
				// Send coordinator message
				if err := sendCoordinatorToAll(ctx, m, e.NodeID, e.NeighbourIDs); err != nil {
					return err
				}
				// Become the leader

				e.State = leader.StateLeading
				e.LeaderID = e.NodeID
				break Exit
			case msg := <-reader:
				body := msg.GetBody()
				slog.Debug("electing", "body", body)
				var electionMessage leader.ElectionMessage
				electionMessage.Unmarshal(body)
				switch electionMessage.Type {
				case leader.ElectionMessageTypeAnnounceElection:
					slog.Debug("received announce election message", "from", electionMessage.ID)
					// Send alive message
					sendAliveTo(ctx, m, e.NodeID, int(electionMessage.ID))
					msg.Acknowledge()
				case leader.ElectionMessageTypeAlive:
					slog.Debug("received alive message", "from", electionMessage.ID)
					receivedAlive = true
					msg.Acknowledge()
				case leader.ElectionMessageTypeCoordinator:
					slog.Debug("received coordinator message", "leader", electionMessage.ID)
					e.State = leader.StateMonitoring
					e.LeaderID = int(electionMessage.ID)
					msg.Acknowledge()
					break Exit
				default:
					slog.Debug("default case shouldn't run", "msg", electionMessage)
					panic("here")
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return nil
}

func (e *electionState) handleMonitoring(ctx context.Context, m *middleware.LeaderMiddleware, reader <-chan middleware.Message) error {
	slog.Debug("state monitoring")
	slog.Debug("elections state", "value", *e)
Exit:
	for {
		select {
		case <-time.After(e.CheckInterval):
			retries := e.MaxRetries
			for retries > 0 {
				leaderStr := fmt.Sprintf("healthcheck_%d", e.LeaderID)
				if !CheckLeader(leaderStr, e.Port, e.RequestTimeout) {
					retries--
				} else {
					break
				}
			}
			if retries <= 0 {
				slog.Debug("leader is dead, starting election")
				e.State = leader.StateElecting
				break Exit
			}
		case msg := <-reader:
			body := msg.GetBody()
			var electionMessage leader.ElectionMessage
			electionMessage.Unmarshal(body)
			switch electionMessage.Type {
			case leader.ElectionMessageTypeAnnounceElection:
				slog.Debug("received announce election message", "from", electionMessage.ID)
				// Send alive message
				sendAliveTo(ctx, m, e.NodeID, int(electionMessage.ID))
				e.State = leader.StateElecting
				msg.Acknowledge()
				break Exit
			case leader.ElectionMessageTypeAlive:
				msg.Acknowledge()
				break Exit
			case leader.ElectionMessageTypeCoordinator:
				slog.Debug("received coordinator message", "leader", electionMessage.ID)
				e.LeaderID = int(electionMessage.ID)
				msg.Acknowledge()
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (e *electionState) handleLeading(
	ctx context.Context,
	m *middleware.LeaderMiddleware,
	h *HealthController,
	reader <-chan middleware.Message,
	reviveCh chan string,
) error {
	slog.Debug("state leading")
	var wg sync.WaitGroup
Exit:
	for {
		select {
		case <-time.After(e.CheckInterval):
			for node, info := range h.State.GetAllNodes() {
				wg.Add(1)
				go h.CheckNode(node, info, reviveCh, &wg)
			}
			wg.Wait()
		case msg := <-reader:
			body := msg.GetBody()
			slog.Debug("leading", "body", body)
			var electionMessage leader.ElectionMessage
			electionMessage.Unmarshal(body)
			switch electionMessage.Type {
			case leader.ElectionMessageTypeAnnounceElection:
				slog.Debug("received announce election message", "from", electionMessage.ID)
				// Send alive message
				sendAliveTo(ctx, m, e.NodeID, int(electionMessage.ID))
				e.State = leader.StateElecting
				msg.Acknowledge()
				break Exit
			case leader.ElectionMessageTypeAlive:
				slog.Debug("shouldn't happen")
				msg.Acknowledge()
			case leader.ElectionMessageTypeCoordinator:
				if electionMessage.ID > uint32(e.NodeID) {
					slog.Debug("received coordinator message", "leader", electionMessage.ID)
					e.State = leader.StateMonitoring
					e.LeaderID = int(electionMessage.ID)
					msg.Acknowledge()
					break Exit
				}
				msg.Acknowledge()
			default:
				slog.Debug("default case shouldn't run", "msg", electionMessage)
				panic("here")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (h *HealthController) Run(ctx context.Context) error {
	defer func() {
		h.done <- struct{}{}
	}()

	reviveCh := make(chan string, 1)
	id := h.Config.ID
	options, err := middleware.GetOptionsFromEnv()
	if err != nil {
		return err
	}
	m, err := middleware.NewLeaderMiddleware(id, &options)
	if err != nil {
		return err
	}
	defer m.Close()
	s := newElectionState(
		h.Config.ID,
		h.Config.NeighboursIDs,
		h.Config.MaxRetries,
		h.Config.Port,
		time.Duration(h.Config.HealthCheckInterval)*time.Second,
		time.Duration(h.Config.MaxTimeout)*time.Second,
	)

	go h.RunReviver(ctx, reviveCh)
	reader, err := m.Reader()
	if err != nil {
		return err
	}
	for {
		switch s.State {
		case leader.StateElecting:
			if err := s.handleElection(ctx, m, reader); err != nil {
				return err
			}
		case leader.StateMonitoring:
			if err := s.handleMonitoring(ctx, m, reader); err != nil {
				return err
			}
		case leader.StateLeading:
			if err := s.handleLeading(ctx, m, h, reader, reviveCh); err != nil {
				return err
			}
		}
	}
}

func startElection(ctx context.Context, m *middleware.LeaderMiddleware, id int, ids []int) error {
	for _, neighbourID := range ids {
		if neighbourID < id {
			continue
		}
		msg := leader.ElectionMessage{
			ID:   uint32(id),
			Type: leader.ElectionMessageTypeAnnounceElection,
		}
		err := func() error {
			slog.Debug("sending election message", "to", neighbourID)
			timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := m.Write(timeoutCtx, msg.Marshal(), neighbourID); err != nil {
				return err

			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func sendCoordinatorToAll(ctx context.Context, m *middleware.LeaderMiddleware, id int, ids []int) error {
	for _, neighbourID := range ids {
		msg := leader.ElectionMessage{
			ID:   uint32(id),
			Type: leader.ElectionMessageTypeCoordinator,
		}
		err := func() error {
			slog.Debug("sending coordinator to all", "id", neighbourID)
			timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := m.Write(timeoutCtx, msg.Marshal(), neighbourID); err != nil {
				return err

			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func sendAliveTo(ctx context.Context, m *middleware.LeaderMiddleware, src, dst int) error {
	msg := leader.ElectionMessage{
		ID:   uint32(src),
		Type: leader.ElectionMessageTypeAlive,
	}
	err := func() error {
		timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := m.Write(timeoutCtx, msg.Marshal(), dst); err != nil {
			return err

		}
		return nil
	}()
	if err != nil {
		return err
	}
	return nil
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
	_, _, err = ReadOkMSG(conn, time.Duration(h.Config.MaxTimeout)*time.Second)
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
		//		slog.Debug("received response from node", "node", node, "response", response)
		h.State.ResetRetries(node)
	}
}

func CheckLeader(node string, port int, timeout time.Duration) bool {
	// Creates Dial for UDP messaging
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", node, port))
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		// Check if the service is available
		if netErr, ok := err.(*net.OpError); ok && strings.Contains(netErr.Err.Error(), "missing address") {
			slog.Info("couldn't connect to node, the container could be down", "node", node)
			return false

		} else {
			slog.Error("failed to connect to node", "node", node, "error", err.Error())
			return false
		}
	}
	defer conn.Close()

	// Sends check alive message
	err = SendCheckMessage(conn)
	if err != nil {
		slog.Error("failed to send check message", "node", node, "error", err.Error())
		return false
	}

	// Set Timeout for recv the node response and wait to recv
	_, _, err = ReadOkMSG(conn, timeout)
	if err != nil {
		return false
	}
	return true
}
