package env

import (
	"strconv"
	"strings"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
)

const OutputRouterTags = "OUTPUT_ROUTER_TAGS"

func GetRouterTags() ([]string, error) {
	tags, err := utils.GetFromEnv(OutputRouterTags)
	if err != nil {
		return nil, err
	}

	tagsList := strings.Split(*tags, ",")
	return tagsList, nil
}

const (
	RouterRoundRobin = "round-robin"
	RouterID         = "id"
	RouterGameReview = "game-review"
)

func GetRouterTypeFromEnv() string {
	routerType, err := utils.GetFromEnv("ROUTER_TYPE")
	if err != nil {
		return RouterID
	}
	return *routerType
}

func GetIsProjection() (bool, error) {
	isProjectionStr, err := utils.GetFromEnv("IS_PROJECTION")
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(*isProjectionStr)
}
