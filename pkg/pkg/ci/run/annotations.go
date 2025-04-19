package main

import (
	"encoding/json"
	"fmt"
	"slices"
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

func PrintFindings(findings []Finding, format string) error {
	err := validateFindings(findings)
	if err != nil {
		return err
	}
	switch format {
	case "human":
		for _, annotation := range findings {
			fmt.Printf("Tool: %s\n", annotation.ToolName)
			fmt.Printf("Rule ID: %s\n", annotation.RuleID)
			fmt.Printf("Level: %s\n", annotation.Level)
			fmt.Printf("File: %s\n", annotation.FilePath+":"+fmt.Sprintf("%d", annotation.StartLine))
			fmt.Printf("Message: %s\n", annotation.Message)
			fmt.Println()
		}
	case "json":
		output, err := json.MarshalIndent(findings, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(output))
	case "github":
		for _, annotation := range findings {
			// See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions
			fmt.Printf("::%s title=%s,file=%s,line=%d,endLine=%d,col=%d,endColumn=%d::%s\n", annotation.Level, annotation.ToolName, annotation.FilePath, annotation.StartLine, annotation.EndLine, annotation.StartCol, annotation.EndCol, annotation.Message)
		}
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

// Based on GitHub workflow commands, see https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#setting-a-debug-message
var validFindingLevels = []string{"debug", "notice", "warning", "error"}

func validateFindings(f []Finding) error {
	for _, annotation := range f {
		if annotation.ToolName == "" {
			return fmt.Errorf("tool name is required for finding: %v", annotation)
		}
		if annotation.RuleID == "" {
			return fmt.Errorf("rule ID is required for finding: %v", annotation)
		}
		if !slices.Contains(validFindingLevels, annotation.Level) {
			return fmt.Errorf("invalid level %s for finding: %v", annotation.Level, annotation)
		}
		if annotation.FilePath == "" {
			return fmt.Errorf("file path is required for finding: %v", annotation)
		}
		if annotation.StartLine == 0 {
			return fmt.Errorf("start line is required for finding: %v", annotation)
		}
		if annotation.EndLine == 0 {
			return fmt.Errorf("end line is required for finding: %v", annotation)
		}
		if annotation.Message == "" {
			return fmt.Errorf("message is required for finding: %v", annotation)
		}
	}
	return nil
}
