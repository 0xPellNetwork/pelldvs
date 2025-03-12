#!/bin/bash

set -x
set -e

logt() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') $1"
}

function hardhat_healthcheck {
  set +e
  while true; do
    ssh hardhat "test -f /root/contracts_deployed_completed"
    if [ $? -eq 0 ]; then
      echo "Contracts deployed, proceeding to the next step..."
      break
    fi
    echo "Contracts not deployed, retrying in 1 second..."
    sleep 1
  done
  set -e
}

export REGISTRY_ROUTER_ADDRESS_FILE="/root/RegistryRouterAddress.json"

function load_defaults {
  export HARDHAT_CONTRACTS_PATH="/app/pell-middleware-contracts/lib/pell-contracts/deployments/localhost"
  export HARDHAT_CONTRACTS_PATH="/app/pell-middleware-contracts/lib/pell-contracts/deployments/localhost"
  export HARDHAT_PROJ_ROOT="/app/pell-middleware-contracts"
  export HARDHAT_DVS_PATH="/app/pell-middleware-contracts/deployments/localhost"

  export AGGREGATOR_INDEXER_START_HEIGHT=${AGGREGATOR_INDEXER_START_HEIGHT:-0}
  export AGGREGATOR_INDEXER_BATCH_SIZE=${AGGREGATOR_INDEXER_BATCH_SIZE:-1000}

  export PELLDVS_HOME=${PELLDVS_HOME:-/root/.pelldvs}
  export ETH_RPC_URL=${ETH_RPC_URL:-http://eth:8545}
  export ETH_WS_URL=${ETH_WS_URL:-ws://eth:8545}
  export SERVICE_CHAIN_RPC_URL=${SERVICE_CHAIN_RPC_URL:-http://eth:8545}
  export ADMIN_KEY_FILE="$PELLDVS_HOME/keys/admin.ecdsa.key.json"
  export CHAIN_ID=${CHAIN_ID:-1337}

  export PELLEMULATOR_HOME=${PELLEMULATOR_HOME:-/root/.pell-emulator}
  export PELLEMULATOR_SERVER_PORT=${PELLEMULATOR_SERVER_PORT:-9092}
}


function setup_admin_key {
  ## create admin key
  # echo  -ne '\n\n' | pelldvs keys create admin --key-type=ecdsa --insecure > /tmp/admin.key
  # ADMIN_KEY=$(cat /tmp/admin.key | sed -n 's/.*\/\/[[:space:]]*\([0-9a-f]\{64\}\)[[:space:]]*\/\/.*/\1/p')

  ## For development purposes, we use a predefined admin key from Hardhat's first account
  ## This key is used to deploy contracts in the contract template repo
  ADMIN_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
  echo -ne '\n\n' | pelldvs keys import --key-type ecdsa --insecure admin $ADMIN_KEY --home $PELLDVS_HOME >/dev/null

  export ADMIN_ADDRESS=$(pelldvs keys show admin --home $PELLDVS_HOME | awk '/Key content:/{getline; print}' | jq -r .address)
}

function update_pelldvs_config {
  pelldvs init --home $PELLDVS_HOME
  update-config() {
    KEY="$1"
    VALUE="$2"
    sed -i "s|${KEY} = \".*\"|${KEY} = \"${VALUE}\"|" $PELLDVS_HOME/config/config.toml
  }
}

function create_registry_router {
  ## Create registry router
  REGISTRY_ROUTER_ADDRESS_FILE="/root/RegistryRouterAddress.json"
  REGISTRY_ROUTER_FACTORY_ADDRESS=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellRegistryRouterFactory.json" | jq -r .address)
  pelldvs client dvs create-registry-router \
    --home $PELLDVS_HOME \
    --from admin \
    --rpc-url $ETH_RPC_URL \
    --registry-router-factory $REGISTRY_ROUTER_FACTORY_ADDRESS \
    --initial-owner $ADMIN_ADDRESS \
    --dvs-chain-approver $ADMIN_ADDRESS \
    --churn-approver $ADMIN_ADDRESS \
    --ejector $ADMIN_ADDRESS \
    --pauser $ADMIN_ADDRESS \
    --unpauser $ADMIN_ADDRESS \
    --initial-paused-status false \
    --save-to-file $REGISTRY_ROUTER_ADDRESS_FILE \
    --force-save true

  ## Export registry router address
  export PELL_REGISTRY_ROUTER=$(cat $REGISTRY_ROUTER_ADDRESS_FILE | jq -r .address)
}

function init_pell_emulator {
  ## Initialize emulator and config will be written in $PELLEMULATOR_HOME/config/config.json
  pell-emulator init --home $PELLEMULATOR_HOME
  export EMULATOR_CONFIG_FILE="$PELLEMULATOR_HOME/config/config.json"

#  ## Get contracts addresses from Hardhat
  PELL_DELEGATION_MNAGER=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellDelegationManager-Proxy.json" | jq -r .address)
  PELL_DVS_DIRECTORY=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellDVSDirectory-Proxy.json" | jq -r .address)
  PELL_STRATEGY_MANAGER=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellStrategyManager-Proxy.json" | jq -r .address)
  PELL_REGISTRY_INTERACTOR=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/RegistryInteractor.json" | jq -r .address)
  STAKING_STRATEGY_MANAGER=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/StrategyManager-Proxy.json" | jq -r .address)
  STAKING_DELEGATION_MANAGER=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/DelegationManager-Proxy.json" | jq -r .address)
  SERVICE_OMNI_OPERATOR_SHARES_MANAGER=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/OmniOperatorSharesManager-Proxy.json" | jq -r .address)

  DVS_CENTRAL_SCHEDULER=$(ssh hardhat "cat $HARDHAT_DVS_PATH/CentralScheduler-Proxy.json" | jq -r .address)
  DVS_OPERATOR_STAKE_MANAGER=$(ssh hardhat "cat $HARDHAT_DVS_PATH/OperatorStakeManager-Proxy.json" | jq -r .address)

  ## Update emulator contracts addresses
  update-emulator-config() {
    JQ_EXPR="$1"
    jq "$JQ_EXPR" "$EMULATOR_CONFIG_FILE" >/tmp/tmp.json &&
      mv /tmp/tmp.json "$EMULATOR_CONFIG_FILE"
  }

  update-emulator-config '.contract_address.PellDelegationManager = "'$PELL_DELEGATION_MNAGER'"'
  update-emulator-config '.contract_address.PellDVSDirectory = "'$PELL_DVS_DIRECTORY'"'
  update-emulator-config '.contract_address.PellStrategyManager = "'$PELL_STRATEGY_MANAGER'"'
  update-emulator-config '.contract_address.PellRegistryRouter = "'$PELL_REGISTRY_ROUTER'"'
  update-emulator-config '.contract_address.StakingStrategyManager = "'$STAKING_STRATEGY_MANAGER'"'
  update-emulator-config '.contract_address.StakingDelegationManager = "'$STAKING_DELEGATION_MANAGER'"'
  update-emulator-config '.contract_address.ServiceOmniOperatorSharesManager = "'$SERVICE_OMNI_OPERATOR_SHARES_MANAGER'"'
  update-emulator-config '.contract_address.PellRegistryInteractor = "'$PELL_REGISTRY_INTERACTOR'"'

  update-emulator-config '.contract_address.DVSCentralScheduler = "'$DVS_CENTRAL_SCHEDULER'"'
  update-emulator-config '.contract_address.DVSOperatorStakeManager = "'$DVS_OPERATOR_STAKE_MANAGER'"'

  update-emulator-config ".port = $PELLEMULATOR_SERVER_PORT"
  update-emulator-config '.rpc_url = "'$ETH_RPC_URL'"'
  update-emulator-config '.ws_url = "'$ETH_WS_URL'"'
  update-emulator-config '.auto_update_connector = true'
  update-emulator-config '.deployer_key_file = "'$ADMIN_KEY_FILE'"'

  cat "$EMULATOR_CONFIG_FILE" | jq

}

function start_pell_emulator {
  ## start emulator
  pell-emulator start --home "$PELLEMULATOR_HOME"
}

# start sshd
/usr/sbin/sshd &

logt "Load Default Values for ENV Vars if not set."
load_defaults

logt "Wait for Hardhat to be ready"
hardhat_healthcheck

logt "Update PellDVS Config"
update_pelldvs_config

if [ ! -f /root/emulator_initialized ]; then
  logt "Setup Admin Key"
  setup_admin_key

  tail -n 30 $PELLDVS_HOME/config/config.toml

  logt "Create Registry Router"
  create_registry_router

  cat $REGISTRY_ROUTER_ADDRESS_FILE

  logt "Initialize Pell Emulator"
  init_pell_emulator

  touch /root/emulator_initialized
else
  logt "Pell Emulator already initialized, skipping..."
fi

logt "Start Pell Emulator"
start_pell_emulator
