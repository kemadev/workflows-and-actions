package runner

import (
	"encoding/json"
	"fmt"
)

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

func PrintFindings(annotations []Finding, format string) (int, error) {
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
			return 1, err
		}
		fmt.Println(string(output))
	case "github":
		for _, annotation := range annotations {
			// See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions
			fmt.Printf("::%s title=%s,file=%s,line=%d,endLine=%d,col=%d,endColumn=%d::%s\n", annotation.Level, annotation.ToolName, annotation.FilePath, annotation.StartLine, annotation.EndLine, annotation.StartCol, annotation.EndCol, annotation.Message)
		}
	default:
		return 1, fmt.Errorf("unknown format: %s", format)
	}
	rc := 0
	if len(annotations) > 0 {
		rc = 1
	}
	return rc, nil
}
