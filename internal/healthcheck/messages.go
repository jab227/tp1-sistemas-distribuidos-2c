package healthcheck

import "net"

const MaxSizeMessage = 1024
const CheckMSG = "CHECK"
const OkMSG = "OK"

func SendCheckMessage(conn *net.UDPConn) error {
	_, err := conn.Write([]byte(CheckMSG))
	return err
}

// TODO - Handle Short Write
func SendOkMessage(conn *net.UDPConn, addr *net.UDPAddr) error {
	_, err := conn.WriteToUDP([]byte(OkMSG), addr)
	return err
}

// TODO - Handle Short Read
func ReadResponseMessage(conn *net.UDPConn) (string, *net.UDPAddr, error) {
	buf := make([]byte, MaxSizeMessage)
	_, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return "", nil, err
	}

	return string(buf), addr, nil
}
