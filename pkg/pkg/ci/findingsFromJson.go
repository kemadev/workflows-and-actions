package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type jsonInfos struct {
	Type     string
	Mappings jsonToFindingsMappings
}

type jsonToFindingsMapping struct {
	Key           string
	OverrideKey   string
	SelectorRegex string
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

// not 100% SARIF compliant (rule overrides / default config, ...), however sufficient for simple annotations using default values if key is not found
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
	if finding.ToolName, err = getValueFromMapping(jsonData, m.ToolName.Key, m.ToolName.OverrideKey); err != nil {
		return nil, err
	}
	if finding.RuleID, err = getValueFromMapping(jsonData, m.RuleID.Key, m.RuleID.OverrideKey); err != nil {
		return nil, err
	}
	if finding.Level, err = getValueFromMapping(jsonData, m.Level.Key, m.Level.OverrideKey); err != nil {
		if err == keyNorFoundError {
			finding.Level = "warning"
		} else {
			return nil, err
		}
	}
	if finding.FilePath, err = getValueFromMapping(jsonData, m.FilePath.Key, m.FilePath.OverrideKey); err != nil {
		return nil, err
	}
	if finding.StartLine, err = getIntValueFromMapping(jsonData, m.StartLine.Key, m.StartLine.OverrideKey); err != nil {
		if err == keyNorFoundError {
			finding.StartLine = 1
		} else {
			return nil, err
		}
	}
	if finding.EndLine, err = getIntValueFromMapping(jsonData, m.EndLine.Key, m.EndLine.OverrideKey); err != nil {
		if err == keyNorFoundError {
			finding.EndLine = finding.StartLine
		} else {
			return nil, err
		}
	}
	if finding.StartCol, err = getIntValueFromMapping(jsonData, m.StartCol.Key, m.StartCol.OverrideKey); err != nil {
		if err == keyNorFoundError {
			finding.StartCol = 1
		} else {
			return nil, err
		}
	}
	if finding.EndCol, err = getIntValueFromMapping(jsonData, m.EndCol.Key, m.EndCol.OverrideKey); err != nil {
		if err == keyNorFoundError {
			finding.EndCol = finding.StartCol
		} else {
			return nil, err
		}
	}
	if finding.Message, err = getValueFromMapping(jsonData, m.Message.Key, m.Message.OverrideKey); err != nil {
		return nil, err
	}
	findings = append(findings, finding)
	return findings, nil
}

var keyNorFoundError error = errors.New("key not found")

func getValueFromMapping(data map[string]interface{}, mapping string, override string) (string, error) {
	if mapping == "" {
		if override != "" {
			return override, nil
		}
		return "", keyNorFoundError
	}
	parts := strings.Split(mapping, ".")
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
		return str, nil
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
	return "", fmt.Errorf("value for mapping %s is not a string, int, float, or bool", mapping)
}

func getIntValueFromMapping(data map[string]interface{}, mapping string, override string) (int, error) {
	value, err := getValueFromMapping(data, mapping, override)
	if err != nil {
		return 1, err
	}
	return strconv.Atoi(value)
}
