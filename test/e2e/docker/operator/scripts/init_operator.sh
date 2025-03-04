logt() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') $1"
}

function load_defaults {
  export OPERATOR_KEY_NAME=${OPERATOR_KEY_NAME:-operator}
  export AGGREGATOR_RPC_URL=${AGGREGATOR_RPC_URL:-aggregator:26653}

  export AGGREGATOR_INDEXER_START_HEIGHT=${AGGREGATOR_INDEXER_START_HEIGHT:-0}
  export AGGREGATOR_INDEXER_BATCH_SIZE=${AGGREGATOR_INDEXER_BATCH_SIZE:-1000}

  export PELLDVS_HOME=${PELLDVS_HOME:-/root/.pelldvs}
  export ETH_RPC_URL=${ETH_RPC_URL:-http://eth:8545}
  export ETH_WS_URL=${ETH_WS_URL:-ws://eth:8545}
  export CHAIN_ID=${CHAIN_ID:-1337}

  export SERVICE_CHAIN_RPC_URL=${SERVICE_CHAIN_RPC_URL:-http://eth:8545}
  export SERVICE_CHAIN_WS_URL=${SERVICE_CHAIN_WS_URL:-ws://eth:8545}

  source "$(dirname "$0")/utils.sh"
}

## TODO: move operator config to seperated location
function init_pelldvs_config {
  pelldvs init --home $PELLDVS_HOME
  update-config() {
    KEY="$1"
    VALUE="$2"
    sed -i "s|${KEY} = \".*\"|${KEY} = \"${VALUE}\"|" $PELLDVS_HOME/config/config.toml
  }

  ## update config
  REGISTRY_ROUTER_FACTORY_ADDRESS=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellRegistryRouterFactory.json" | jq -r .address)
  PELL_DELEGATION_MNAGER=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellDelegationManager-Proxy.json" | jq -r .address)
  PELL_DVS_DIRECTORY=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellDVSDirectory-Proxy.json" | jq -r .address)
  REGISTRY_ROUTER_ADDRESS=$(ssh emulator "cat /root/RegistryRouterAddress.json" | jq -r .address)

  update-config aggregator_rpc_url "$AGGREGATOR_RPC_URL"

  ## FIXME: don't use absolute path for key
  update-config operator_bls_private_key_store_path "$PELLDVS_HOME/keys/$OPERATOR_KEY_NAME.bls.key.json"
  update-config operator_ecdsa_private_key_store_path "$PELLDVS_HOME/keys/$OPERATOR_KEY_NAME.ecdsa.key.json"

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

function setup_operator_key {
  if pelldvs keys show $OPERATOR_KEY_NAME --home "$PELLDVS_HOME" >/dev/null 2>&1; then
    echo "Operator key already exists, skipping import"
    return
  fi

  ## Create operator key
  echo  -ne '\n\n' | pelldvs keys create $OPERATOR_KEY_NAME --key-type=ecdsa --insecure > /tmp/operator.key
  export OPERATOR_KEY=$(cat /tmp/operator.key | sed -n 's/.*\/\/[[:space:]]*\([0-9a-f]\{64\}\)[[:space:]]*\/\/.*/\1/p')
  export OPERATOR_ADDRESS=$(pelldvs keys show $OPERATOR_KEY_NAME --home $PELLDVS_HOME | awk '/Key content:/{getline; print}' | head -n 1 | jq -r .address)

  ## To register operator in the DVS, we need the operator's BLS key with the same name
  echo  -ne '\n\n' | pelldvs keys create $OPERATOR_KEY_NAME --key-type=bls --insecure > /tmp/operator_bls.key

	set +e
  ## Faucet operator with 1 PELL
  faucet $OPERATOR_ADDRESS 1e18
  set -e
}

function set_operator_address() {
    export OPERATOR_ADDRESS=$(cat $PELLDVS_HOME/keys/$OPERATOR_KEY_NAME.ecdsa.key.json | jq -r '.address')
    echo "OPERATOR_ADDRESS=$OPERATOR_ADDRESS"
}

function register_operator_on_pell {
  OPERATOR_METADATA_URI=https://raw.githubusercontent.com/matthew7251/Metadata/main/Matthew_Metadata.json
  PELL_DELEGATION_MNAGER=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellDelegationManager-Proxy.json" | jq -r .address)
  pelldvs client operator register-operator \
    --home $PELLDVS_HOME \
    --from $OPERATOR_KEY_NAME \
    --rpc-url $ETH_RPC_URL \
    --delegation-manager $PELL_DELEGATION_MNAGER \
    --metadata-uri $OPERATOR_METADATA_URI

  show_operator_registered "$OPERATOR_ADDRESS"
}

function register_operator_to_dvs {
  REGISTRY_ROUTER_ADDRESS=$(ssh emulator "cat /root/RegistryRouterAddress.json" | jq -r .address)
  pelldvs client operator register-operator-to-dvs \
    --home $PELLDVS_HOME \
    --from $OPERATOR_KEY_NAME \
    --rpc-url $ETH_RPC_URL \
    --registry-router $REGISTRY_ROUTER_ADDRESS \
    --groups 0 \
    --socket http://operator:26657
  show_dvs_operator_info $OPERATOR_ADDRESS
}

logt "Load Default Values for ENV Vars if not set."
load_defaults

logt "Init Pelldvs Config"
init_pelldvs_config

logt "Setup Operator Key"
setup_operator_key

logt "Set Operator Address"
set_operator_address

logt "Register Operator"
register_operator_on_pell

logt "Register Operator to DVS"
register_operator_to_dvs
