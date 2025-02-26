FROM node:18

ARG GITHUB_TOKEN
ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG CONTRACT_VERSION

RUN apt-get update -yqq && apt-get install -yqq openssh-server curl jq
RUN mkdir -p /run/sshd && chmod 0755 /run/sshd

COPY ./docker/ssh /root/.ssh
RUN chmod 700 /root/.ssh
RUN chmod 600 /root/.ssh/*

########## Install dev env ##########
COPY ./docker/contracts/scripts/install_dev_env.sh /app/scripts/install_dev_env.sh
RUN bash /app/scripts/install_dev_env.sh

########## Setup env ##########
WORKDIR /app

RUN curl -L https://foundry.paradigm.xyz | bash
ENV PATH="$PATH:/root/.foundry/bin"
RUN foundryup

ARG GITHUB_TOKEN
RUN if [ -z "$GITHUB_TOKEN" ]; then echo "GITHUB_TOKEN is not set" && exit 1; fi
RUN git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/0xPellNetwork".insteadOf "https://github.com/0xPellNetwork"

RUN git clone https://github.com/0xPellNetwork/pell-middleware-contracts.git --branch $CONTRACT_VERSION --depth 1

########## Install contracts ##########
WORKDIR /app/pell-middleware-contracts

RUN forge install && yarn install --frozen-lockfile
RUN cd lib/pell-contracts && yarn install --frozen-lockfile \
    && export ACCOUNT_SECRETKEY=0x0000000000000000000000000000000000000000000000000000000000000000 \
    && forge compile && npx hardhat compile
RUN export ACCOUNT_SECRETKEY=0x0000000000000000000000000000000000000000000000000000000000000000 \
    && forge compile && npx hardhat compile

# expose port
EXPOSE 8545
EXPOSE 22

COPY ./docker/contracts/.env.contracts .env
COPY ./docker/contracts/scripts/docker_init.sh ./scripts/docker_init.sh

# expose port
EXPOSE 8545
EXPOSE 22

# run hardhat node
CMD /app/pell-middleware-contracts/scripts/docker_init.sh

