FROM golang:1.23-bullseye AS build

RUN --mount=type=secret,id=github_token \
    if [ ! -s /run/secrets/github_token ]; then echo "github token is not set via secrets" >&2; exit 1; fi; \
    git config --global url."https://$(cat /run/secrets/github_token):x-oauth-basic@github.com/0xPellNetwork".insteadOf "https://github.com/0xPellNetwork"

ARG HTTP_PROXY
ARG HTTPS_PROXY

## Setup env
WORKDIR /app/pelldvs

# Fetch dependencies separately (for layer caching)
COPY go.mod go.sum ./
RUN --mount=type=cache,target="/go/pkg/mod" \
    go mod download

COPY . .
ARG DEBUG_ENABLED
RUN --mount=type=cache,target="/go/pkg/mod" \
    --mount=type=cache,target="/root/.cache/go-build" \
    if [ "${DEBUG_ENABLED}" = "true" ]; then \
        make build-debug; \
    else \
        make build; \
    fi

WORKDIR /app/
ARG EMULATOR_VERSION
## install pell-emulator
RUN git clone https://github.com/0xPellNetwork/pell-emulator.git --branch $EMULATOR_VERSION --depth 1
########## Install contracts ##########
WORKDIR /app/pell-emulator
RUN --mount=type=cache,target="/go/pkg/mod" \
    --mount=type=cache,target="/root/.cache/go-build" \
    make build

########## Setup runtime env ##########
FROM golang:1.23-bullseye AS runtime
RUN apt-get update -yqq && apt-get install -yqq openssh-server curl jq less
RUN mkdir -p /run/sshd && chmod 0755 /run/sshd

RUN curl -L https://foundry.paradigm.xyz | bash
ENV PATH="$PATH:/root/.foundry/bin"
RUN foundryup

RUN go install github.com/mikefarah/yq/v4@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
COPY --from=build /app/pelldvs/build/pelldvs /usr/bin/pelldvs
COPY --from=build /app/pell-emulator/build/pell-emulator /usr/bin/pell-emulator
COPY ./test/e2e/docker/pelldvs/scripts /root/scripts

COPY ./test/e2e/docker/ssh /root/.ssh
RUN chmod 600 /root/.ssh/*


WORKDIR /root

# expose port
EXPOSE 22

CMD ["bash"]
