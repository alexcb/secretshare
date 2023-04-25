VERSION 0.7
FROM golang:1.20-alpine3.17

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
    openssh-keygen \
    shellcheck \
    util-linux



WORKDIR /secretshare

deps:
    RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.50.0
    COPY go.mod go.sum .
    RUN go mod download

code:
    FROM +deps
    COPY --dir cmd ./

lint:
    FROM +code
    COPY ./.golangci.yaml ./
    RUN golangci-lint run

secretshare:
    FROM +code
    ARG RELEASE_TAG="dev"
    ARG GOOS
    ARG GO_EXTRA_LDFLAGS
    ARG GOARCH
    RUN test -n "$GOOS" && test -n "$GOARCH"
    ARG GOCACHE=/go-cache
    RUN mkdir -p build
    RUN --mount=type=cache,target=$GOCACHE \
        go build \
            -o build/secretshare \
            -ldflags "-X main.Version=$RELEASE_TAG $GO_EXTRA_LDFLAGS" \
            cmd/secretshare/main.go
    SAVE ARTIFACT build/secretshare AS LOCAL "build/$GOOS/$GOARCH/secretshare"

secretshare-darwin-amd64:
    COPY \
        --build-arg GOOS=darwin \
        --build-arg GOARCH=amd64 \
        --build-arg GO_EXTRA_LDFLAGS= \
        +secretshare/secretshare /build/secretshare
    SAVE ARTIFACT /build/secretshare AS LOCAL "build/darwin/amd64/secretshare"

secretshare-darwin-arm64:
    COPY \
        --build-arg GOOS=darwin \
        --build-arg GOARCH=arm64 \
        --build-arg GO_EXTRA_LDFLAGS= \
        +secretshare/secretshare /build/secretshare
    SAVE ARTIFACT /build/secretshare AS LOCAL "build/darwin/arm64/secretshare"

secretshare-linux-amd64:
    COPY \
        --build-arg GOOS=linux \
        --build-arg GOARCH=amd64 \
        --build-arg GO_EXTRA_LDFLAGS="-linkmode external -extldflags -static" \
        +secretshare/secretshare /build/secretshare
    SAVE ARTIFACT /build/secretshare AS LOCAL "build/linux/amd64/secretshare"

secretshare-linux-arm64:
    COPY \
        --build-arg GOOS=linux \
        --build-arg GOARCH=arm64 \
        --build-arg GO_EXTRA_LDFLAGS= \
        +secretshare/secretshare /build/secretshare
    SAVE ARTIFACT /build/secretshare AS LOCAL "build/linux/arm64/secretshare"

secretshare-all:
    BUILD +secretshare-linux-amd64
    BUILD +secretshare-linux-arm64
    BUILD +secretshare-darwin-amd64
    BUILD +secretshare-darwin-arm64


test:
    COPY +secretshare-linux-amd64/secretshare ./secretshare
    RUN ./secretshare
    RUN bash -c "echo -n hello | openssl pkeyutl -encrypt -pubin -inkey <(ssh-keygen -f ~/.secretshare.pub -e -m PKCS8) -pkeyopt rsa_padding_mode:oaep -pkeyopt rsa_oaep_md:sha256 -pkeyopt rsa_mgf1_md:sha256 | base64 | ./secretshare > output"
    RUN bash -c "diff output <( echo -n hello)"

release:
    BUILD +test
    FROM node:13.10.1-alpine3.11
    RUN npm install -g github-release-cli@v1.3.1
    RUN apk add file
    WORKDIR /release
    COPY +secretshare-linux-amd64/secretshare ./secretshare-linux-amd64
    COPY +secretshare-linux-arm64/secretshare ./secretshare-linux-arm64
    COPY +secretshare-darwin-amd64/secretshare ./secretshare-darwin-amd64
    COPY +secretshare-darwin-arm64/secretshare ./secretshare-darwin-arm64
    RUN file secretshare-linux-amd64 | grep "ELF 64-bit LSB executable, x86-64"
    RUN file secretshare-linux-arm64 | grep "ELF 64-bit LSB executable, ARM aarch64"
    RUN file secretshare-darwin-amd64 | grep "Mach-O 64-bit x86_64 executable"
    RUN file secretshare-darwin-arm64 | grep "Mach-O 64-bit arm64 executable"
    ARG --required RELEASE_TAG
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
