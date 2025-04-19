package main

import (
	"encoding/json"
	"strings"
)

type Finding struct {
	ToolName  string
	RuleID    string
	Level     string
	FilePath  string
	StartLine int
	EndLine   int
	StartCol  int
	EndCol    int
	Message   string
}

type jsonToFindingsMapping struct {
	// JSON key to find
	Key string
	// Do not try to find key and use this value instead
	OverrideKey string
	// Transform found value using this regex
	ValueTransformerRegex string
	// Discard whole finding if this regex for current key does not match
	GlobalSelectorRegex string
}

type jsonInfos struct {
	Type     string
	Mappings jsonToFindingsMappings
}

const JsonStreamArrayKey = "jsonStreamArrayKey"

func FindingsFromJson(s string, i jsonInfos) ([]Finding, error) {
	if i.Type == "stream" {
		for _, line := range strings.Split(s, "\n") {
			if strings.TrimSpace(line) != "" {
				s += line + ","
			}
		}
		s = "[" + strings.TrimSuffix(s, ",") + "]"
	}

	j := []byte()
	err := json.Unmarshal([]byte(s), &i.Mappings)
	if err != nil {
		return nil, err
	}
}
