package healthcheck

import (
	"fmt"
	"log/slog"
	"net"
)

type HealthService struct {
	Port int
}

func (h *HealthService) Run() error {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", h.Port))
	if err != nil {
		return fmt.Errorf("failed to resolve udp address: %w", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer conn.Close()

	for {
		response, addr, err := ReadCheckMSG(conn)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		slog.Debug("health - received check msg", "msg", response, "addr", addr)
		if err := SendOkMessage(conn, addr); err != nil {
			return fmt.Errorf("failed to send ok message: %w", err)
		}
		slog.Debug("health - sent ok message", "addr", addr)
	}
}
