FROM pelldvs/e2e-pelldvs:latest AS build

########## Setup runtime env ##########
FROM golang:1.23-bookworm AS runtime
RUN apt-get update -yqq && apt-get install -yqq openssh-server curl jq less
RUN mkdir -p /run/sshd && chmod 0755 /run/sshd

RUN curl -L https://foundry.paradigm.xyz | bash
ENV PATH="$PATH:/root/.foundry/bin"
RUN foundryup

RUN go install github.com/mikefarah/yq/v4@v4.45.2
COPY --from=build /app/pelldvs/build/pelldvs /usr/bin/pelldvs
COPY ./docker/operator/scripts /root/scripts
COPY ./docker/ssh /root/.ssh

RUN chmod 600 /root/.ssh/*

WORKDIR /root

# expose port
EXPOSE 22
