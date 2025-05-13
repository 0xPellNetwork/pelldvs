FROM golang:1.23-bullseye AS build

RUN --mount=type=secret,id=github_token \
    if [ ! -s /run/secrets/github_token ]; then echo "github token is not set via secrets" >&2; exit 1; fi; \
    git config --global url."https://$(cat /run/secrets/github_token):x-oauth-basic@github.com/0xPellNetwork/".insteadOf "https://github.com/0xPellNetwork/"

ARG HTTP_PROXY
ARG HTTPS_PROXY

## Setup env
WORKDIR /app/pelldvs


ARG DEBUG_ENABLED
ARG PELLDVS_VERSION
COPY go.mod go.sum ./
RUN --mount=type=cache,target="/go/pkg/mod" \
    go mod download

COPY . .
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

RUN go install github.com/mikefarah/yq/v4@v4.45.2
RUN go install github.com/go-delve/delve/cmd/dlv@latest
COPY --from=build /app/pelldvs/build/pelle2e /usr/bin/pelle2e
COPY ./test/e2e/docker/pelldvs/scripts /root/scripts

COPY ./test/e2e/docker/ssh /root/.ssh
RUN chmod 600 /root/.ssh/*


WORKDIR /root

# expose port
EXPOSE 22

CMD ["bash"]
