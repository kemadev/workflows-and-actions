package sarifparser

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/kemadev/workflows-and-actions/pkg/pkg/logger/runner"
)

type SarifFile struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []struct {
		Tool struct {
			Driver struct {
				Name             string `json:"name"`
				DownloadURI      string `json:"downloadUri"`
				FullName         string `json:"fullName"`
				ShortDescription struct {
					Text string `json:"text"`
				} `json:"shortDescription"`
				Version string `json:"version"`
			} `json:"driver"`
		} `json:"tool"`
		Results []struct {
			Message struct {
				Text string `json:"text"`
			} `json:"message"`
			Level     string `json:"level"`
			RuleID    string `json:"ruleId"`
			Locations []struct {
				PhysicalLocation struct {
					ArtifactLocation struct {
						URI string `json:"uri"`
					} `json:"artifactLocation"`
					Region struct {
						StartLine      int    `json:"startLine"`
						EndLine        int    `json:"endLine"`
						StartColumn    int    `json:"startColumn"`
						EndColumn      int    `json:"endColumn"`
						SourceLanguage string `json:"sourceLanguage"`
					} `json:"region"`
				} `json:"physicalLocation"`
			} `json:"locations"`
		} `json:"results"`
	} `json:"runs"`
}

type Finding struct {
	ToolName  string `json:"tool_name"`
	RuleID    string `json:"rule_id"`
	Level     string `json:"level"`
	FilePath  string `json:"file_path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	StartCol  int    `json:"start_col"`
	EndCol    int    `json:"end_col"`
	Message   string `json:"message"`
}

func ParseSarifString(s string, format string) int {
	var sarif SarifFile
	if err := json.Unmarshal([]byte(s), &sarif); err != nil {
		slog.Error("Error unmarshalling SARIF string", slog.String("error", err.Error()))
		return 1
	}

	rc := printFindings(sarif, format)
	return rc
}

func ParseSarifFile(path string, format string) int {
	file, err := os.Open(path)
	if err != nil {
		slog.Error("Error opening SARIF file", slog.String("error", err.Error()))
		return 1
	}
	defer file.Close()

	var sarif SarifFile
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&sarif); err != nil {
		slog.Error("Error decoding SARIF file", slog.String("error", err.Error()))
		return 1
	}

	rc := printFindings(sarif, format)
	return rc
}

func printFindings(sarif SarifFile, format string) int {
	var annotations []Finding
	for _, run := range sarif.Runs {
		for _, result := range run.Results {
			for _, location := range result.Locations {
				annotation := Finding{
					ToolName:  run.Tool.Driver.Name,
					RuleID:    result.RuleID,
					Level:     result.Level,
					FilePath:  location.PhysicalLocation.ArtifactLocation.URI,
					StartLine: location.PhysicalLocation.Region.StartLine,
					EndLine:   location.PhysicalLocation.Region.EndLine,
					StartCol:  location.PhysicalLocation.Region.StartColumn,
					EndCol:    location.PhysicalLocation.Region.EndColumn,
					Message:   result.Message.Text,
				}
				annotations = append(annotations, annotation)
			}
		}
	}

	switch format {
	case "human":
		for _, annotation := range annotations {
			fmt.Printf("Tool: %s\n", annotation.ToolName)
			fmt.Printf("Rule ID: %s\n", annotation.RuleID)
			fmt.Printf("Level: %s\n", annotation.Level)
			fmt.Printf("File: %s\n", annotation.FilePath+":"+fmt.Sprintf("%d", annotation.StartLine))
			fmt.Printf("Message: %s\n", annotation.Message)
			fmt.Println()
		}
	case "json":
		output, err := json.MarshalIndent(annotations, "", "  ")
		if err != nil {
			slog.Error("Error marshalling JSON", slog.String("error", err.Error()))
			return 1
		}
		fmt.Println(string(output))
	case "github":
		for _, annotation := range annotations {
			fmt.Printf("::%s title=%s file=%s,line=%d,endLine=%d,col=%d,endColumn=%d::%s\n", annotation.Level, annotation.ToolName, annotation.FilePath, annotation.StartLine, annotation.EndLine, annotation.StartCol, annotation.EndCol, annotation.Message)
		}
	default:
		slog.Error("Unknown format", slog.String("format", format))
		return 1
	}
	rc := 0
	if len(annotations) > 0 {
		rc = 1
	}
	return rc
}
