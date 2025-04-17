package sarifparser

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

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

var outputFormat string

func init() {
	flag.StringVar(&outputFormat, "output-format", "human", "Output format (human, json, github)")
}

func ParseSarifFile(path string, gha bool) int {
	s := time.Now()
	flag.Parse()
	sarifFilePath := path
	if sarifFilePath == "" {
		sarifFilePath = flag.Arg(0)
		if sarifFilePath == "" {
			flag.Usage()
			return 1
		}
	}
	if gha {
		outputFormat = "github"
	} else if outputFormat != "human" && outputFormat != "json" && outputFormat != "github" {
		flag.Usage()
		return 1
	}
	slog.Debug("SARIF file path: ", slog.String("path", sarifFilePath))
	slog.Debug("Output format: ", slog.String("format", outputFormat))

	file, err := os.Open(sarifFilePath)
	if err != nil {
		fmt.Printf("Error opening SARIF file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var sarif SarifFile
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&sarif); err != nil {
		fmt.Printf("Error decoding SARIF file: %v\n", err)
		os.Exit(1)
	}

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

	if outputFormat == "human" {
		for _, annotation := range annotations {
			fmt.Printf("Tool: %s\n", annotation.ToolName)
			fmt.Printf("Rule ID: %s\n", annotation.RuleID)
			fmt.Printf("Level: %s\n", annotation.Level)
			fmt.Printf("File: %s\n", annotation.FilePath+":"+fmt.Sprintf("%d", annotation.StartLine))
			fmt.Printf("Message: %s\n", annotation.Message)
			fmt.Println()
		}
	}
	if outputFormat == "json" {
		output, err := json.MarshalIndent(annotations, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling annotations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	}
	if outputFormat == "github" {
		for _, annotation := range annotations {
			fmt.Printf("::%s title=%s file=%s,line=%d,endLine=%d,col=%d,endColumn=%d::%s\n", annotation.Level, annotation.ToolName, annotation.FilePath, annotation.StartLine, annotation.EndLine, annotation.StartCol, annotation.EndCol, annotation.Message)
		}
	}
	rc := 0
	if len(annotations) > 0 {
		rc = 1
	}
	slog.Debug("Execution time: ", slog.Duration("duration", time.Since(s)))
	return rc
}
