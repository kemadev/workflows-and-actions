package main

import (
	"log"
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

func logFatal(msg string, err error) {
	log.Fatalf("::error title=%s::%s", msg, err.Error())
}

func checkVariables() error {
	if GORELEASER_CONFIG_TEMPLATE_DIR == "" {
		logFatal("GORELEASER_CONFIG_TEMPLATE_DIR is not set", nil)
	}
	if GORELEASER_CONFIG_TEMPLATE_FILENAME == "" {
		logFatal("GORELEASER_CONFIG_TEMPLATE_FILENAME is not set", nil)
	}
	if GORELEASER_CONFIG_OUTPUT_FILE == "" {
		logFatal("GORELEASER_CONFIG_OUTPUT_FILE is not set", nil)
	}
	if BUILDS_DIR == "" {
		logFatal("BUILDS_DIR is not set", nil)
	}
	if BUILDS_DIR_PARENT == "" {
		logFatal("BUILDS_DIR_PARENT is not set", nil)
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

func renderGoreleaserConfig(dirs []string) {
	outputFile, err := os.Create(GORELEASER_CONFIG_OUTPUT_FILE)
	if err != nil {
		logFatal("Failed to create goreleaser config file", err)
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
		logFatal("Failed to parse goreleaser config template", nil)
	}
	err = tmpl.Execute(outputFile, data)
	if err != nil {
		logFatal("Failed to render goreleaser config", err)
	}
}

func main() {
	startDate := time.Now()
	log.Println("Rendering goreleaser config")
	checkVariables()
	dirs, err := listDirs(BUILDS_DIR_PARENT + "/" + BUILDS_DIR)
	if err != nil {
		logFatal("Failed to list directories", err)
	}
	renderGoreleaserConfig(dirs)
	log.Printf("Rendering goreleaser config took %v\n", time.Since(startDate))
}
