package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v68/github"
	"golang.org/x/tools/cover"
)

var (
	// GitHub Actions provided GitHub authentication token
	ghToken = os.Getenv("GH_TOKEN")
	// Number of the pull request that triggered the workflow
	prNumber = os.Getenv("PR_NUMBER")
	// Number of the pull request that triggered the workflow as an integer
	prNumberInt int
	// Path to the coverage file for which the report is to be generated
	coverageFile = os.Getenv("COVERAGE_FILE")
	// GitHub repository
	repo = os.Getenv("GITHUB_REPOSITORY")
	// GitHub repository owner
	repoOwner string
	// GitHub repository name
	repoName string
	// GitHub host
	ghHost = os.Getenv("GITHUB_SERVER_URL")
	// Head branch
	branch = os.Getenv("GITHUB_HEAD_REF")
)

const (
	coverageCommentDelimiterStart = "<!-- START gha:coverage-report -->"
	coverageCommentDelimiterEnd   = "<!-- END gha:coverage-report -->"
)

func initLogger() {
	var logLevel slog.Level
	if os.Getenv("RUNNER_DEBUG") == "1" {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))
}

func checkVariables() error {
	if ghToken == "" {
		return fmt.Errorf("Environment variable GH_TOKEN is not set")
	}
	if prNumber == "" {
		return fmt.Errorf("Environment variable PR_NUMBER is not set")
	} else {
		prNum, err := strconv.Atoi(prNumber)
		if err != nil {
			return fmt.Errorf("PR_NUMBER is not a valid number")
		}
		prNumberInt = prNum
	}
	if coverageFile == "" {
		return fmt.Errorf("Environment variable COVERAGE_FILE is not set")
	}
	if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
		return fmt.Errorf("Coverage file does not exist")
	}
	if repo == "" {
		return fmt.Errorf("Environment variable GITHUB_REPOSITORY is not set")
	}
	repoParts := strings.Split(repo, "/")
	if len(repoParts) != 2 {
		return fmt.Errorf("Invalid GITHUB_REPOSITORY format")
	}
	repoOwner = repoParts[0]
	if repoOwner == "" {
		return fmt.Errorf("Can't infer repository owner from GITHUB_REPOSITORY environment variable")
	}
	repoName = repoParts[1]
	if repoName == "" {
		return fmt.Errorf("Can't infer repository name from GITHUB_REPOSITORY environment variable")
	}
	if ghHost == "" {
		return fmt.Errorf("Environment variable GITHUB_SERVER_URL is not set")
	}
	return nil
}

func generateCoverageReport() ([]*cover.Profile, error) {
	profiles, err := cover.ParseProfiles(coverageFile)
	if err != nil {
		return nil, err
	}
	slog.Debug(fmt.Sprintf("Parsed %d coverage profiles", len(profiles)))
	return profiles, nil
}

// From https://github.com/golang/tools/blob/aa82965741a9fecd12b026fbb3d3c6ed3231b8f8/cmd/cover/html.go#L93
func percentCovered(p *cover.Profile) float64 {
	var total, covered int64
	for _, b := range p.Blocks {
		total += int64(b.NumStmt)
		if b.Count > 0 {
			covered += int64(b.NumStmt)
		}
	}
	if total == 0 {
		return 0
	}
	slog.Debug(fmt.Sprintf("Profile %s: %d/%d", p.FileName, covered, total))
	return float64(covered) / float64(total) * 100
}

func parseCoverageReport(profiles []*cover.Profile) (string, error) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("## Test Coverage Report :test_tube:\n"))
	buffer.WriteString(fmt.Sprintf("\n"))
	buffer.WriteString(fmt.Sprintf("| File Path | Coverage |\n"))
	buffer.WriteString(fmt.Sprintf("| --------- | -------- |\n"))
	for _, profile := range profiles {
		fileCoveragePercentage := percentCovered(profile)
		filenameWithoutRepoPathArray := strings.Split(profile.FileName, repoName+"/")
		if len(filenameWithoutRepoPathArray) < 2 {
			return "", fmt.Errorf("Failed to extract file path from coverage profile: %s", profile.FileName)
		}
		filenameWithoutRepoPath := filenameWithoutRepoPathArray[1]
		// https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#setting-a-warning-message
		fmt.Println("::warning file=" + filenameWithoutRepoPath + ",line=1,title=Poor test coverage::" + fmt.Sprintf("Coverage on file %s: %.2f%%", filenameWithoutRepoPath, fileCoveragePercentage))
		var emoji string
		switch {
		case fileCoveragePercentage < 70:
			emoji = ":boom:"
		case fileCoveragePercentage < 85:
			emoji = ":mending_heart:"
		default:
			emoji = ":heart:"
		}
		buffer.WriteString(fmt.Sprintf("| [%s](%s) | %s %.2f%% |\n", filenameWithoutRepoPath, ghHost+"/"+repo+"/tree/"+branch+"/"+filenameWithoutRepoPath, emoji, fileCoveragePercentage))
	}
	buffer.WriteString(fmt.Sprintf("\n"))
	slog.Debug("Generated coverage report", slog.String("report", buffer.String()))
	return buffer.String(), nil
}

func updatePrWithCoverageReport(report string) error {
	client := github.NewClient(nil).WithAuthToken(ghToken)
	pr, resp, err := client.PullRequests.Get(context.TODO(), repoOwner, repoName, prNumberInt)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to get pull request")
	}
	slog.Info(fmt.Sprintf("Found pull request: %d", prNumberInt))
	var buffer bytes.Buffer

	buffer.WriteString(coverageCommentDelimiterStart + "\n")
	buffer.WriteString("<!-- This field is auto-generated by a workflow, do not edit manually -->" + "\n")
	buffer.WriteString(report)
	buffer.WriteString(coverageCommentDelimiterEnd)
	var body string
	if pr.Body != nil {
		body = *pr.Body
	} else {
		body = ""
	}
	// Get the body, excluding the coverage report delimited by the start and end delimiters
	startIndex := strings.Index(body, coverageCommentDelimiterStart)
	endIndex := strings.Index(body, coverageCommentDelimiterEnd)
	if startIndex != -1 && endIndex != -1 {
		body = body[:startIndex] + body[endIndex+len(coverageCommentDelimiterEnd):]
	}
	// If last character before start is not a newline, add it
	if startIndex != -1 && startIndex != 0 && body[startIndex-1] != '\n' {
		body = body[:startIndex] + "\n" + body[startIndex:]
	}
	body += buffer.String()
	// Update the pull request with the new body
	editedPr := &github.PullRequest{
		Body: &body,
	}
	slog.Debug("Updating pull request", slog.String("body", body))
	_, resp, err = client.PullRequests.Edit(context.TODO(), repoOwner, repoName, prNumberInt, editedPr)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to update pull request")
	}
	slog.Info("Updated pull request with coverage report!")
	return nil
}

func main() {
	startTime := time.Now()
	defer func() {
		slog.Debug("Execution time", slog.String("duration", time.Since(startTime).String()))
	}()
	initLogger()
	err := checkVariables()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	coverageReport, err := generateCoverageReport()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	report, err := parseCoverageReport(coverageReport)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	err = updatePrWithCoverageReport(report)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
