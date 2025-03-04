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
	if os.Getenv("RUNNER_DEBUG") == "1" {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))
}

func checkAndSetVariables() error {
	slog.Debug("START checkAndSetVariables")
	if githubRepository == "" {
		return fmt.Errorf("GITHUB_REPOSITORY is not set")
	}
	repoOwner = strings.Split(githubRepository, "/")[0]
	repoName = strings.Split(githubRepository, "/")[1]
	if headBranch == "" {
		return fmt.Errorf("HEAD_BRANCH is not set")
	}
	if workflowName == "" {
		return fmt.Errorf("WORKFLOW_NAME is not set")
	}
	if workflowRunTitle == "" {
		return fmt.Errorf("WORKFLOW_RUN_TITLE is not set")
	}
	if conclusion == "" {
		return fmt.Errorf("CONCLUSION is not set")
	}
	if htmlUrl == "" {
		return fmt.Errorf("HTML_URL is not set")
	}
	if createdAt == "" {
		return fmt.Errorf("CREATED_AT is not set")
	}
	if updatedAt == "" {
		return fmt.Errorf("UPDATED_AT is not set")
	}
	if actorType == "" {
		return fmt.Errorf("ACTOR_TYPE is not set")
	}
	if actorHtmlUrl == "" {
		return fmt.Errorf("ACTOR_HTML_URL is not set")
	}
	if triggeringActorType == "" {
		return fmt.Errorf("TRIGGERING_ACTOR_TYPE is not set")
	}
	if triggeringActorHtmlUrl == "" {
		return fmt.Errorf("TRIGGERING_ACTOR_HTML_URL is not set")
	}
	if ghToken == "" {
		return fmt.Errorf("GH_TOKEN is not set")
	}
	slog.Debug("END checkAndSetVariables", slog.Group("variables", slog.Any("githubRepository", githubRepository), slog.Any("repoOwner", repoOwner), slog.Any("repoName", repoName), slog.Any("headBranch", headBranch), slog.Any("workflowName", workflowName), slog.Any("workflowRunTitle", workflowRunTitle), slog.Any("conclusion", conclusion), slog.Any("htmlUrl", htmlUrl), slog.Any("createdAt", createdAt), slog.Any("updatedAt", updatedAt), slog.Any("actorType", actorType), slog.Any("actorHtmlUrl", actorHtmlUrl), slog.Any("triggeringActorType", triggeringActorType), slog.Any("triggeringActorHtmlUrl", triggeringActorHtmlUrl)))
	return nil
}

func initGithubClient() {
	slog.Debug("START initGithubClient")
	gh = github.NewClient(nil).WithAuthToken(ghToken)
	slog.Debug("END initGithubClient")
}

func initGitClient() error {
	slog.Debug("START initGitClient")
	r, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return fmt.Errorf("Failed to open git repository: %s", err)
	}
	repo = r
	slog.Debug("END initGitClient")
	return nil
}

func parseWorkflowsInfos() (allWorkflowsInfos, error) {
	slog.Debug("START parseWorkflowsInfos")
	workflowsInfos := allWorkflowsInfos{}
	workflowInfosStartIndex := strings.Index(issueBody, workflowsInfosIdentifierStart)
	workflowInfosEndIndex := strings.Index(issueBody, workflowsInfosIdentifierEnd)
	if workflowInfosStartIndex != -1 && workflowInfosEndIndex != -1 {
		workflowsInfosString := issueBody[workflowInfosStartIndex+len(workflowsInfosIdentifierStart) : workflowInfosEndIndex]
		err := json.Unmarshal([]byte(workflowsInfosString), &workflowsInfos)
		if err != nil {
			return allWorkflowsInfos{}, fmt.Errorf("Failed to unmarshal workflows infos: %s", err)
		}
	}
	slog.Debug("END parseWorkflowsInfos", slog.Any("workflowsInfos", workflowsInfos))
	return workflowsInfos, nil
}

func trimOldWorkflows(allWorkflows allWorkflowsInfos) (allWorkflowsInfos, error) {
	slog.Debug("START trimOldWorkflows")
	for _, w := range allWorkflows.WorkflowsInfos {
		workflowTime, err := time.Parse(time.RFC3339, w.CreatedAt)
		if err != nil {
			return allWorkflowsInfos{}, err
		}
		// keep only the last day, minus one hour to prevent from inconsistencies in scheduling times
		if workflowTime.Before(time.Now().Add(-time.Hour * (24 - 24))) {
			delete(allWorkflows.WorkflowsInfos, w.WorkflowName)
		}
	}
	slog.Debug("END trimOldWorkflows", slog.Any("allWorkflows", allWorkflows))
	return allWorkflows, nil
}

