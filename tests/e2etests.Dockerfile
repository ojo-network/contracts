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
COPY ./cw-relayer/go.mod ./cw-relayer/go.sum ./

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

# Copy the remaining files
COPY ./cw-relayer .

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
COPY --from=builder /cw-relayer/config.toml /usr/local/config.toml

# Copy Hardhat project and install dependencies
COPY /evm /evm
WORKDIR /evm

COPY --from=builder /cw-relayer/tests/e2e/config/relayer_bootstrap.sh /evm/relayer_bootstrap.sh

RUN apk add nodejs yarn
RUN yarn install

# Run Hardhat local node and deploy smart contracts

ENTRYPOINT ["sh", "-c", "yarn hardhat node --hostname 0.0.0.0"]
