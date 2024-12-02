package healthcheck

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

const PortEnvVariable = "HEALTHCHECK_PORT"
const NodesPortEnvVariable = "HEALTHCHECK_NODES_PORT"
const CheckIntervalEnvVariable = "HEALTHCHECK_INTERVAL"
const CheckTimeoutEnvVariable = "HEALTHCHECK_TIMEOUT"
const MaxCheckRetriesEnvVariable = "HEALTHCHECK_MAX_RETRIES"
const NodeIDEnvVariable = "HEALTHCHECK_NODE_ID"
const NeighboursEnvVariable = "HEALTHCHECK_NEIGHBOURS"

const PortDefaultValue = 1516
const CheckIntervalDefaultValue = 2
const TimeoutDefaultValue = 2
const MaxCheckRetriesDefaultValue = 3

type ControllerConfig struct {
	ID int
	// Add env
	NeighboursIDs []int
	Port          int
	NodesPort     int
	ListOfNodes   []string

	HealthCheckInterval int
	MaxTimeout          int

	MaxRetries int
}

func NewHealthConfigFromEnv(nodes []string) (*ControllerConfig, error) {
	c := ControllerConfig{ListOfNodes: nodes}

	value, err := utils.GetFromEnvInt(PortEnvVariable)
	if err != nil {
		slog.Info("using default value for configuration", "variable", PortEnvVariable, "value", PortDefaultValue)
		c.Port = PortDefaultValue
	} else {
		c.Port = int(*value)
	}

	value, err = utils.GetFromEnvInt(NodesPortEnvVariable)
	if err != nil {
		slog.Info("using default value for configuration", "variable", NodesPortEnvVariable, "value", PortDefaultValue)
		c.NodesPort = PortDefaultValue
	} else {
		c.NodesPort = int(*value)
	}

	value, err = utils.GetFromEnvInt(CheckIntervalEnvVariable)
	if err != nil {
		slog.Info("using default value for configuration", "variable", CheckIntervalEnvVariable, "value", CheckIntervalDefaultValue)
		c.HealthCheckInterval = CheckIntervalDefaultValue
	} else {
		c.HealthCheckInterval = int(*value)
	}

	value, err = utils.GetFromEnvInt(CheckTimeoutEnvVariable)
	if err != nil {
		slog.Info("using default value for configuration", "variable", CheckTimeoutEnvVariable, "value", TimeoutDefaultValue)
		c.MaxTimeout = TimeoutDefaultValue
	} else {
		c.MaxTimeout = int(*value)
	}

	value, err = utils.GetFromEnvInt(MaxCheckRetriesEnvVariable)
	if err != nil {
		slog.Info("using default value for configuration", "variable", MaxCheckRetriesEnvVariable, "value", MaxCheckRetriesDefaultValue)
		c.MaxRetries = MaxCheckRetriesDefaultValue
	} else {
		c.MaxRetries = int(*value)
	}

	value, err = utils.GetFromEnvInt(NodeIDEnvVariable)
	if err != nil {
		return nil, fmt.Errorf("couldn't set node id: %w", err)
	}
	c.ID = int(*value)
	neighboursStr, err := utils.GetFromEnv(NeighboursEnvVariable)
	if err != nil {
		return nil, fmt.Errorf("couldn't set neighbours: %w", err)
	}
	neighboursSplit := strings.Split(*neighboursStr, ",")
	var neighbours []int
	for _, n := range neighboursSplit {
		id, err := strconv.Atoi(n)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse neighbour id: %w", err)
		}
		neighbours = append(neighbours, id)
	}
	c.NeighboursIDs = neighbours
	return &c, nil
}

const ServicePortEnvVariable = "HEALTH_SERVICE_PORT"
const ServiceTimeoutEnvVariable = "HEALTH_SERVICE_TIMEOUT"

const ServicePortDefaultValue = 1516
const ServiceTimeoutDefaultValue = 2

type HealthServiceConfig struct {
	Port    int
	Timeout int
}

func NewHealthServiceConfigFromEnv() *HealthServiceConfig {
	c := HealthServiceConfig{}

	value, err := utils.GetFromEnvInt(ServicePortEnvVariable)
	if err != nil {
		slog.Info("using default value for configuration", "variable", ServicePortEnvVariable, "value", ServicePortDefaultValue)
		c.Port = ServicePortDefaultValue
	} else {
		c.Port = int(*value)
	}

	value, err = utils.GetFromEnvInt(ServiceTimeoutEnvVariable)
	if err != nil {
		slog.Info("using default value for configuration", "variable", ServiceTimeoutEnvVariable, "value", ServiceTimeoutDefaultValue)
		c.Timeout = TimeoutDefaultValue
	} else {
		c.Timeout = int(*value)
	}

	return &c
}

const ExcludedNodesEnvVariable = "HEALTH_EXCLUDED_NODES"
const HealthNetworkEnvVariable = "HEALTH_NETWORK"

type DockerDiscoveryConfig struct {
	Excluded []string
	Network  string
}

func NewDockerDiscoveryConfigFromEnv() (*DockerDiscoveryConfig, error) {
	nodes, err := utils.GetFromEnv(ExcludedNodesEnvVariable)
	if err != nil {
		return nil, fmt.Errorf("could not get exclude nodes %s: %w", ExcludedNodesEnvVariable, err)
	}

	nodesList := strings.Split(*nodes, ",")

	network, err := utils.GetFromEnv(HealthNetworkEnvVariable)
	if err != nil {
		return nil, fmt.Errorf("could not get health network %s: %w", HealthNetworkEnvVariable, err)
	}

	return &DockerDiscoveryConfig{
		Excluded: nodesList,
		Network:  *network,
	}, nil
}
