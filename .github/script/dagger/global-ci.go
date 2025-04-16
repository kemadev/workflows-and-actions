package main

import (
	"context"

	"dagger/dagger/internal/dagger"
)

func (m *Dagger) GlobalCi(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) (string, error) {
	return dag.Container().
		From("semgrep/semgrep:latest").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		// Local
		WithExec([]string{"sh", "-c", "semgrep scan --metrics=off --error --historical-secrets --config 'p/default' --config 'p/gitleaks' --config 'p/kubernetes' --config 'p/dockerfile'"}).
		// PR
		// WithExec([]string{"sh", "-c", "trufflehog --fail --no-update --github-actions--no-verification  git file://. --since-commit main"}).
		// TODO cron
		// WithExec([]string{"sh", "-c", "trufflehog --fail --no-update --github-actions--no-verification  git file://."}).
		Stdout(ctx)
}
