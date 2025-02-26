#!/bin/bash

set -x
set -e

logt() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') $1"
}

function load_defaults {
  export HARDHAT_CONTRACTS_PATH="/app/pell-middleware-contracts/lib/pell-contracts/deployments/localhost"
  export HARDHAT_CONTRACTS_PATH="/app/pell-middleware-contracts/lib/pell-contracts/deployments/localhost"
  export HARDHAT_DVS_PATH="/app/pell-middleware-contracts/deployments/localhost"

  export PELLDVS_HOME=${PELLDVS_HOME:-/root/.pelldvs}
  export ETH_RPC_URL=${ETH_RPC_URL:-http://eth:8545}
  export SERVICE_CHAIN_RPC_URL=${SERVICE_CHAIN_RPC_URL:-http://eth:8545}
  export ETH_WS_URL=${ETH_WS_URL:-ws://eth:8545}

  export AGGREGATOR_INDEXER_START_HEIGHT=${AGGREGATOR_INDEXER_START_HEIGHT:-0}
  export AGGREGATOR_INDEXER_BATCH_SIZE=${AGGREGATOR_INDEXER_BATCH_SIZE:-1000}
  export CHAIN_ID=${CHAIN_ID:-1337}

  export AGGREGATOR_RPC_LADDR=${AGGREGATOR_RPC_LADDR:-0.0.0.0:26653}
	export DEBUG_PORT=${DEBUG_PORT:-2345}
}

function init_aggregator {

  update-config() {
    KEY="$1"
    VALUE="$2"
    sed -i "s|${KEY} = \".*\"|${KEY} = \"${VALUE}\"|" $PELLDVS_HOME/config/config.toml
  }

  mkdir -p $PELLDVS_HOME/config
  REGISTRY_ROUTER_ADDRESS=$(ssh emulator "cat /root/RegistryRouterAddress.json" | jq -r .address)
  cat <<EOF > $PELLDVS_HOME/config/aggregator.json
{
    "aggregator_rpc_server": "$AGGREGATOR_RPC_LADDR",
    "operator_response_timeout": "10s"
}
EOF
}

function start_aggregator {
  if [ "$DEBUG_ENABLED" = "true" ]; then
    dlv exec /usr/bin/pelldvs \
      --listen=:$DEBUG_PORT --headless=true --api-version=2 --accept-multiclient\
      --log --log-output=debugger \
      -- start-aggregator --home $PELLDVS_HOME
  else
		pelldvs start-aggregator --home $PELLDVS_HOME
  fi
}

logt "Load Default Values for ENV Vars if not set."
load_defaults

if [ ! -f /root/aggregator_initialized ]; then
  logt "Init aggregator"
  init_aggregator
  touch /root/aggregator_initialized
fi

logt "Starting aggregator..."
start_aggregator
