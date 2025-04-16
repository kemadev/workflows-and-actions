package main

import (
	"context"

	"dagger/dagger/internal/dagger"
)

func (m *Dagger) DockerCi(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) (string, error) {
	return dag.Container().
		From("ghcr.io/hadolint/hadolint:latest-alpine").
		WithUser("root").
		WithExec([]string{"apk", "add", "--no-cache", "jq"}).
		WithUser("guest").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithEnvVariable("FINDINGS_PATH", "/tmp/findings.json").
		WithExec([]string{"sh", "-c", `
find . -name '*Dockerfile*' -print0 | xargs -0 -I {} sh -c 'hadolint --no-fail --format json {} >> ${FINDINGS_PATH}'
`}).
		WithExec([]string{"sh", "-c", `
if [ -f "${FINDINGS_PATH}" ] && [ "$(cat "${FINDINGS_PATH}")" != "[]" ]; then
	jq -r '.[] | "::error file=\(.file),line=\(.line),col=\(.column)::\(.message) - \(.code)"' "${FINDINGS_PATH}"
fi
`}).
		Stdout(ctx)
}
