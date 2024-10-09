package env

import (
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
