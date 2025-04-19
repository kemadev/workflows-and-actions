package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type jsonInfos struct {
	Type     string
	Mappings jsonToFindingsMappings
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

type jsonToFindingsMappings struct {
	ToolName  jsonToFindingsMapping
	RuleID    jsonToFindingsMapping
	Level     jsonToFindingsMapping
	FilePath  jsonToFindingsMapping
	StartLine jsonToFindingsMapping
	EndLine   jsonToFindingsMapping
	StartCol  jsonToFindingsMapping
	EndCol    jsonToFindingsMapping
	Message   jsonToFindingsMapping
}

// NOTE not 100% SARIF compliant (rule overrides / default config, ...), however sufficient for simple annotations using default values if key is not found
var sarifToFindingsMappings = jsonInfos{
	Mappings: jsonToFindingsMappings{
		ToolName:  jsonToFindingsMapping{Key: "runs[].tool.driver.name"},
		RuleID:    jsonToFindingsMapping{Key: "runs[].results[].ruleId"},
		Level:     jsonToFindingsMapping{Key: "runs[].results[].level"},
		FilePath:  jsonToFindingsMapping{Key: "runs[].results[].locations[].physicalLocation.artifactLocation.uri"},
		StartLine: jsonToFindingsMapping{Key: "runs[].results[].locations[].physicalLocation.region.startLine"},
		EndLine:   jsonToFindingsMapping{Key: "runs[].results[].locations[].physicalLocation.region.endLine"},
		StartCol:  jsonToFindingsMapping{Key: "runs[].results[].locations[].physicalLocation.region.startColumn"},
		EndCol:    jsonToFindingsMapping{Key: "runs[].results[].locations[].physicalLocation.region.endColumn"},
		Message:   jsonToFindingsMapping{Key: "runs[].results[].message.text"},
	},
}

const JsonStreamArrayKey = "jsonStreamArrayKey"

func FindingsFromJson(s string, i jsonInfos) ([]Finding, error) {
	if i.Type == "stream" {
		for _, line := range strings.Split(s, "\n") {
			if strings.TrimSpace(line) != "" {
				s += line + ","
			}
		}
		s = strings.TrimSuffix(s, ",")
		s = `{"` + JsonStreamArrayKey + `":[` + s + "]}"
	}
	m := i.Mappings
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(s), &jsonData); err != nil {
		return nil, fmt.Errorf("error unmarshalling json: %w", err)
	}

	var findings []Finding
	finding := Finding{}
	var err error
	if finding.ToolName, err = getValueFromMapping(jsonData, m.ToolName); err != nil {
		return nil, err
	}
	if finding.RuleID, err = getValueFromMapping(jsonData, m.RuleID); err != nil {
		return nil, err
	}
	if finding.Level, err = getValueFromMapping(jsonData, m.Level); err != nil {
		if err == keyNorFoundError {
			finding.Level = "warning"
		} else {
			return nil, err
		}
	}
	if finding.FilePath, err = getValueFromMapping(jsonData, m.FilePath); err != nil {
		return nil, err
	}
	if finding.StartLine, err = getIntValueFromMapping(jsonData, m.StartLine); err != nil {
		if err == keyNorFoundError {
			finding.StartLine = 1
		} else {
			return nil, err
		}
	}
	if finding.EndLine, err = getIntValueFromMapping(jsonData, m.EndLine); err != nil {
		if err == keyNorFoundError {
			finding.EndLine = finding.StartLine
		} else {
			return nil, err
		}
	}
	if finding.StartCol, err = getIntValueFromMapping(jsonData, m.StartCol); err != nil {
		if err == keyNorFoundError {
			finding.StartCol = 1
		} else {
			return nil, err
		}
	}
	if finding.EndCol, err = getIntValueFromMapping(jsonData, m.EndCol); err != nil {
		if err == keyNorFoundError {
			finding.EndCol = finding.StartCol
		} else {
			return nil, err
		}
	}
	if finding.Message, err = getValueFromMapping(jsonData, m.Message); err != nil {
		return nil, err
	}
	findings = append(findings, finding)
	return findings, nil
}

var keyNorFoundError error = errors.New("key not found")

func getValueFromMapping(data map[string]interface{}, m jsonToFindingsMapping) (string, error) {
	if m.OverrideKey != "" {
		return m.OverrideKey, nil
	}
	if m.GlobalSelectorRegex != "" {
		s, err := regexp.Compile(m.GlobalSelectorRegex)
		if err != nil {
			return "", err
		}
		if !s.MatchString(m.Key) {
			return "", keyNorFoundError
		}
	}
	parts := strings.Split(m.Key, ".")
	var current interface{} = data

	for _, part := range parts {
		if strings.HasSuffix(part, "[]") {
			part = strings.TrimSuffix(part, "[]")
			if array, ok := current.(map[string]interface{})[part].([]interface{}); ok && len(array) > 0 {
				current = array[0]
			} else {
				return "", errors.New("array not found or empty for part: " + part)
			}
		} else {
			if next, ok := current.(map[string]interface{})[part]; ok {
				current = next
			} else {
				return "", keyNorFoundError
			}
		}
	}

	// TODO check if null JSON type is handled correctly
	if current == nil {
		return "", nil
	}
	if str, ok := current.(string); ok {
		s, err := regexp.Compile(m.ValueTransformerRegex)
		if err != nil {
			return "", err
		}
		return s.FindString(str), nil
	}
	if num, ok := current.(int); ok {
		return strconv.Itoa(num), nil
	}
	if float, ok := current.(float64); ok {
		return strconv.FormatFloat(float, 'f', -1, 64), nil
	}
	if boolVal, ok := current.(bool); ok {
		if boolVal {
			return "true", nil
		}
		return "false", nil
	}
	return "", fmt.Errorf("value for mapping %s is not a string, int, float, or bool", m.Key)
}

func getIntValueFromMapping(data map[string]interface{}, m jsonToFindingsMapping) (int, error) {
	value, err := getValueFromMapping(data, m)
	if err != nil {
		return 1, err
	}
	return strconv.Atoi(value)
}
