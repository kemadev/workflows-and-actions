package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type jsonToFindingsMappings struct {
	ToolName  string
	RuleID    string
	Level     string
	FilePath  string
	StartLine string
	EndLine   string
	StartCol  string
	EndCol    string
	Message   string
}

// not 100% SARIF compliant (rule overrides / default config, ...), however sufficient for simple annotations using default values if key is not found
var sarifToFindingsMappings = jsonToFindingsMappings{
	ToolName:  "runs[].tool.driver.name",
	RuleID:    "runs[].results[].ruleId",
	Level:     "runs[].results[].level",
	FilePath:  "runs[].results[].locations[].physicalLocation.artifactLocation.uri",
	StartLine: "runs[].results[].locations[].physicalLocation.region.startLine",
	EndLine:   "runs[].results[].locations[].physicalLocation.region.endLine",
	StartCol:  "runs[].results[].locations[].physicalLocation.region.startColumn",
	EndCol:    "runs[].results[].locations[].physicalLocation.region.endColumn",
	Message:   "runs[].results[].message.text",
}

func FindingsFromJsonMappings(s string, m jsonToFindingsMappings) ([]Finding, error) {
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
	} else {
		if strings.HasPrefix(finding.FilePath, GitRepoBasePath) {
			finding.FilePath = strings.TrimPrefix(finding.FilePath, GitRepoBasePath)
		}
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

func getValueFromMapping(data map[string]interface{}, mapping string) (string, error) {
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

func getIntValueFromMapping(data map[string]interface{}, mapping string) (int, error) {
	value, err := getValueFromMapping(data, mapping)
	if err != nil {
		return 1, err
	}
	return strconv.Atoi(value)
}
