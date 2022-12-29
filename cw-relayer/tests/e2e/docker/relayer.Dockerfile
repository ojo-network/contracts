# syntax=docker/dockerfile:1
# https://github.com/osmosis-labs/osmosis/blob/v12.3.0/Makefile
# Modified to use the apline image instead of distroless and include bash

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

# Cosmwasm - Download correct libwasmvm version
RUN WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | cut -d ' ' -f 2) && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$(uname -m).a \
        -O /lib/libwasmvm_muslc.a && \
    # verify checksum
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep $(uname -m) | cut -d ' ' -f 1)

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
COPY --from=cosmwasm/wasmd:latest /usr/bin/wasmd /usr/bin/wasmd

ENV HOME /
WORKDIR $HOME

# tendermint p2p
EXPOSE 26656
# tendermint rpc
EXPOSE 26657
# grpc rpc
EXPOSE 8080

CMD ["cw-relayer"]
