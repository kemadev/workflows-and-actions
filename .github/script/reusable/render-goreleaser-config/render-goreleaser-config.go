package main

import (
	"fmt"
	"log/slog"
	"os"
	"text/template"
	"time"
)

var (
	// Path to the directory containing the goreleaser config template file
	GORELEASER_CONFIG_TEMPLATE_DIR = os.Getenv("GORELEASER_CONFIG_TEMPLATE_DIR")
	// Name of the goreleaser config template file
	GORELEASER_CONFIG_TEMPLATE_FILENAME = os.Getenv("GORELEASER_CONFIG_TEMPLATE_FILENAME")
	// Path this script should output the rendered goreleaser config file to
	GORELEASER_CONFIG_OUTPUT_FILE = os.Getenv("GORELEASER_CONFIG_OUTPUT_FILE")
	// Parent directory containing the directories to build
	BUILDS_DIR_PARENT = os.Getenv("BUILDS_DIR_PARENT")
	// Directory containing the directories to build
	BUILDS_DIR = os.Getenv("BUILDS_DIR")
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
	if GORELEASER_CONFIG_TEMPLATE_DIR == "" {
		return fmt.Errorf("GORELEASER_CONFIG_TEMPLATE_DIR is not set")
	}
	if GORELEASER_CONFIG_TEMPLATE_FILENAME == "" {
		return fmt.Errorf("GORELEASER_CONFIG_TEMPLATE_FILENAME is not set")
	}
	if GORELEASER_CONFIG_OUTPUT_FILE == "" {
		return fmt.Errorf("GORELEASER_CONFIG_OUTPUT_FILE is not set")
	}
	if BUILDS_DIR == "" {
		return fmt.Errorf("BUILDS_DIR is not set")
	}
	if BUILDS_DIR_PARENT == "" {
		return fmt.Errorf("BUILDS_DIR_PARENT is not set")
	}
	return nil
}

func listDirs(root string) ([]string, error) {
	var dirs []string
	files, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}
	return dirs, nil
}

func renderGoreleaserConfig(dirs []string) error {
	outputFile, err := os.Create(GORELEASER_CONFIG_OUTPUT_FILE)
	if err != nil {
		return err
	}
	type Build struct {
		ID     string
		CmdDir string
	}
	type Builds struct {
		Builds []Build
	}
	var data Builds
	for _, dir := range dirs {
		data.Builds = append(data.Builds, Build{
			ID:     dir,
			CmdDir: BUILDS_DIR,
		})
	}
	tmpl := template.Must(template.New(GORELEASER_CONFIG_TEMPLATE_FILENAME).ParseFiles(GORELEASER_CONFIG_TEMPLATE_DIR + "/" + GORELEASER_CONFIG_TEMPLATE_FILENAME))
	if tmpl == nil {
		return fmt.Errorf("Failed to parse goreleaser config template")
	}
	err = tmpl.Execute(outputFile, data)
	if err != nil {
		return err
	}
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
	dirs, err := listDirs(BUILDS_DIR_PARENT + "/" + BUILDS_DIR)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	err = renderGoreleaserConfig(dirs)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info("Rendered goreleaser config")
}
