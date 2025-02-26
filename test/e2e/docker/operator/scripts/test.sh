
set -x

function load_defaults {
  export HARDHAT_CONTRACTS_PATH="/app/pell-middleware-contracts/lib/pell-contracts/deployments/localhost"
  export HARDHAT_DVS_PATH="/app/pell-middleware-contracts/deployments/localhost"

  export PELLDVS_HOME=${PELLDVS_HOME:-/root/.pelldvs}
  export ETH_RPC_URL=${ETH_RPC_URL:-http://eth:8545}
  export ETH_WS_URL=${ETH_WS_URL:-ws://eth:8545}
  export OPERATOR_RPC_SERVER=${OPERATOR_RPC_SERVER:-operator:26657}
}

function operator_healthcheck {
  set +e
  while true; do
    ssh operator "test -f /root/operator_initialized"
    if [ $? -eq 0 ]; then
      echo "✅ operator_healthcheck_1: Operator initialized, proceeding to the next step..."
      break
    fi
    echo "⏳ operator_healthcheck_1:Operator not initialized, retrying in 2 second..."
    sleep 2
  done
  ## Wait for operator to be ready
  sleep 3
  set -e
}

function operator_healthcheck2 {
  set +e
  while true; do
    curl -s $OPERATOR_RPC_SERVER >/dev/null
    if [ $? -eq 0 ]; then
      echo "✅ operator_healthcheck_2: RPC port is ready, proceeding to the next step..."
      break
    fi
    echo "⏳ operator_healthcheck_2: RPC port not ready, retrying in 2 seconds..."
    sleep 2
  done
  sleep 3
  set -e
}

function assert_eq {
  if [ "$1" != "$2" ]; then
    echo "[FAIL] Expected $1 to be equal to $2"
    exit 1
  fi
  echo "[PASS] Expected $1 to be equal to $2"
}

load_defaults
echo "load_defaults: ok"
echo -e "\n\n"

echo "OPERATOR_RPC_SERVER: $OPERATOR_RPC_SERVER"

echo -e "operator_healthcheck\n\n"
operator_healthcheck
echo -e "\n\n"

echo "operator_healthcheck2"
operator_healthcheck2
echo -e "\n\n"

syncURL="http://operator:26657/request_dvs?data=%22111111111112=12345678901234567890123456789017%22&height=111&chainid=1337"
syncResponse=$(curl -sS -H "Accept: application/json" -X GET "$syncURL")

syncExpectedStr="{\"jsonrpc\":\"2.0\",\"id\":-1,\"result\":{\"code\":0,\"data\":\"\",\"log\":\"\",\"codespace\":\"\"}}"

if echo "$syncResponse" | grep -q "$syncExpectedStr"; then
	echo "$syncResponse"
	echo "test sync rpc task: ok"
else
	exit 1
fi


asyncURL="http://operator:26657/request_dvs_async?data=%22111111111111=57945678901234567890123456789017%22&height=111&chainid=1337"
asyncResponse=$(curl -sS -H "Accept: application/json" -X GET "$asyncURL")

asyncExpectedStr="{\"jsonrpc\":\"2.0\",\"id\":-1,\"result\":{\"hash\":\"C7B7DD51C31DA8E27D28856A5DA8FE964393E56B47696F28DA7B3CA7AAD9BE51\"}}"

if echo "$asyncResponse" | grep -q "$asyncExpectedStr"; then
	echo "$asyncResponse"
	echo "test async rpc task: ok"
else
	exit 1
fi


sleep 10
searchByEventURL="http://operator:26657/search_request?query=\"SecondEventType.SecondEventKey='SecondEventValue'\""
searchByEventResponse=$(curl -sS -H "Accept: application/json" -X GET "$searchByEventURL")

searchByEventExpectedStr='{"jsonrpc":"2.0","id":-1,"result":{"dvs_requests":[{"dvs_request":{"data":"MTExMTExMTExMTEyPTEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE3","height":"111","chain_id":"1337"},"dvs_response":{"data":"MTExMTExMTExMTEy","hash":"MTY2YmY0N2JmZmY4OWNiYjI3MDRiNmRiNjkxNWU2MTZhMDJhYWQ2MTVkNDY2NWM4OWRhYjQ3NWE4MThhYTgyNA==","signers_apk_g2":"CMofsne/bI+/aAeOs1AR+0Fg8ck1goOUxepOuO5V/esYJHPC1AW/EOanarHgtDb+4dgyzgOvRyHuMIulBVlvVgM759+yjXBaZj6S5J+byP1+TrbAaUMTtc/2GNWuhYu9Bt4k/e2h1r56sodVH+2YdEYoiIdVFUsUwgpoBoESY6k=","signers_agg_sig_g1":"ETCrLaP65UjewDTfj0WNinKV0ZmacHul3vQajwBRwFcJBfoA3Ybf8CpTarp9BXONE8xa0dyzVp9TIG6+/i3BAQ=="},"response_dvs_request":{"events":[{"type":"FirstEventType","attributes":[{"key":"First Event Key","value":"First Event Value","index":true}]},{"type":"SecondEventType","attributes":[{"key":"First Event Key","value":"First Event Value","index":true},{"key":"SecondEventKey","value":"SecondEventValue","index":true}]},{"type":"ThirdEventType","attributes":[{"key":"First Event Key","value":"First Event Value","index":true},{"key":"SecondEventKey","value":"SecondEventValue","index":true},{"key":"ThirdEventKey","value":"Third Event Value","index":true}]}],"response":"MTExMTExMTExMTEy","response_digest":"MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTc="},"response_dvs_response":{"events":[{"type":"FourthEventType","attributes":[{"key":"FourthEventKey","value":"Fourth Event Value","index":true}]}]},"hash":"E23A1CA929CE8A02251A258FF3E5EF070A6D44364500A6003577E96E0FDA4CE5"},{"dvs_request":{"data":"MTExMTExMTExMTExPTU3OTQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE3","height":"111","chain_id":"1337"},"dvs_response":{"data":"MTExMTExMTExMTEx","hash":"OWVmMTg1ZDlmODc2NmU0MTU5ZmY4ZGNkZGUyNzUxMmIzZWYzMmU0ZTMyNDkxZmZhZWM2MTc5OTgwZjAyNGY3Mg==","signers_apk_g2":"CMofsne/bI+/aAeOs1AR+0Fg8ck1goOUxepOuO5V/esYJHPC1AW/EOanarHgtDb+4dgyzgOvRyHuMIulBVlvVgM759+yjXBaZj6S5J+byP1+TrbAaUMTtc/2GNWuhYu9Bt4k/e2h1r56sodVH+2YdEYoiIdVFUsUwgpoBoESY6k=","signers_agg_sig_g1":"ABUEC5hM7YKIg8NOtVGPAELYnN1VYUfMoZfpIRaCRckDNPg0snvJgIOsIiIAgAydAgBNFRCGEhmEU9SVOZTK8Q=="},"response_dvs_request":{"events":[{"type":"FirstEventType","attributes":[{"key":"First Event Key","value":"First Event Value","index":true}]},{"type":"SecondEventType","attributes":[{"key":"First Event Key","value":"First Event Value","index":true},{"key":"SecondEventKey","value":"SecondEventValue","index":true}]},{"type":"ThirdEventType","attributes":[{"key":"First Event Key","value":"First Event Value","index":true},{"key":"SecondEventKey","value":"SecondEventValue","index":true},{"key":"ThirdEventKey","value":"Third Event Value","index":true}]}],"response":"MTExMTExMTExMTEx","response_digest":"NTc5NDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTc="},"response_dvs_response":{"events":[{"type":"FourthEventType","attributes":[{"key":"FourthEventKey","value":"Fourth Event Value","index":true}]}]},"hash":"C7B7DD51C31DA8E27D28856A5DA8FE964393E56B47696F28DA7B3CA7AAD9BE51"}],"total_count":"2"}}'
normalized_json1=$(echo "$searchByEventResponse" | jq 'del(.result.dvs_requests[].dvs_response) | .result.dvs_requests |= sort_by(.hash)')
normalized_json2=$(echo "$searchByEventExpectedStr" | jq 'del(.result.dvs_requests[].dvs_response) | .result.dvs_requests |= sort_by(.hash)')

if [ "$normalized_json1" == "$normalized_json2" ]; then
	echo "$searchByEventResponse"
	echo "test search by event rpc task: ok"
else
	exit 1
fi

