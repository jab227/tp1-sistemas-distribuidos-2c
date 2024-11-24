package healthcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
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

func GetDockerNodes(excluded []string, network string) ([]string, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	url := fmt.Sprintf("http://unix/containers/json?all=true&filter={\"network\":\"%s\"}", network)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	var containers []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	var names []string
	for _, container := range containers {
		if nameList, ok := container["Names"].([]interface{}); ok && len(nameList) > 0 {
			if name, ok := nameList[0].(string); ok {
				contains := false
				cleanName := strings.TrimPrefix(name, "/")
				for _, node := range excluded {
					if cleanName == node {
						contains = true
						break
					}
				}

				if contains {
					continue
				}

				names = append(names, cleanName)
			}
		}
	}

	return names, nil
}
