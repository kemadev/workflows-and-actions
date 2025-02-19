declare -a unused_deps=()
find . -name go.mod -execdir go mod tidy \;
declare git_status=$(git status --porcelain)
if [ -n "${git_status}" ]; then
	for file in $(echo "${git_status}" | awk '{print $2}'); do
		# if [ "${file}" == "go.mod" ]; then
		# unused_deps+=($(git diff --name-only "${file}"))
		echo "Unused dependencies found in ${file}"
		# fi
	done
fi
