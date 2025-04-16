package main

type Dagger struct{}

const (
	jqSarifToGithubAnnotations = `
if [ ! -f "${FINDINGS_PATH}" ]; then
	echo "Findings file not found!"
	exit 1
fi
results=$(jq -r '.runs[].results[]' "${FINDINGS_PATH}")
if [ -z "${results}" ]; then
	echo "No finding found in findings file!"
	exit 0
fi
jq -r '.runs[].results[] | "::error file=\(.locations[0].physicalLocation.artifactLocation.uri),line=\(.locations[0].physicalLocation.region.startLine),endLine=\(.locations[0].physicalLocation.region.endLine),col=\(.locations[0].physicalLocation.region.startColumn),endColumn=\(.locations[0].physicalLocation.region.endColumn)::\(.ruleId) - \(.message.text)"' "${FINDINGS_PATH}"
if [ $? -ne 0 ]; then
	echo "Error processing findings file!"
	exit 1
fi
exit 1
`
	findingsJsonPathVarName = "FINDINGS_PATH"
)