func computeIssueBody() error {
	slog.Debug("START computeIssueBody")
	currentIssue, resp, err := gh.Issues.Get(context.TODO(), repoOwner, repoName, issueNumber)
	if err != nil {
		return fmt.Errorf("Failed to get issue body: %s", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to get issue body: %s", resp.Status)
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
	allWorkflows, err := parseWorkflowsInfos()
	if err != nil {
		return fmt.Errorf("Failed to parse workflows infos: %s", err)
	}
	if allWorkflows.WorkflowsInfos == nil {
		allWorkflows.WorkflowsInfos = make(map[string]workflowInfos)
	}
	allWorkflows, err = trimOldWorkflows(allWorkflows)
	if err != nil {
		return fmt.Errorf("Failed to trim old workflows: %s", err)
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
		return fmt.Errorf("Failed to marshal workflows infos: %s", err)
	}

	var buffer bytes.Buffer
	buffer.WriteString(issueBodyIdentifier + "\n")
	buffer.WriteString("<!-- This issue is auto-generated by a workflow, do not edit manually -->\n")
	buffer.WriteString(workflowsInfosIdentifierStart + string(workflowsInfosBytes) + workflowsInfosIdentifierEnd + "\n\n")
	if !atLeastOneWorkflowFailed {
		buffer.WriteString("## :confetti_ball: All workflows successful\n\n")
	} else {
		buffer.WriteString("## :rotating_light: Following workflows failed\n\n")
		buffer.WriteString("> [!NOTE]\n")
		buffer.WriteString("> Renaming a workflow will create a new entry in the list below and previous entries will not be updated.\n\n")
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
	return nil
}

func createIssueIfNeeded() error {
	slog.Debug("START createIssueIfNeeded")
	issueNumber, err := getIssueNumber()
	if err != nil {
		return fmt.Errorf("Failed to get issue number: %s", err)
	}
	if issueNumber == -1 {
		slog.Info("Issue not found, creating a new one")
		issueNumber, err = createIssue()
		if err != nil {
			return fmt.Errorf("Failed to create issue: %s", err)
		}
		slog.Info("Issue created", slog.Int("issueNumber", issueNumber))
	}
	slog.Debug("END createIssueIfNeeded", slog.Int("issueNumber", issueNumber))
	return nil
}

func createOrUpdateIssue() error {
	slog.Debug("START createOrUpdateIssue")
	issueNumber, err := getIssueNumber()
	if err != nil {
		return fmt.Errorf("Failed to get issue number: %s", err)
	}
	if issueNumber == -1 {
		slog.Info("Issue not found, creating a new one")
		issueNumber, err := createIssue()
		if err != nil {
			return fmt.Errorf("Failed to create issue: %s", err)
		}
		slog.Info("Issue created", slog.Int("issueNumber", issueNumber))
	} else {
		updateIssue(issueNumber)
	}
	pinIssue(issueNumber)
	slog.Debug("END createOrUpdateIssue", slog.Int("issueNumber", issueNumber))
	return nil
}

func main() {
	startTime := time.Now()
	defer func() {
		slog.Info("Execution time", slog.String("duration", time.Since(startTime).String()))
	}()
	initLogger()
	err := checkAndSetVariables()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	initGithubClient()
	err = initGitClient()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	err = createIssueIfNeeded()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	err = computeIssueBody()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	err = createOrUpdateIssue()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info("Workflow completed successfully")
}

func getIssueNumber() (int, error) {
	slog.Debug("START getIssueNumber")
	issues, resp, err := gh.Search.Issues(context.TODO(), "is:issue in:title "+issueTitle+" repo:"+githubRepository, nil)
	if err != nil {
		return -1, fmt.Errorf("Failed to get issue number: %s", err)
	}
	if resp.StatusCode != 200 {
		return -1, fmt.Errorf("Failed to get issue number: %s", resp.Status)
	}
	issueNumber = -1
	for _, issue := range issues.Issues {
		if strings.Contains(issue.GetBody(), issueBodyIdentifier) {
			issueNumber = issue.GetNumber()
			slog.Info("Issue found", slog.Int("issueNumber", issueNumber))
			break
		}
	}
	slog.Debug("END getIssueNumber", slog.Int("issueNumber", issueNumber))
	return issueNumber, nil
}

func createIssue() (int, error) {
	slog.Debug("START createIssue")
	i := github.IssueRequest{
		Title: github.Ptr(issueTitle),
		Body:  &issueBody,
	}
	issue, resp, err := gh.Issues.Create(context.TODO(), repoOwner, repoName, &i)
	if err != nil {
		return -1, fmt.Errorf("Failed to create issue: %s", err)
	}
	if resp.StatusCode != 201 {
		return -1, fmt.Errorf("Failed to create issue: %s", resp.Status)
	}
	issueNumber = issue.GetNumber()
	slog.Debug("END createIssue", slog.Int("issueNumber", issueNumber))
	return issueNumber, nil
}

func updateIssue(issueNumber int) error {
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
		return fmt.Errorf("Failed to update issue: %s", err)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Failed to update issue: %s", resp.Status)
	}
	slog.Info("Issue updated", slog.Int("issueNumber", issueNumber))
	slog.Debug("END updateIssue")
	return nil
}

func pinIssue(issueNumber int) {
	slog.Debug("START pinIssue")
	exec.Command("gh", "issue", "pin", strconv.Itoa(issueNumber)).Run()
	slog.Info("Issue pinned", slog.Int("issueNumber", issueNumber))
	slog.Debug("END pinIssue")
}
