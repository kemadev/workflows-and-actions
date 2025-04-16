package main

// func (m *Dagger) YamlCi(
// 	ctx context.Context,
// 	// +defaultPath="/"
// 	source *dagger.Directory,
// ) (string, error) {
// 	const yamllintConfigPath = "./config/reusable/.yamllint.yaml"
// 	return dag.Container().
// 		From("pipelinecomponents/yamllint:latest").
// 		WithMountedDirectory("/src", source).
// 		WithWorkdir("/src").
// 		WithEnvVariable(findingsJsonPathVarName, "/tmp/findings.json").
// 		WithExec([]string{"sh", "-c", `
// find . -name "*.yml" -print0 | xargs -I {} echo ::warning title=Unsupported .yml file extension,file={},line=0,col=0::Please use .yaml file extension
// `}).
// 		WithExec([]string{"sh", "-c", `
// yamllint --format github --config-file ` + yamllintConfigPath + ` $(find . -name "*.yaml" -print)
// `}).
// 		Stdout(ctx)
// }
