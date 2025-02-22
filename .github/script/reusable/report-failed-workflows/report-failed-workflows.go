package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/v69/github"
)

const (
	issueTitle                    = ":rotating_light: Failed workflows"
	issueBodyIdentifier           = "<!-- gha:report-failed-workflows -->"
	workflowsInfosIdentifierStart = "<!-- gha:report-failed-workflows-workflows-infos START\n"
	workflowsInfosIdentifierEnd   = "\ngha:report-failed-workflows-workflows-infos END -->"
)

type staleBranch struct {
	name                string
	author              string
	date                string
	timeSinceLastCommit string
}

var (
	gh                       *github.Client
	githubRepository         = os.Getenv("GITHUB_REPOSITORY")
	repoOwner                string
	repoName                 string
	issueBody                string
	issueNumber              int
	repo                     *git.Repository
	atLeastOneWorkflowFailed bool = false

	ghToken                string = os.Getenv("GH_TOKEN")
	headBranch             string = os.Getenv("HEAD_BRANCH")
	workflowName           string = os.Getenv("WORKFLOW_NAME")
	workflowRunTitle       string = os.Getenv("WORKFLOW_RUN_TITLE")
	conclusion             string = os.Getenv("CONCLUSION")
	htmlUrl                string = os.Getenv("HTML_URL")
	createdAt              string = os.Getenv("CREATED_AT")
	updatedAt              string = os.Getenv("UPDATED_AT")
	actorType              string = os.Getenv("ACTOR_TYPE")
	actorHtmlUrl           string = os.Getenv("ACTOR_HTML_URL")
	triggeringActorType    string = os.Getenv("TRIGGERING_ACTOR_TYPE")
	triggeringActorHtmlUrl string = os.Getenv("TRIGGERING_ACTOR_HTML_URL")
)

type workflowInfos struct {
	HeadBranch             string `json:"headBranch"`
	WorkflowName           string `json:"workflowName"`
	WorkflowRunTitle       string `json:"workflowRunTitle"`
	Conclusion             string `json:"conclusion"`
	HtmlUrl                string `json:"htmlUrl"`
	CreatedAt              string `json:"createdAt"`
	UpdatedAt              string `json:"updatedAt"`
	ActorType              string `json:"actorType"`
	ActorHtmlUrl           string `json:"actorHtmlUrl"`
	TriggeringActorType    string `json:"triggeringActorType"`
	TriggeringActorHtmlUrl string `json:"triggeringActorHtmlUrl"`
}

type allWorkflowsInfos struct {
	WorkflowsInfos map[string]workflowInfos `json:"workflowsInfos"`
}

func initLogger() {
	var logLevel slog.Level
	if os.Getenv("RUNNER_DEBUG") == "true" {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))
}

func checkAndSetVariables() {
	slog.Debug("START checkAndSetVariables")
	if githubRepository == "" {
		slog.Error("GITHUB_REPOSITORY is not set")
		os.Exit(1)
	}
	repoOwner = strings.Split(githubRepository, "/")[0]
	repoName = strings.Split(githubRepository, "/")[1]
	if headBranch == "" {
		slog.Error("HEAD_BRANCH is not set")
		os.Exit(1)
	}
	if workflowName == "" {
		slog.Error("WORKFLOW_NAME is not set")
		os.Exit(1)
	}
	if workflowRunTitle == "" {
		slog.Error("WORKFLOW_RUN_TITLE is not set")
		os.Exit(1)
	}
	if conclusion == "" {
		slog.Error("CONCLUSION is not set")
		os.Exit(1)
	}
	if htmlUrl == "" {
		slog.Error("HTML_URL is not set")
		os.Exit(1)
	}
	if createdAt == "" {
		slog.Error("CREATED_AT is not set")
		os.Exit(1)
	}
	if updatedAt == "" {
		slog.Error("UPDATED_AT is not set")
		os.Exit(1)
	}
	if actorType == "" {
		slog.Error("ACTOR_TYPE is not set")
		os.Exit(1)
	}
	if actorHtmlUrl == "" {
		slog.Error("ACTOR_HTML_URL is not set")
		os.Exit(1)
	}
	if triggeringActorType == "" {
		slog.Error("TRIGGERING_ACTOR_TYPE is not set")
		os.Exit(1)
	}
	if triggeringActorHtmlUrl == "" {
		slog.Error("TRIGGERING_ACTOR_HTML_URL is not set")
		os.Exit(1)
	}
	if ghToken == "" {
		slog.Error("GH_TOKEN is not set")
		os.Exit(1)
	}
	slog.Debug("END checkAndSetVariables", slog.Group("variables", slog.Any("githubRepository", githubRepository), slog.Any("repoOwner", repoOwner), slog.Any("repoName", repoName), slog.Any("headBranch", headBranch), slog.Any("workflowName", workflowName), slog.Any("workflowRunTitle", workflowRunTitle), slog.Any("conclusion", conclusion), slog.Any("htmlUrl", htmlUrl), slog.Any("createdAt", createdAt), slog.Any("updatedAt", updatedAt), slog.Any("actorType", actorType), slog.Any("actorHtmlUrl", actorHtmlUrl), slog.Any("triggeringActorType", triggeringActorType), slog.Any("triggeringActorHtmlUrl", triggeringActorHtmlUrl)))
}

