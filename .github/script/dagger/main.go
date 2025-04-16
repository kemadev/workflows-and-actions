package main

import (
	"context"
	"encoding/json"
	"fmt"

	"dagger/dagger/internal/dagger"
)

type Dagger struct{}

type Results struct {
	Logs     string
	ExitCode int
	Error    error
}

const (
	findingsJsonPathVarName  = "FINDINGS_PATH"
	findingsJsonPathVarValue = "/tmp/findings.json"
)

func processResults(ctx context.Context, f *dagger.File) (string, error) {
	// TODO pass arg or w/e to check for human or gh output format
	dummyHumanSwitch := true
	s, err := f.Contents(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get file contents: %w", err)
	}
	fm := make(map[string]interface{})
	err = json.Unmarshal([]byte(s), &fm)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	l := ""
	if dummyHumanSwitch {
		l, err = processFindingsGithub(fm)
		if err != nil {
			return "", fmt.Errorf("failed to output findings: %w", err)
		}
	} else {
		// TODO obv human output
		l, err = processFindingsGithub(fm)
		if err != nil {
			return "", fmt.Errorf("failed to output findings: %w", err)
		}
	}

	return l, nil
}

func processFindingsGithub(findings map[string]interface{}) (string, error) {
	toolName, err := extractToolName(findings)
	if err != nil {
		return "", fmt.Errorf("failed to extract tool name: %w", err)
	}

	results, err := extractResults(findings)
	if err != nil {
		return "", fmt.Errorf("failed to extract results: %w", err)
	}

	l := ""
	for _, result := range results {
		a, err := getGithubAnnotation(toolName, result)
		if err != nil {
			return "", fmt.Errorf("failed to print GitHub annotation: %w", err)
		}
		l += a + "\n"
	}
	return "Processed findings successfully", nil
}

func extractToolName(findings map[string]interface{}) (string, error) {
	tool, ok := findings["tool"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("missing or invalid 'tool' field")
	}
	driver, ok := tool["driver"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("missing or invalid 'driver' field")
	}
	name, ok := driver["name"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid 'name' field")
	}
	return name, nil
}

func extractResults(findings map[string]interface{}) ([]map[string]interface{}, error) {
	runs, ok := findings["runs"].([]interface{})
	if !ok || len(runs) == 0 {
		return nil, fmt.Errorf("missing or invalid 'runs' field")
	}
	firstRun, ok := runs[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid 'runs[0]' field")
	}
	results, ok := firstRun["results"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'results' field")
	}

	var parsedResults []map[string]interface{}
	for _, result := range results {
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid result entry")
		}
		parsedResults = append(parsedResults, resultMap)
	}
	return parsedResults, nil
}

func getGithubAnnotation(toolName string, result map[string]interface{}) (string, error) {
	loc, err := extractLocation(result)
	if err != nil {
		return "", fmt.Errorf("failed to extract location: %w", err)
	}

	message, ok := result["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("missing or invalid 'message' field")
	}
	text, ok := message["text"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid 'text' field")
	}

	a := fmt.Sprintf("::error title=%s,file=%s,line=%d,endLine=%d,col=%d,endColumn=%d::%s\n",
		toolName,
		loc.ArtifactURI,
		loc.StartLine,
		loc.EndLine,
		loc.StartColumn,
		loc.EndColumn,
		text,
	)
	return a, nil
}

type Location struct {
	ArtifactURI string
	StartLine   int
	EndLine     int
	StartColumn int
	EndColumn   int
}

func extractLocation(result map[string]interface{}) (Location, error) {
	locations, ok := result["locations"].([]interface{})
	if !ok || len(locations) == 0 {
		return Location{}, fmt.Errorf("missing or invalid 'locations' field")
	}
	physicalLocation, ok := locations[0].(map[string]interface{})["physicalLocation"].(map[string]interface{})
	if !ok {
		return Location{}, fmt.Errorf("missing or invalid 'physicalLocation' field")
	}
	artifactLocation, ok := physicalLocation["artifactLocation"].(map[string]interface{})
	if !ok {
		return Location{}, fmt.Errorf("missing or invalid 'artifactLocation' field")
	}
	region, ok := physicalLocation["region"].(map[string]interface{})
	if !ok {
		return Location{}, fmt.Errorf("missing or invalid 'region' field")
	}

	return Location{
		ArtifactURI: artifactLocation["uri"].(string),
		StartLine:   int(region["startLine"].(int)),
		EndLine:     int(region["endLine"].(int)),
		StartColumn: int(region["startColumn"].(int)),
		EndColumn:   int(region["endColumn"].(int)),
	}, nil
}
