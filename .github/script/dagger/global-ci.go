package main

import (
	"context"
	"fmt"

	"dagger/dagger/internal/dagger"
)

func (m *Dagger) SemgrepCi(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) (string, error) {
	return dag.Container().
		From("semgrep/semgrep:latest").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		// Local
		WithExec([]string{"sh", "-c", "semgrep scan --metrics=off --error --config 'p/default' --config 'p/kubernetes' --config 'p/dockerfile'"}).
		// PR
		// WithExec([]string{"sh", "-c", "trufflehog --fail --no-update --github-actions--no-verification  git file://. --since-commit main"}).
		// TODO cron
		// WithExec([]string{"sh", "-c", "trufflehog --fail --no-update --github-actions--no-verification  git file://."}).
		Stdout(ctx)
}

func (m *Dagger) GitleaksCi(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) (string, error) {
	return dag.Container().
		From("ghcr.io/gitleaks/gitleaks:latest").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		// Local
		WithExec([]string{"sh", "-c", "gitleaks git --verbose --redact=80"}).
		// PR
		// WithExec([]string{"sh", "-c", "trufflehog --fail --no-update --github-actions--no-verification  git file://. --since-commit main"}).
		// TODO cron
		// WithExec([]string{"sh", "-c", "trufflehog --fail --no-update --github-actions--no-verification  git file://."}).
		Stdout(ctx)
}

func (m *Dagger) GlobalCi(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) (string, error) {
	s1, err1 := m.SemgrepCi(ctx, source)
	s2, err2 := m.GitleaksCi(ctx, source)
	fs := s1 + s2
	ferr := fmt.Errorf("%v %v", err1, err2)
	if ferr != nil {
		return fs, ferr
	}
	return fs, nil
}
