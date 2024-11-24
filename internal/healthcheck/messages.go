package healthcheck

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"net"
	"time"
)

const CheckMSG = "CHECK"
const OkMSG = "OK"

func SendCheckMessage(conn *net.UDPConn) error {
	return utils.WriteToSocket(conn, []byte(CheckMSG))

}

func SendOkMessage(conn *net.UDPConn, addr *net.UDPAddr) error {
	return utils.WriteToUDPSocket(conn, []byte(OkMSG), addr)
}

func ReadCheckMSG(conn *net.UDPConn, timeout time.Duration) (string, *net.UDPAddr, error) {
	buf := make([]byte, len(CheckMSG))
	addr, err := utils.ReadFromUDPSocketWithTimeout(conn, &buf, len(CheckMSG), timeout)
	if err != nil {
		return "", nil, err
	}

	return string(buf), addr, nil
}

func ReadOkMSG(conn *net.UDPConn, timeout time.Duration) (string, *net.UDPAddr, error) {
	buf := make([]byte, len(OkMSG))
	addr, err := utils.ReadFromUDPSocketWithTimeout(conn, &buf, len(OkMSG), timeout)
	if err != nil {
		return "", nil, err
	}

	return string(buf), addr, nil
}
