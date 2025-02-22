package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/svu/pkg/svu"
)

var (
	// GitHub Actions provided GitHub authentication token
	ghToken = os.Getenv("GH_TOKEN")
	// GitHub branch that triggered the workflow
	branch = os.Getenv("GITHUB_REF_NAME")
	// GitHub output file
	ghOutput = os.Getenv("GITHUB_OUTPUT")
	// GitHub runner's temp directory
	runnerTemp = os.Getenv("RUNNER_TEMP")
	// File in which to write the tag version
	tagVersionFileName = os.Getenv("TAG_VERSION_FILE_NAME")
	// GitHub repository owner
	repoOwner = strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")[0]
	// GitHub repository name
	repoName = strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")[1]
)

const (
	// Production branch
	prodBranch = "main"
	// Prerelease prefix
	preReleasePrefix = "next"
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
	if branch == "" {
		return fmt.Errorf("Environment variable GITHUB_REF_NAME is not set")
	}
	if ghOutput == "" {
		return fmt.Errorf("Environment variable GITHUB_OUTPUT is not set")
	}
	if runnerTemp == "" {
		return fmt.Errorf("Environment variable RUNNER_TEMP is not set")
	}
	if tagVersionFileName == "" {
		return fmt.Errorf("Environment variable TAG_VERSION_FILE is not set")
	}
	if repoOwner == "" {
		return fmt.Errorf("Can't infer repository owner from GITHUB_REPOSITORY environment variable")
	}
	if repoName == "" {
		return fmt.Errorf("Can't infer repository name from GITHUB_REPOSITORY environment variable")
	}
	return nil
}

func checkForcePatchVersion() (bool, error) {
	// if first arg is true, force patch version
	if len(os.Args) > 1 {
		if os.Args[1] == "" {
			return false, nil
		}
		forcePatchVersion, err := strconv.ParseBool(os.Args[1])
		if err != nil {
			return false, err
		}
		slog.Debug("Force patch version argument provided", slog.Bool("forcePatchVersion", forcePatchVersion))
		return forcePatchVersion, nil
	}
	slog.Debug("No force patch version argument provided")
	return false, nil
}

func getNextVersion(forcePatchVersion bool) (string, error) {
	if branch != prodBranch {
		nextVersion, err := svu.PreRelease(svu.WithPrefix("v"), svu.WithPreRelease(preReleasePrefix), svu.WithForcePatchIncrement(forcePatchVersion))
		if err != nil {
			return "", err
		}
		slog.Debug("Next version is pre-release", slog.String("nextVersion", nextVersion))
		return nextVersion, nil
	}
	nextVersion, err := svu.Next(svu.WithPrefix("v"), svu.WithForcePatchIncrement(forcePatchVersion))
	if err != nil {
		return "", err
	}
	slog.Debug("Next version is release", slog.String("nextVersion", nextVersion))
	return nextVersion, nil
}

func outputTagVersion(version string) error {
	file, err := os.OpenFile(runnerTemp+"/"+tagVersionFileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = fmt.Fprintf(file, "%s", version)
	if err != nil {
		return err
	}
	slog.Debug("Outputted tag version to file", slog.String("tagVersionFileName", tagVersionFileName))
	return nil
}

func main() {
	startTime := time.Now()
	defer func() {
		slog.Info("Execution time", slog.String("duration", time.Since(startTime).String()))
	}()
	initLogger()
	err := checkVariables()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	forcePatchVersion, err := checkForcePatchVersion()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	nextVersion, err := getNextVersion(forcePatchVersion)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	fmt.Println("Next version is " + nextVersion)
	err = outputTagVersion(nextVersion)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	fmt.Println("Outputted tag version to file")
}
