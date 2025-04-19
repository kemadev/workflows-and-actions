package main

import (
	"encoding/json"
	"fmt"
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

type jsonToFindingsMappings struct {
	ToolName  jsonMappingInfo
	RuleID    jsonMappingInfo
	Level     jsonMappingInfo
	FilePath  jsonMappingInfo
	StartLine jsonMappingInfo
	EndLine   jsonMappingInfo
	StartCol  jsonMappingInfo
	EndCol    jsonMappingInfo
	Message   jsonMappingInfo
}

type jsonMappingInfo struct {
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

func assignValue(overrideKey string, value interface{}, defaultValue interface{}) interface{} {
	if overrideKey != "" {
		return overrideKey
	}
	if value == nil {
		return defaultValue
	}
	return value
}

func FindingsFromJson(s string, i jsonInfos) ([]Finding, error) {
	if i.Type == "stream" {
		for _, line := range strings.Split(s, "\n") {
			if strings.TrimSpace(line) != "" {
				s += line + ","
			}
		}
		s = "[" + strings.TrimSuffix(s, ",") + "]"
	}

	var j interface{}
	if err := json.Unmarshal([]byte(s), &j); err != nil {
		return nil, fmt.Errorf("error unmarshalling json: %w", err)
	}

	a, ok := j.([]interface{})
	if !ok {
		return nil, fmt.Errorf("json is not an array")
	}

	var findings []Finding
	for _, item := range a {
		m, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("json item is not a map")
		}
		f, err := findingFromJsonObject(m, i.Mappings)
		if err != nil {
			return nil, fmt.Errorf("error parsing json object: %w", err)
		}
		findings = append(findings, f)
	}
	return findings, nil
}

func findingFromJsonObject(m map[string]interface{}, mappings jsonToFindingsMappings) (Finding, error) {
	f := Finding{}
	def := setDefaultValue(m, mappings.ToolName.Key, mappings.ToolName.OverrideKey, "unknown")
	return Finding{}, nil
}

func setDefaultValue(v interface{}, overrideKey string, defaultValue string) interface{} {
	if overrideKey != "" {
		return overrideKey
	}
	if v == nil {
		return defaultValue
	}
	return v
}