func initGithubClient() {
	slog.Debug("START initGithubClient")
	gh = github.NewClient(nil).WithAuthToken(ghToken)
	slog.Debug("END initGithubClient")
}

func initGitClient() {
	slog.Debug("START initGitClient")
	r, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		slog.Error("Failed to open git repository", slog.String("error", err.Error()))
		os.Exit(1)
	}
	repo = r
	slog.Debug("END initGitClient")
}

func parseWorkflowsInfos() allWorkflowsInfos {
	slog.Debug("START parseWorkflowsInfos")
	workflowsInfos := allWorkflowsInfos{}
	workflowInfosStartIndex := strings.Index(issueBody, workflowsInfosIdentifierStart)
	workflowInfosEndIndex := strings.Index(issueBody, workflowsInfosIdentifierEnd)
	if workflowInfosStartIndex != -1 && workflowInfosEndIndex != -1 {
		workflowsInfosString := issueBody[workflowInfosStartIndex+len(workflowsInfosIdentifierStart) : workflowInfosEndIndex]
		err := json.Unmarshal([]byte(workflowsInfosString), &workflowsInfos)
		if err != nil {
			slog.Error("Failed to parse workflows infos", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}
	slog.Debug("END parseWorkflowsInfos", slog.Any("workflowsInfos", workflowsInfos))
	return workflowsInfos
}

func computeIssueBody() {
	slog.Debug("START computeIssueBody")
	currentIssue, resp, err := gh.Issues.Get(context.TODO(), repoOwner, repoName, issueNumber)
	if err != nil {
		slog.Error("Failed to get issue body", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if resp.StatusCode != 200 {
		slog.Error("Failed to get issue body", slog.String("status", resp.Status))
		os.Exit(1)
	}
	issueBody = currentIssue.GetBody()
	newWorkflow := workflowInfos{
		HeadBranch:             headBranch,
		WorkflowName:           workflowName,
		WorkflowRunTitle:       workflowRunTitle,
		Conclusion:             conclusion,
		HtmlUrl:                htmlUrl,
		CreatedAt:              createdAt,
		UpdatedAt:              updatedAt,
		ActorType:              actorType,
		ActorHtmlUrl:           actorHtmlUrl,
		TriggeringActorType:    triggeringActorType,
		TriggeringActorHtmlUrl: triggeringActorHtmlUrl,
	}
	allWorkflows := parseWorkflowsInfos()
	if allWorkflows.WorkflowsInfos == nil {
		allWorkflows.WorkflowsInfos = make(map[string]workflowInfos)
	}
	allWorkflows.WorkflowsInfos[workflowName] = newWorkflow

	// search if at least one workflow failed
	for _, workflow := range allWorkflows.WorkflowsInfos {
		if workflow.Conclusion == "failure" {
			atLeastOneWorkflowFailed = true
			break
		}
	}
	workflowsInfosBytes, err := json.Marshal(allWorkflows)
	if err != nil {
		slog.Error("Failed to marshal workflows infos", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var buffer bytes.Buffer
	buffer.WriteString(issueBodyIdentifier + "\n")
	buffer.WriteString("<!-- This issue is auto-generated by a workflow, do not edit manually -->\n")
	buffer.WriteString(workflowsInfosIdentifierStart + string(workflowsInfosBytes) + workflowsInfosIdentifierEnd + "\n\n")
	if !atLeastOneWorkflowFailed {
		buffer.WriteString("## :confetti_ball: All workflows successful\n\n")
	} else {
		buffer.WriteString("## :rotating_light: Failed workflows\n\n")
		buffer.WriteString("| Workflow | Conclusion | Run | Time | Actor | Triggering Actor |\n")
		buffer.WriteString("| -------- | ---------- | --- | ---- | ----- | ---------------- |\n")
		for _, workflow := range allWorkflows.WorkflowsInfos {
			if workflow.Conclusion != "success" {
				buffer.WriteString(fmt.Sprintf("| %s | %s | [%s](%s) | %s | %s (%s) | %s (%s) |\n", workflow.WorkflowName, workflow.Conclusion, workflow.WorkflowRunTitle, workflow.HtmlUrl, workflow.UpdatedAt, workflow.ActorHtmlUrl, workflow.ActorType, workflow.TriggeringActorHtmlUrl, workflow.TriggeringActorType))
			}
		}
	}
	issueBody = buffer.String()
	slog.Info("Issue body computed", slog.String("issueBody", issueBody))
	slog.Debug("END computeIssueBody", slog.String("issueBody", issueBody))
}

func createIssueIfNeeded() {
	slog.Debug("START createIssueIfNeeded")
	issueNumber = getIssueNumber()
	if issueNumber == -1 {
		slog.Info("Issue not found, creating a new one")
		issueNumber = createIssue()
		slog.Info("Issue created", slog.Int("issueNumber", issueNumber))
	}
	slog.Debug("END createIssueIfNeeded", slog.Int("issueNumber", issueNumber))
}

func createOrUpdateIssue() {
	slog.Debug("START createOrUpdateIssue")
	issueNumber = getIssueNumber()
	if issueNumber == -1 {
		slog.Info("Issue not found, creating a new one")
		issueNumber = createIssue()
		slog.Info("Issue created", slog.Int("issueNumber", issueNumber))
	} else {
		updateIssue(issueNumber)
	}
	pinIssue(issueNumber)
	slog.Debug("END createOrUpdateIssue", slog.Int("issueNumber", issueNumber))
}

func main() {
	startTime := time.Now()
	defer func() {
		slog.Info("Execution time", slog.String("duration", time.Since(startTime).String()))
	}()
	initLogger()
	checkAndSetVariables()
	initGithubClient()
	initGitClient()
	createIssueIfNeeded()
	computeIssueBody()
	createOrUpdateIssue()
	slog.Info("Workflow completed successfully")
}

func getIssueNumber() int {
	slog.Debug("START getIssueNumber")
	issues, resp, err := gh.Search.Issues(context.TODO(), "is:issue in:title "+issueTitle+" repo:"+githubRepository, nil)
	if err != nil {
		slog.Error("Failed to get issue number", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if resp.StatusCode != 200 {
		slog.Error("Failed to get issue number", slog.String("status", resp.Status))
		os.Exit(1)
	}
	issueNumber := -1
	for _, issue := range issues.Issues {
		if strings.Contains(issue.GetBody(), issueBodyIdentifier) {
			issueNumber = issue.GetNumber()
			slog.Info("Issue found", slog.Int("issueNumber", issueNumber))
			break
		}
	}
	slog.Debug("END getIssueNumber", slog.Int("issueNumber", issueNumber))
	return issueNumber
}

func createIssue() int {
	slog.Debug("START createIssue")
	i := github.IssueRequest{
		Title: github.Ptr(issueTitle),
		Body:  &issueBody,
	}
	issue, resp, err := gh.Issues.Create(context.TODO(), repoOwner, repoName, &i)
	if err != nil {
		slog.Error("Failed to create issue", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if resp.StatusCode != 201 {
		slog.Error("Failed to create issue", slog.String("status", resp.Status))
		os.Exit(1)
	}
	issueNumber := issue.GetNumber()
	slog.Debug("END createIssue", slog.Int("issueNumber", issueNumber))
	return issueNumber
}

func updateIssue(issueNumber int) {
	slog.Debug("START updateIssue")
	var issueDesiredStatus string
	if atLeastOneWorkflowFailed {
		issueDesiredStatus = "open"
	} else {
		issueDesiredStatus = "closed"
	}
	i := github.IssueRequest{
		Body:  &issueBody,
		State: github.Ptr(issueDesiredStatus),
	}
	_, resp, err := gh.Issues.Edit(context.TODO(), repoOwner, repoName, issueNumber, &i)
	if err != nil {
		slog.Error("Failed to update issue", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if resp.StatusCode != 200 {
		slog.Error("Failed to update issue", slog.String("status", resp.Status))
		os.Exit(1)
	}
	slog.Info("Issue updated", slog.Int("issueNumber", issueNumber))
	slog.Debug("END updateIssue")
}

func pinIssue(issueNumber int) {
	slog.Debug("START pinIssue")
	exec.Command("gh", "issue", "pin", strconv.Itoa(issueNumber)).Run()
	slog.Info("Issue pinned", slog.Int("issueNumber", issueNumber))
	slog.Debug("END pinIssue")
}
