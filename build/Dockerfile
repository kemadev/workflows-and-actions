# TODO add renovate tracking
# Base image version
ARG ALPINE_MAJOR_MINOR_VERSION=3.21
ARG ALPINE_MAJOR_MINOR_PATCH_VERSION=3.21.3

# Tools versions

## Single binary copy from container
ARG HADOLINT_VERSION=v2.12.0
ARG ACTIONLINT_VERSION=1.7.7
ARG GITLEAKS_VERSION=v8.24.3
ARG SEMGREP_VERSION=1.119.0
ARG GOLANGCI_LINT_VERSION=v2.1.2
ARG YAMLLINT_VERSION=0.34.0
ARG MARKDOWNLINT_VERSION=v0.17.2
ARG SHELLCHECK_VERSION=v0.10.0
ARG KUBECTL_VERSION=1.32.3
ARG PULUMI_VERSION=3.162.0
ARG GRYPE_VERSION=v0.91.0
ARG SYFT_VERSION=v1.22.0
ARG GORELEASER_VERSION=v2.8.2
ARG GO_VERSION=1.24.2

## Multi-directory copy from container
ARG NODEJS_VERSION=23.11.0
ARG PYTHON_VERSION=3.13.3

# Define images to copy from (can't use ARG in COPY statements)
FROM --platform=${BUILDPLATFORM} docker.io/hadolint/hadolint:${HADOLINT_VERSION} AS hadolint
FROM --platform=${BUILDPLATFORM} docker.io/rhysd/actionlint:${ACTIONLINT_VERSION} AS actionlint
FROM --platform=${BUILDPLATFORM} docker.io/zricethezav/gitleaks:${GITLEAKS_VERSION} AS gitleaks
FROM --platform=${BUILDPLATFORM} docker.io/semgrep/semgrep:${SEMGREP_VERSION} AS semgrep
FROM --platform=${BUILDPLATFORM} docker.io/golangci/golangci-lint:${GOLANGCI_LINT_VERSION} AS golangci-lint
FROM --platform=${BUILDPLATFORM} docker.io/pipelinecomponents/yamllint:${YAMLLINT_VERSION} AS yamllint
FROM --platform=${BUILDPLATFORM} docker.io/davidanson/markdownlint-cli2:${MARKDOWNLINT_VERSION} AS markdownlint
FROM --platform=${BUILDPLATFORM} docker.io/koalaman/shellcheck:${SHELLCHECK_VERSION} AS shellcheck
FROM --platform=${BUILDPLATFORM} docker.io/bitnami/kubectl:${KUBECTL_VERSION} AS kubectl
FROM --platform=${BUILDPLATFORM} docker.io/pulumi/pulumi:${PULUMI_VERSION} AS pulumi
FROM --platform=${BUILDPLATFORM} docker.io/anchore/grype:${GRYPE_VERSION} AS grype
FROM --platform=${BUILDPLATFORM} docker.io/anchore/syft:${SYFT_VERSION} AS syft
FROM --platform=${BUILDPLATFORM} docker.io/goreleaser/goreleaser:${GORELEASER_VERSION} AS goreleaser
FROM --platform=${BUILDPLATFORM} docker.io/golang:${GO_VERSION}-alpine${ALPINE_MAJOR_MINOR_VERSION} AS golang
FROM --platform=${BUILDPLATFORM} docker.io/node:${NODEJS_VERSION}-alpine${ALPINE_MAJOR_MINOR_VERSION} AS nodejs
FROM --platform=${BUILDPLATFORM} docker.io/python:${PYTHON_VERSION}-alpine${ALPINE_MAJOR_MINOR_VERSION} AS python

# Define base image
FROM --platform=${BUILDPLATFORM} docker.io/alpine:${ALPINE_MAJOR_MINOR_PATCH_VERSION} AS base

# Install core dependencies
# hadolint ignore=DL3018
RUN apk add --no-cache \
	libstdc++ \
	bash \
	curl \
	jq \
	git \
	tar \
	gzip \
	zip

SHELL ["/bin/bash", "-uo", "pipefail", "-c"]

FROM --platform=${BUILDPLATFORM} base AS misc-deps

# Docker buildx ARGs
ARG TARGETOS
ARG TARGETARCH

# Misc tools (can't add those args with ones above as ARG has stage scope)
# TODO add renovate tracking
ARG HELM_VERSION=v3.17.3
ARG GITHUB_CLI_VERSION=2.70.0
ARG RELEASE_PLEASE_VERSION=17.0.0

# Install helm
RUN curl -sSL "https://get.helm.sh/helm-${HELM_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz" | tar -xz --strip-components=1 -C /usr/local/bin "${TARGETOS}-${TARGETARCH}/helm"

# Install gh (note the v prefix after download)
RUN curl -sSL "https://github.com/cli/cli/releases/download/v${GITHUB_CLI_VERSION}/gh_${GITHUB_CLI_VERSION}_${TARGETOS}_${TARGETARCH}.tar.gz" | tar -xz --strip-components=2 -C /usr/local/bin "gh_${GITHUB_CLI_VERSION}_${TARGETOS}_${TARGETARCH}/bin/gh"

