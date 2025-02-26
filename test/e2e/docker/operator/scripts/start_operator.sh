#!/bin/bash

set -x
set -e

logt() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') $1"
}

function load_defaults {
  export HARDHAT_CONTRACTS_PATH="/app/pell-middleware-contracts/lib/pell-contracts/deployments/localhost"
  export HARDHAT_DVS_PATH="/app/pell-middleware-contracts/deployments/localhost"
  export TASKGATEWAY_RPC_CLIENT_URL=${TASKGATEWAY_RPC_CLIENT_URL:-task-gateway:8949}
  export PELLDVS_HOME=${PELLDVS_HOME:-/root/.pelldvs}
  export ETH_RPC_URL=${ETH_RPC_URL:-http://eth:8545}
  export ETH_WS_URL=${ETH_WS_URL:-ws://eth:8545}

  export AGGREGATOR_RPC_SERVER=${AGGREGATOR_RPC_SERVER:-aggregator:26653}
	export DEBUG_ENABLED=${DEBUG_ENABLED:-false}
	export DEBUG_PORT=${DEBUG_PORT:-2345}
}

function dvs_healthcheck {
  set +e
  while true; do
    curl -s $AGGREGATOR_RPC_SERVER >/dev/null
    if [ $? -eq 52 ]; then
      echo "DVS RPC port is ready, proceeding to the next step..."
      break
    fi
    echo "DVS RPC port not ready, retrying in 2 seconds..."
    sleep 2
  done
  ## Wait for aggregator to be ready
  sleep 3
  set -e
}

## FIXME: remove this logic after fix. Operator should never use admin key.
function setup_admin_key {
  export ADMIN_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
  if ! pelldvs keys show admin --home "$PELLDVS_HOME" >/dev/null 2>&1; then
    echo -ne '\n\n' | pelldvs keys import --key-type ecdsa --insecure admin $ADMIN_KEY --home $PELLDVS_HOME >/dev/null
  fi

  export ADMIN_ADDRESS=$(pelldvs keys show admin --home $PELLDVS_HOME | awk '/Key content:/{getline; print}' | head -n 1 | jq -r .address)
}


function setup_operator_config {
  setup_admin_key
}

function start_operator {
	if [ "$DEBUG_ENABLED" = "true" ]; then
    dlv exec /usr/bin/pelldvs \
      --listen=:$DEBUG_PORT --headless=true --api-version=2 --accept-multiclient\
      --log --log-output=debugger \
      -- start --proxy_app kvstore --home "$PELLDVS_HOME" --rpc.laddr tcp://0.0.0.0:26657
	else
		pelldvs start --proxy_app kvstore --home "$PELLDVS_HOME" --rpc.laddr tcp://0.0.0.0:26657
	fi
}

## start sshd
/usr/sbin/sshd &

logt "Load Default Values for ENV Vars if not set."
load_defaults

logt "Check if DVS is ready"
dvs_healthcheck

logt "Setup operator config"
setup_operator_config

if [ ! -f /root/operator_initialized ]; then
  logt "Init operator"
  source "$(dirname "$0")/init_operator.sh"
  touch /root/operator_initialized
fi

logt "Starting operator..."
start_operator
