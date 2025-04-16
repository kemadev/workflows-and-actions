package main

import (
	"context"

	"dagger/dagger/internal/dagger"
)

func (m *Dagger) GHActionsCi(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) (string, error) {
	f, err := dag.Container().
		From("rhysd/actionlint:latest").
		WithUser("root").
		WithExec([]string{"apk", "add", "--no-cache", "jq"}).
		WithUser("guest").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithEnvVariable("FINDINGS_PATH", "/tmp/findings.json").
		WithExec([]string{"sh", "-c", "actionlint -format '{{json .}}' > ${FINDINGS_PATH} || true"}).
		WithExec([]string{"sh", "-c", `
if [ -f "${FINDINGS_PATH}" ] && [ "$(cat "${FINDINGS_PATH}")" != "[]" ]; then
	jq -r '.[] | "::error file=\(.filepath),line=\(.line),col=\(.column)::\(.message) - \(.kind)"' "${FINDINGS_PATH}"
fi
`}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}
	return f, nil
}
