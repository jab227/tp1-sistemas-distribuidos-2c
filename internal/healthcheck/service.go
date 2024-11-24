package healthcheck

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"
)

type HealthService struct {
	Port    int
	Timeout int
}

type RecvData struct {
	msg  string
	addr *net.UDPAddr
	err  error
}

func (h *HealthService) Run(ctx context.Context) error {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", h.Port))
	if err != nil {
		return fmt.Errorf("failed to resolve udp address: %w", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer conn.Close()

	recvCh := make(chan RecvData)
	go h.listen(ctx, conn, recvCh)

	for {
		select {
		case msg := <-recvCh:
			if msg.err != nil {
				return fmt.Errorf("failed to read response: %w", msg.err)
			}

			slog.Debug("health - received check msg", "msg", msg.msg, "addr", msg.addr)
			if err := SendOkMessage(conn, msg.addr); err != nil {
				return fmt.Errorf("failed to send ok message: %w", err)
			}
			slog.Debug("health - sent ok message", "addr", msg.addr)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (h *HealthService) listen(ctx context.Context, conn *net.UDPConn, ch chan<- RecvData) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, add, err := ReadCheckMSG(conn, time.Duration(h.Timeout)*time.Second)
			slog.Debug("health - read message", "msg", msg, "add", add, "err", err)
			if err != nil {
				// Timeout case
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				} else {
					// There was an error
					ch <- RecvData{err: err}
				}
			}

			// Send response
			ch <- RecvData{msg: msg, addr: add, err: nil}
		}
	}
}
