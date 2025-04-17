# TODO add renovate tracking
# Base image version
ARG ALPINE_MAJOR_MINOR_VERSION=3.21
ARG ALPINE_MAJOR_MINOR_PATCH_VERSION=3.21.3

# Tools versions

## Single binary copy from container
ARG HADOLINT_VERSION=v2.12.0
ARG ACTIONLINT_VERSION=1.7.7
ARG TRUFFLEHOG_VERSION=3.88.23
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
FROM --platform=${BUILDPLATFORM} docker.io/trufflesecurity/trufflehog:${TRUFFLEHOG_VERSION} AS trufflehog
FROM --platform=${BUILDPLATFORM} docker.io/semgrep/semgrep:${SEMGREP_VERSION} AS semgrep
FROM --platform=${BUILDPLATFORM} docker.io/golangci/golangci-lint:${GOLANGCI_LINT_VERSION} AS golangci_lint
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

FROM --platform=${BUILDPLATFORM} base AS clean

# Copy tools from the respective stages
COPY --from=hadolint /bin/hadolint /usr/local/bin/hadolint
COPY --from=actionlint /usr/local/bin/actionlint /usr/local/bin/actionlint
COPY --from=trufflehog /usr/bin/trufflehog /usr/local/bin/trufflehog
COPY --from=semgrep /usr/bin/semgrep /usr/local/bin/semgrep
COPY --from=golangci_lint /usr/bin/golangci-lint /usr/local/bin/golangci-lint
COPY --from=yamllint /app/bin/yamllint /usr/local/bin/yamllint
COPY --from=markdownlint /usr/local/bin/markdownlint-cli2 /usr/local/bin/markdownlint
COPY --from=shellcheck /bin/shellcheck /usr/local/bin/shellcheck
COPY --from=kubectl /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/kubectl
COPY --from=pulumi /usr/bin/pulumi /usr/local/bin/pulumi
COPY --from=grype /grype /usr/local/bin/grype
COPY --from=syft /syft /usr/local/bin/syft
COPY --from=goreleaser /usr/bin/goreleaser /usr/local/bin/goreleaser
COPY --from=golang /usr/local/go/ /usr/local/go/
COPY --from=nodejs /usr/local /usr/local
COPY --from=python /usr/local /usr/local

# Docker buildx ARGs
ARG TARGETOS
ARG TARGETARCH

# Single binary tar archives
# TODO add renovate tracking
# TODO move this closer to the top of the file if possible (ARG scope is limited to the stage)
ARG HELM_VERSION=v3.17.3
ARG GITHUB_CLI_VERSION=2.70.0
ARG RELEASE_PLEASE_VERSION=17.0.0

# Install helm
RUN curl -sSL "https://get.helm.sh/helm-${HELM_VERSION}-${TARGETOS}-${TARGETARCH}.tar.gz" | tar -xz --strip-components=1 -C /usr/local/bin "${TARGETOS}-${TARGETARCH}/helm"

# Install gh (note the v prefix after download)
RUN curl -sSL "https://github.com/cli/cli/releases/download/v${GITHUB_CLI_VERSION}/gh_${GITHUB_CLI_VERSION}_${TARGETOS}_${TARGETARCH}.tar.gz" | tar -xz --strip-components=2 -C /usr/local/bin "gh_${GITHUB_CLI_VERSION}_${TARGETOS}_${TARGETARCH}/bin/gh"

# Install release-please
RUN npm --cache=false install --global "release-please@${RELEASE_PLEASE_VERSION}"

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
RUN mkdir -p /src
WORKDIR /src

CMD [ "bash" ]
