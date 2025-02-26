
function load_defaults {
  export HARDHAT_CONTRACTS_PATH="/app/pell-middleware-contracts/lib/pell-contracts/deployments/localhost"
  export HARDHAT_DVS_PATH="/app/pell-middleware-contracts/deployments/localhost"

  export PELLDVS_HOME=${PELLDVS_HOME:-/root/.pelldvs}
  export ETH_RPC_URL=${ETH_RPC_URL:-http://eth:8545}
  export ETH_WS_URL=${ETH_WS_URL:-ws://eth:8545}
}

function setup_root_key {
  ## For development purposes, we use a predefined root key to faucet funds to the receiver
  ## This key is the second account from Hardhat's test network
  ROOT_KEY=0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d
  echo -ne '\n\n' | pelldvs keys import --key-type ecdsa --insecure root $ROOT_KEY --home $PELLDVS_HOME >/dev/null

  export ROOT_ADDRESS=$(pelldvs keys show root --home $PELLDVS_HOME | awk '/Key content:/{getline; print}' | jq -r .address)
}


function faucet {
  RECEIVER_ADDRESS="$1"
  AMOUNT=$(printf "%0.f" "${2:-1e18}")

  ## By default, cast will use $ETH_RPC_URL environment variable as the RPC URL
  ROOT_BALANCE=$(cast balance "$ROOT_ADDRESS" --rpc-url "$ETH_RPC_URL")
  echo "Root balance: $ROOT_BALANCE"

  ## If cast send throws an error like "duplicate field",
  ## please update the version of forge of eth container to the latest version
  cast send "$RECEIVER_ADDRESS" --value "$AMOUNT" --private-key "$ROOT_KEY" --rpc-url "$ETH_RPC_URL"
  RECEIVER_BALANCE=$(cast balance "$RECEIVER_ADDRESS" --rpc-url "$ETH_RPC_URL")
  echo "Receiver balance: $RECEIVER_BALANCE"
}

function show_operator_registered {
  local ADDRESS=$1
  local STAKING_DELEGATION_MANAGER=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/DelegationManager-Proxy.json" | jq -r .address)
  local PELL_DELEGATION_MNAGER=$(ssh hardhat "cat $HARDHAT_CONTRACTS_PATH/PellDelegationManager-Proxy.json" | jq -r .address)
  local IS_STAKING_OPERATOR=$(cast call $STAKING_DELEGATION_MANAGER "isOperator(address)" $ADDRESS)
  local IS_PELL_OPERATOR=$(cast call $PELL_DELEGATION_MNAGER "isOperator(address)" $ADDRESS)
  echo "Is staking operator: $IS_STAKING_OPERATOR"
  echo "Is pell operator: $IS_PELL_OPERATOR"
}

function show_dvs_operator_info {
  local OPERATOR_ADDRESS=$1
  local DVS_REGISTRY_COORDINATOR=$(ssh hardhat "cat $HARDHAT_DVS_PATH/CentralScheduler-Proxy.json" | jq -r .address)

  ## Get operator info -> (operator_id, status),
  ## status: 0 -> NEVER, 1 -> REGISTERED, 2 -> DEREGISTERED
  cast call "$DVS_REGISTRY_COORDINATOR" "getOperator(address)((bytes32,uint8))" $OPERATOR_ADDRESS
}

function get_operator_list {
  local GROUP_NUMBER=${1:-0}
  local BLOCK_NUMBER=${2:-$(cast block-number)}
  OPERATOR_INDEX_MANAGER_ADDRESS=$(ssh hardhat "cat $HARDHAT_DVS_PATH/OperatorIndexManager-Proxy.json" | jq -r .address)
  cast call "$OPERATOR_INDEX_MANAGER_ADDRESS" "getOperatorListAtBlockNumber(uint8,uint32)(bytes32[])" $GROUP_NUMBER $BLOCK_NUMBER
  cast call "$OPERATOR_INDEX_MANAGER_ADDRESS" "totalOperatorsForGroup(uint8)(uint32)" $GROUP_NUMBER
}

load_defaults
## If root key is not imported, import it
if ! pelldvs keys show root --home "$PELLDVS_HOME" >/dev/null 2>&1; then
  setup_root_key
fi
