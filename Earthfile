FROM golang:1.13-alpine3.11

RUN apk add --update --no-cache \
    bash \
    bash-completion \
    binutils \
    ca-certificates \
    coreutils \
    curl \
    findutils \
    g++ \
    git \
    grep \
    less \
    make \
    openssl \
    shellcheck \
    util-linux

WORKDIR /secretshare

deps:
    RUN go get golang.org/x/tools/cmd/goimports
    RUN go get golang.org/x/lint/golint
    COPY go.mod go.sum .
	RUN go mod download
    SAVE IMAGE

code:
    FROM +deps
    COPY --dir cmd ./
    SAVE IMAGE

lint:
    FROM +code
    RUN output="$(ineffassign .)" ; \
        if [ -n "$output" ]; then \
            echo "$output" ; \
            exit 1 ; \
        fi
    RUN output="$(goimports -d $(find . -type f -name '*.go' | grep -v \.pb\.go) 2>&1)"  ; \
        if [ -n "$output" ]; then \
            echo "$output" ; \
            exit 1 ; \
        fi
    RUN golint -set_exit_status ./...
    RUN output="$(go vet ./... 2>&1)" ; \
        if [ -n "$output" ]; then \
            echo "$output" ; \
            exit 1 ; \
        fi

secretshare:
    FROM +code
    ARG GOOS=linux
    ARG GOARCH=amd64
    ARG GO_EXTRA_LDFLAGS="-linkmode external -extldflags -static"
    RUN test -n "$GOOS" && test -n "$GOARCH"
    ARG GOCACHE=/go-cache
    RUN mkdir -p build
    RUN --mount=type=cache,target=$GOCACHE \
        go build \
            -o build/secretshare \
            -ldflags "$GO_EXTRA_LDFLAGS" \
            cmd/secretshare/main.go
    SAVE ARTIFACT build/secretshare AS LOCAL "build/$GOOS/$GOARCH/secretshare"

secretshare-darwin:
    COPY \
        --build-arg GOOS=darwin \
        --build-arg GOARCH=amd64 \
        --build-arg GO_EXTRA_LDFLAGS= \
        +secretshare/* ./
    SAVE ARTIFACT ./*

secretshare-all:
    COPY +secretshare/secretshare ./secretshare-linux-amd64
    COPY +secretshare-darwin/secretshare ./secretshare-darwin-amd64
    SAVE ARTIFACT ./*

release:
    FROM node:13.10.1-alpine3.11
    RUN npm install -g github-release-cli@v1.3.1
    WORKDIR /release
    COPY +secretshare/secretshare ./secretshare-linux-amd64
    COPY +secretshare-darwin/secretshare ./secretshare-darwin-amd64
    ARG RELEASE_TAG
    ARG EARTHLY_GIT_HASH
    ARG BODY="No details provided"
    RUN --secret GITHUB_TOKEN=+secrets/GITHUB_TOKEN test -n "$GITHUB_TOKEN"
    RUN --push \
        --secret GITHUB_TOKEN=+secrets/GITHUB_TOKEN \
        github-release upload \
        --owner alexcb \
        --repo secret-share \
        --commitish "$EARTHLY_GIT_HASH" \
        --tag "$RELEASE_TAG" \
        --name "$RELEASE_TAG" \
        --body "$BODY" \
        ./secretshare-*
