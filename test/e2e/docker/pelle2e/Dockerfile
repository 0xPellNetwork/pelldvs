FROM golang:1.23-bullseye AS build

ARG GITHUB_TOKEN
ARG HTTP_PROXY
ARG HTTPS_PROXY

## Setup env
WORKDIR /app/pelldvs

ARG GITHUB_TOKEN
ARG DEBUG_ENABLED
ARG PELLDVS_VERSION
RUN if [ -z "$GITHUB_TOKEN" ]; then echo "GITHUB_TOKEN is not set" && exit 1; fi
RUN git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/0xPellNetwork".insteadOf "https://github.com/0xPellNetwork"

# Fetch dependencies separately (for layer caching)
COPY go.mod go.sum ./
RUN --mount=type=cache,target="/go/pkg/mod" \
    go mod download

COPY . .
#RUN --mount=type=cache,target="/go/pkg/mod" \
#    --mount=type=cache,target="/root/.cache/go-build" \
#    if [ "${DEBUG_ENABLED}" = "true" ]; then \
#        make build-debug; \
#    else \
#        make build; \
#    fi

########## Build pelle2e ##########
RUN --mount=type=cache,target="/go/pkg/mod" \
    --mount=type=cache,target="/root/.cache/go-build" \
    go build -o build/pelle2e test/e2e/pelle2e/cmd/pelle2e/main.go

########## Setup runtime env ##########
FROM golang:1.23-bullseye AS runtime
RUN apt-get update -yqq && apt-get install -yqq openssh-server curl jq less
RUN mkdir -p /run/sshd && chmod 0755 /run/sshd

RUN curl -L https://foundry.paradigm.xyz | bash
ENV PATH="$PATH:/root/.foundry/bin"
RUN foundryup

RUN go install github.com/mikefarah/yq/v4@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
COPY --from=build /app/pelldvs/build/pelle2e /usr/bin/pelle2e
COPY ./test/e2e/docker/pelldvs/scripts /root/scripts

COPY ./test/e2e/docker/ssh /root/.ssh
RUN chmod 600 /root/.ssh/*


WORKDIR /root

# expose port
EXPOSE 22

CMD ["bash"]
