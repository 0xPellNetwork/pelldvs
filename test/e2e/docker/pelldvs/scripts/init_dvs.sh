#! /bin/bash

set -e

logt() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') $1"
}

function load_defaults {
  export HARDHAT_CONTRACTS_PATH="/app/pell-middleware-contracts/lib/pell-contracts/deployments/localhost"
  export HARDHAT_CONTRACTS_PATH="/app/pell-middleware-contracts/lib/pell-contracts/deployments/localhost"
  export HARDHAT_DVS_PATH="/app/pell-middleware-contracts/deployments/localhost"

	export AGGREGATOR_INDEXER_START_HEIGHT=${AGGREGATOR_INDEXER_START_HEIGHT:-0}
	export AGGREGATOR_INDEXER_BATCH_SIZE=${AGGREGATOR_INDEXER_BATCH_SIZE:-1000}

	export SERVICE_CHAIN_RPC_URL=${SERVICE_CHAIN_RPC_URL:-http://eth:8545}
	export SERVICE_CHAIN_WS_URL=${SERVICE_CHAIN_WS_URL:-ws://eth:8545}
	export CHAIN_ID=${CHAIN_ID:-1337}

  export PELLDVS_HOME=${PELLDVS_HOME:-/root/.pelldvs}
  export ETH_RPC_URL=${ETH_RPC_URL:-http://eth:8545}
  export ETH_WS_URL=${ETH_WS_URL:-ws://eth:8545}
}

function update_pelldvs_config {
  pelldvs init --home "$PELLDVS_HOME"

  ## Update config
  REGISTRY_ROUTER_FACTORY_ADDRESS=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellRegistryRouterFactory.json" | jq -r .address)
  REGISTRY_ROUTER_ADDRESS=$(ssh emulator "cat /root/RegistryRouterAddress.json" | jq -r .address)
  update-config() {
    KEY="$1"
    VALUE="$2"
    sed -i "s|${KEY} = \".*\"|${KEY} = \"${VALUE}\"|" ~/.pelldvs/config/config.toml
  }
	update-config interactor_config_path "$PELLDVS_HOME/config/interactor_config.json"

  DVS_OPERATOR_KEY_MANAGER=$(ssh hardhat "cat $HARDHAT_DVS_PATH/OperatorKeyManager-Proxy.json" | jq -r .address)
  DVS_CENTRAL_SCHEDULER=$(ssh hardhat "cat $HARDHAT_DVS_PATH/CentralScheduler-Proxy.json" | jq -r .address)
  DVS_OPERATOR_INFO_PROVIDER=$(ssh hardhat "cat $HARDHAT_DVS_PATH/OperatorInfoProvider.json" | jq -r .address)
  DVS_OPERATOR_INDEX_MANAGER=$(ssh hardhat "cat $HARDHAT_DVS_PATH/OperatorIndexManager-Proxy.json" | jq -r .address)

  cat <<EOF > $PELLDVS_HOME/config/interactor_config.json
{
    "rpc_url": "$ETH_RPC_URL",
    "chain_id": $CHAIN_ID,
    "contract_config": {
    "indexer_start_height": $AGGREGATOR_INDEXER_START_HEIGHT,
    "indexer_batch_size": $AGGREGATOR_INDEXER_BATCH_SIZE,
    "pell_registry_router_factory": "$REGISTRY_ROUTER_FACTORY_ADDRESS",
    "pell_dvs_directory": "$PELL_DVS_DIRECTORY",
    "pell_delegation_manager": "$PELL_DELEGATION_MNAGER",
    "pell_registry_router": "$REGISTRY_ROUTER_ADDRESS",
    "dvs_configs": {
      "$CHAIN_ID": {
        "chain_id": $CHAIN_ID,
        "rpc_url": "$SERVICE_CHAIN_RPC_URL",
        "ws_url": "$SERVICE_CHAIN_WS_URL",
        "operator_info_provider": "$DVS_OPERATOR_INFO_PROVIDER",
        "operator_key_manager": "$DVS_OPERATOR_KEY_MANAGER",
        "central_scheduler": "$DVS_CENTRAL_SCHEDULER",
        "operator_index_manager": "$DVS_OPERATOR_INDEX_MANAGER"
      }
    }
  }
}
EOF
cat $PELLDVS_HOME/config/interactor_config.json
}

function setup_admin_key {
  ## create admin key
  # echo  -ne '\n\n' | pelldvs keys create admin --key-type=ecdsa --insecure > /tmp/admin.key
  # ADMIN_KEY=$(cat /tmp/admin.key | sed -n 's/.*\/\/[[:space:]]*\([0-9a-f]\{64\}\)[[:space:]]*\/\/.*/\1/p')

  ## For development purposes, we use a predefined admin key from Hardhat's first account
  ## This key is used to deploy contracts in the contract template repo
  export ADMIN_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
  if ! pelldvs keys show admin --home "$PELLDVS_HOME" >/dev/null 2>&1; then
    echo -ne '\n\n' | pelldvs keys import --key-type ecdsa --insecure admin $ADMIN_KEY --home $PELLDVS_HOME >/dev/null
  fi

  export ADMIN_ADDRESS=$(pelldvs keys show admin --home $PELLDVS_HOME | awk '/Key content:/{getline; print}' | head -n 1 | jq -r .address)
}

function register_chain_to_pell() {
  set -e
  set -x

  REGISTRY_ROUTER_ADDRESS=$(ssh emulator "cat /root/RegistryRouterAddress.json" | jq -r .address)
  DVS_CENTRAL_SCHEDULER=$(ssh hardhat "cat $HARDHAT_DVS_PATH/CentralScheduler-Proxy.json" | jq -r .address)
  pelldvs client dvs register-chain-to-pell \
      --home $PELLDVS_HOME \
      --rpc-url $ETH_RPC_URL \
      --registry-router "$REGISTRY_ROUTER_ADDRESS" \
      --central-scheduler "$DVS_CENTRAL_SCHEDULER" \
      --dvs-rpc-url $ETH_RPC_URL \
      --dvs-from admin \
      --approver-key-name admin

  set +x
}

function show_supported_chain() {
	logt ""
  REGISTRY_ROUTER_ADDRESS=$(ssh emulator "cat /root/RegistryRouterAddress.json" | jq -r .address)
  SUPPORTED_CHAIN_RESULT=$(cast call $REGISTRY_ROUTER_ADDRESS "supportedChainInfos(uint256)(uint256,address,address,address)" 0 \
    --rpc-url $ETH_RPC_URL \
    --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80)

  logt "Supported Chain Info: $SUPPORTED_CHAIN_RESULT"
  logt ""
}

function create_group {
  STBTC_STRATEGY_ADDRESS=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/stBTC-Strategy-Proxy.json" | jq -r .address)
  PBTC_STRATEGY_ADDRESS=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/pBTC-Strategy-Proxy.json" | jq -r .address)
  cat > ./group-0-config.json <<EOF
{
  "minimum_stake": 0,
  "pool_params": [
    {
      "chain_id": 1337,
      "multiplier": 1,
      "pool": "$STBTC_STRATEGY_ADDRESS"
    },
    {
      "chain_id": 1337,
      "multiplier": 1,
      "pool": "$PBTC_STRATEGY_ADDRESS"
    }
  ],
  "operator_set_params": {
    "kick_bi_ps_of_operator_stake": 10,
    "kick_bi_ps_of_total_stake": 10,
    "max_operator_count": 1000
  }
}
EOF

  pelldvs client dvs create-group \
    --home $PELLDVS_HOME \
    --from admin \
    --rpc-url $ETH_RPC_URL \
    --registry-router $REGISTRY_ROUTER_ADDRESS \
    --config ./group-0-config.json
}

function show_group {
  GROUP_COUNT=$(cast call "$REGISTRY_ROUTER_ADDRESS" "groupCount()" --rpc-url "$ETH_RPC_URL")
  logt "Group Count From Registry Router in Pell EVM: $GROUP_COUNT"

  DVS_CENTRAL_SCHEDULER=$(ssh hardhat "cat $HARDHAT_DVS_PATH/CentralScheduler-Proxy.json" | jq -r .address)
  GROUP_COUNT=$(cast call "$DVS_CENTRAL_SCHEDULER" "groupCount()" --rpc-url "$ETH_RPC_URL")
  logt "Group Count From Registry CentralScheduler in Service EVM: $GROUP_COUNT"
}

logt "Load Default Values for ENV Vars if not set."
load_defaults

logt "Update PellDVS Config"
update_pelldvs_config

logt "Setup Admin Key"
setup_admin_key

logt "Register Chain to Pell"
register_chain_to_pell

sleep 2

logt "show supported chain"
show_supported_chain

logt "show group before create"
show_group

logt "Create Group"
sleep 1
create_group

logt "show group after create"
show_group
