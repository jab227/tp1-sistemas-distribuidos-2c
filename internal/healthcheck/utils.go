package healthcheck

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
)

func RestartNode(node string) error {
	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (conn net.Conn, er error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}
	defer client.CloseIdleConnections()

	restartURL := fmt.Sprintf("http://xxxx.xxx/containers/%s/restart", node)
	req, err := http.NewRequest("POST", restartURL, nil)
	if err != nil {
		return err
	}

	response, err := client.Do(req)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(response.Body)
		defer response.Body.Close()
		return fmt.Errorf("response: %s", body)
	}

	return nil
}
