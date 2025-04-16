package main

// func (m *Dagger) SemgrepCi(
// 	ctx context.Context,
// 	// +defaultPath="/"
// 	source *dagger.Directory,
// ) (string, error) {
// 	return dag.Container().
// 		From("semgrep/semgrep:latest").
// 		WithExec([]string{"apk", "add", "--no-cache", "jq"}).
// 		WithMountedDirectory("/src", source).
// 		WithWorkdir("/src").
// 		WithEnvVariable(findingsJsonPathVarName, "/tmp/findings.json").
// 		WithExec([]string{"sh", "-c", "semgrep scan --metrics=off --no-error --sarif --output=${" + findingsJsonPathVarName + "} --config 'p/default' --config 'p/kubernetes' --config 'p/dockerfile'"}).
// 		WithExec([]string{"sh", "-c", jqSarifToGithubAnnotations}).
// 		Stdout(ctx)
// }

// func (m *Dagger) GitleaksCi(
// 	ctx context.Context,
// 	// +defaultPath="/"
// 	source *dagger.Directory,
// ) (string, error) {
// 	return dag.Container().
// 		From("ghcr.io/gitleaks/gitleaks:latest").
// 		WithExec([]string{"apk", "add", "--no-cache", "jq"}).
// 		WithMountedDirectory("/src", source).
// 		WithWorkdir("/src").
// 		WithEnvVariable("FINDINGS_PATH", "/tmp/findings.json").
// 		WithExec([]string{"sh", "-c", "gitleaks git --exit-code 0 --verbose --redact=80 --platform github --report-format sarif --report-path ${FINDINGS_PATH}"}).
// 		WithExec([]string{"sh", "-c", jqSarifToGithubAnnotations}).
// 		Stdout(ctx)
// }

// func (m *Dagger) GlobalCi(
// 	ctx context.Context,
// 	// +defaultPath="/"
// 	source *dagger.Directory,
// ) (string, error) {
// 	s1, err1 := m.SemgrepCi(ctx, source)
// 	s2, err2 := m.GitleaksCi(ctx, source)
// 	fs := s1 + s2
// 	ferr := fmt.Errorf("%v\n%v", err1, err2)
// 	if ferr != nil {
// 		return fs, ferr
// 	}
// 	return fs, nil
// }
