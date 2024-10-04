package env

import (
	"fmt"
	"os"
	"strconv"
)

func GetFromEnv(name string) (*string, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return nil, fmt.Errorf("environment variable %s not found", name)
	}

	return &value, nil
}

func GetFromEnvUint(name string) (*uint64, error) {
	value, err := GetFromEnv(name)
	if err != nil {
		return nil, err
	}

	valueUint, err := strconv.ParseUint(*value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("environment variable %s is not a valid uint: %s", name, err)
	}

	return &valueUint, nil
}

func GetFromEnvInt(name string) (*int64, error) {
	value, err := GetFromEnv(name)
	if err != nil {
		return nil, err
	}

	valueInt, err := strconv.ParseInt(*value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("environment variable %s is not a valid uint: %s", name, err)
	}

	return &valueInt, nil
}
