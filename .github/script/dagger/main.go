package main

type Dagger struct{}

const (
	findingsJsonPathVarName    = "FINDINGS_PATH"
	jqSarifToGithubAnnotations = `
if [ ! -f "${` + findingsJsonPathVarName + `}" ]; then
	echo "No findings file found!"
	exit 0
fi
results=$(jq -r '.runs[].results[]' "${` + findingsJsonPathVarName + `}")
if [ -z "${results}" ]; then
	echo "No finding found in findings file!"
	exit 0
fi
driver_name=$(jq -r '.runs[].tool.driver.name' "${` + findingsJsonPathVarName + `}")
if [ -z "${driver_name}" ]; then
	echo "No driver name found in findings file!"
	exit 1
fi
jq -r '.runs[].results[] | "::error title='"${driver_name}"',file=\(.locations[0].physicalLocation.artifactLocation.uri),line=\(.locations[0].physicalLocation.region.startLine),endLine=\(.locations[0].physicalLocation.region.endLine),col=\(.locations[0].physicalLocation.region.startColumn),endColumn=\(.locations[0].physicalLocation.region.endColumn)::\(.message.text)"' "${` + findingsJsonPathVarName + `}"
if [ $? -ne 0 ]; then
	echo "Error processing findings file!"
	exit 1
fi
exit 1
`
)