# Install nodejs
COPY --link --from=nodejs /usr/local /usr/local

# Install release-please
RUN npm --cache=false install --global "release-please@${RELEASE_PLEASE_VERSION}"

# Intermediate image providing dependencies for nodejs
FROM --platform=${BUILDPLATFORM} nodejs AS nodejs-deps

# Cleanup nodejs mess
RUN rm -rf /usr/local/CHANGELOG.md /usr/local/LICENSE /usr/local/README.md /usr/local/bin/yarn /usr/local/bin/yarnpkg /usr/local/bin/docker-entrypoint.sh

# Final image
FROM --platform=${BUILDPLATFORM} base AS clean

# Copy tools from the respective stages
COPY --link --from=hadolint /bin/hadolint /usr/local/bin/hadolint
COPY --link --from=actionlint /usr/local/bin/actionlint /usr/local/bin/actionlint
COPY --link --from=gitleaks /usr/bin/gitleaks /usr/local/bin/gitleaks
COPY --link --from=semgrep /usr/bin/semgrep /usr/local/bin/semgrep
COPY --link --from=golangci-lint /usr/bin/golangci-lint /usr/local/bin/golangci-lint
COPY --link --from=yamllint /app/bin/yamllint /usr/local/bin/yamllint
COPY --link --from=markdownlint /usr/local/bin/markdownlint-cli2 /usr/local/bin/markdownlint
COPY --link --from=shellcheck /bin/shellcheck /usr/local/bin/shellcheck
COPY --link --from=kubectl /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/kubectl
COPY --link --from=pulumi /usr/bin/pulumi /usr/local/bin/pulumi
COPY --link --from=grype /grype /usr/local/bin/grype
COPY --link --from=syft /syft /usr/local/bin/syft
COPY --link --from=goreleaser /usr/bin/goreleaser /usr/local/bin/goreleaser
COPY --link --from=golang /usr/local/go/ /usr/local/go/
COPY --link --from=nodejs-deps /usr/local/ /usr/local/
COPY --link --from=python /usr/local/ /usr/local/
COPY --link --from=misc-deps /usr/local/bin/helm /usr/local/bin/helm
COPY --link --from=misc-deps /usr/local/bin/gh /usr/local/bin/gh
COPY --link --from=misc-deps /usr/local/lib/node_modules/release-please /usr/local/lib/node_modules/release-please

# Link nodejs tools
RUN ln -s /usr/local/lib/node_modules/release-please/build/src/bin/release-please.js /usr/local/bin/release-please

# Make all tools root-owned and executable
RUN chmod 755 /usr/local/bin/* && chown root:root /usr/local/bin/* && chmod 755 /usr/local/go/bin/* && chown root:root /usr/local/go/bin/*

# Configure Go
ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV PATH="${PATH}:${GOROOT}/bin:${GOPATH}/bin"

# Git args
ARG GIT_USERNAME="github-actions[bot]"
ARG GIT_EMAIL="kemadev+github-actions[bot]@users.noreply.github.com"

# Configure git
RUN git config --global user.name "${GIT_USERNAME}" && git config --global user.email "${GIT_EMAIL}"

# Create base directory
RUN mkdir -p /src && chown root:root /src
WORKDIR /src

# Run as root
# hadolint ignore=DL3002
USER root:root

CMD [ "bash" ]

# Final image as non-root
FROM --platform=${BUILDPLATFORM} clean AS clean-nonroot

# Set home directory
ENV HOME=/home/nobody

# Create home directory
RUN mkdir -p ${HOME} && chown nobody:nobody ${HOME}

# Run as non-root
USER nobody:nobody

# Final image for workflows
FROM --platform=${BUILDPLATFORM} clean AS workflow-base

# Create tool directory
RUN mkdir -p /tool && chown root:root /tool

WORKDIR /tool/pkg/pkg/ci/run

# Add cache layer for dependencies, mount from host to avoid COPY layer, verbose output for cache behavior logging
RUN --mount=type=cache,target=/root/go/pkg \
	--mount=type=bind,source=./pkg/pkg/ci/run/go.mod,target=go.mod \
	--mount=type=bind,source=./pkg/pkg/ci/run/go.sum,target=go.sum \
	go mod download -x

# Add cache layer for dependencies, mount from host to avoid COPY layer
RUN --mount=type=cache,target=/root/go/pkg \
	--mount=type=cache,target=/root/.cache \
	--mount=type=bind,source=./pkg,target=/tool/pkg \
	go build -ldflags='-s -w' -o /usr/local/bin/kema-ci .

FROM --platform=${BUILDPLATFORM} clean AS workflow-clean

COPY --link --from=workflow-base /usr/local/bin/kema-ci /usr/local/bin/kema-ci

ENTRYPOINT [ "/usr/local/bin/kema-ci" ]
