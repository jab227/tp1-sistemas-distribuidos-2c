package env

import "strings"

const OutputRouterTags = "OUTPUT_ROUTER_TAGS"

func GetRouterTags() ([]string, error) {
	tags, err := GetFromEnv(OutputRouterTags)
	if err != nil {
		return nil, err
	}

	tagsList := strings.Split(*tags, ",")
	return tagsList, nil
}
