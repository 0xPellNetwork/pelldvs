# Use a build arg to ensure that both stages use the same,
# hopefully current, go version.
ARG GOLANG_BASE_IMAGE=golang:1.23-alpine

# stage 1 Generate Pelldvs Binary
FROM --platform=$BUILDPLATFORM $GOLANG_BASE_IMAGE AS builder
RUN apk update && \
    apk upgrade && \
    apk --no-cache add make git
COPY / /pelldvs
WORKDIR /pelldvs

RUN --mount=type=secret,id=github_token \
    git config --global url."https://$(cat /run/secrets/github_token):x-oauth-basic@github.com/0xPellNetwork/".insteadOf "https://github.com/0xPellNetwork/"

RUN --mount=type=cache,target="/go/pkg/mod" \
    --mount=type=cache,target="/root/.cache/go-build" \
    TARGETPLATFORM=$TARGETPLATFORM make build-linux

# stage 2
FROM $GOLANG_BASE_IMAGE
LABEL maintainer="hello@informal.systems"

# PellDVS will be looking for the genesis file in /pelldvs/config/genesis.json
# (unless you change `genesis_file` in config.toml). You can put your config.toml and
# private validator file into /pelldvs/config.
#
# The /pelldvs/data dir is used by PellDVS to store state.
ENV PELLDVSHOME=/pelldvs

# OS environment setup
# Set user right away for determinism, create directory for persistence and give our user ownership
# jq and curl used for extracting `pub_key` from private validator while
# deploying PellDVS with Kubernetes. It is nice to have bash so the users
# could execute bash commands.
RUN apk update && \
    apk upgrade && \
    apk --no-cache add curl jq bash && \
    addgroup tmuser && \
    adduser -S -G tmuser tmuser -h "$PELLDVSHOME"

# Run the container with tmuser by default. (UID=100, GID=1000)
USER tmuser

WORKDIR $PELLDVSHOME

# p2p, rpc and prometheus port
EXPOSE 26656 26657 26660

STOPSIGNAL SIGTERM

COPY --from=builder /pelldvs/build/pelldvs /usr/bin/pelldvs

# You can overwrite these before the first run to influence
# config.json and genesis.json. Additionally, you can override
# CMD to add parameters to `pelldvs node`.
ENV PROXY_APP=kvstore MONIKER=dockernode CHAIN_ID=dockerchain

COPY ./DOCKER/docker-entrypoint.sh /usr/local/bin/

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["node"]

# Expose the data directory as a volume since there's mutable state in there
VOLUME [ "$PELLDVSHOME" ]

