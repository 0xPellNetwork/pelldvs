#!/usr/bin/env bash
set -e
set -x

logt() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') $1"
}

function load_defaults {
  export HARDHAT_DVS_PATH="/app/pell-middleware-contracts/deployments/localhost"

  export PELLDVS_HOME=${PELLDVS_HOME:-/root/.pelldvs}
  export ETH_RPC_URL=${ETH_RPC_URL:-http://eth:8545}
  export ETH_WS_URL=${ETH_WS_URL:-ws://eth:8545}
  export OPERATOR_KEY_NAME_LIST=${OPERATOR_KEY_NAME_LIST:-operator01}
}

function operator_healthcheck {
  local container_name=$1
  local timeout=120  # 2 minutes timeout
  local elapsed=0

  while true; do
    ssh $container_name "test -f /root/operator_initialized"
    if [ $? -eq 0 ]; then
      echo "✅️ Operator initialized, proceeding to the next step..."
      break
    fi

    echo "⌛️ Operator not initialized, retrying in 2 second..."
    sleep 2
    elapsed=$((elapsed + 2))

    if [ $elapsed -ge $timeout ]; then
      echo "❌ Timeout reached! Exiting script."
      exit 1
    fi
  done

  ## Wait for operator to be ready
  sleep 3
}

function assert_eq {
  if [ "$1" != "$2" ]; then
    echo "[FAIL] Expected $1 to be equal to $2"
    exit 1
  fi
  echo "[PASS] Expected $1 to be equal to $2"
}

function pelle2etest() {
    export SERVICE_MANAGER_ADDRESS=$(ssh hardhat "cat $HARDHAT_DVS_PATH/MockDVSServiceManager-Proxy.json" | jq -r .address)

    pelle2e check-aggr-sigs \
      --node-url "http://operator:26657" \
      --rpc-url $ETH_RPC_URL \
      --service-manager $SERVICE_MANAGER_ADDRESS
}

load_defaults

logt "Waiting for operator to be ready"
operator_healthcheck operator

logt "Running pelle2e test"
pelle2etest
