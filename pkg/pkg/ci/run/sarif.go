package main

import (
	"encoding/json"
	"log/slog"

	"github.com/kemadev/workflows-and-actions/pkg/pkg/ci/filesfinder"
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

type dummyJson struct {
	Schema string `json:"$schema"`
	Foo    string `json:"foo.bar.baz"`
}

func HandleSarifString(s string, format string) (int, error) {
	var sarif SarifFile
	if err := json.Unmarshal([]byte(s), &sarif); err != nil {
		slog.Error("Error unmarshalling SARIF string", slog.String("error", err.Error()))
		return 1, err
	}

	annotations, err := annotationsFromSarif(sarif)
	if err != nil {
		slog.Error("Error converting SARIF to annotations", slog.String("error", err.Error()))
		return 1, err
	}

	rc, err := PrintFindings(annotations, format)
	if err != nil {
		return 1, err
	}
	return rc, nil
}

func annotationsFromSarif(sarif SarifFile) ([]Finding, error) {
	var annotations []Finding
	for _, run := range sarif.Runs {
		for _, result := range run.Results {
			for _, location := range result.Locations {
				// repo is mounted at /src
				relpath := location.PhysicalLocation.ArtifactLocation.URI
				l := len(filesfinder.RootPath) + 1
				if len(relpath) > l && relpath[:l] == filesfinder.RootPath+"/" {
					relpath = relpath[l:]
				}
				if result.Level == "" {
					result.Level = "error"
				}
				annotation := Finding{
					ToolName:  run.Tool.Driver.Name,
					RuleID:    result.RuleID,
					Level:     result.Level,
					FilePath:  relpath,
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
	return annotations, nil
}
