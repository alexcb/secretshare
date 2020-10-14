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
    RUN go get golang.org/x/crypto/ssh
    RUN go get github.com/gordonklaus/ineffassign
    SAVE IMAGE

code:
    FROM +deps
    COPY main.go ./
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
            main.go
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

