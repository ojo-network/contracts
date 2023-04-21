# syntax=docker/dockerfile:1

ARG GO_VERSION="1.18"
ARG RUNNER_IMAGE="alpine"

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine as builder

ARG GIT_VERSION
ARG GIT_COMMIT

RUN apk add --no-cache \
    ca-certificates \
    build-base \
    linux-headers

# Download go dependencies
WORKDIR /cw-relayer
COPY go.mod go.sum ./

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

# Copy the remaining files
COPY . .

# Build cw-relayer binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go build \
        -mod=readonly \
        -tags "muslc" \
        -trimpath \
        -o /cw-relayer/build/cw-relayer \
        /cw-relayer/main.go

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM ${RUNNER_IMAGE}

RUN apk add bash \
    libgcc \
    jq

COPY --from=builder /cw-relayer/build/cw-relayer /usr/bin/cw-relayer

ENV HOME /
WORKDIR $HOME