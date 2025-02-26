#!/usr/bin/env bash

docker compose up -d

sleep 120

URL="http://127.0.0.1:26657/request_dvs?data=%22111111111111=12345678901234567890123456789017%22&height=111&chainid=1337" 
response=$(curl -sS -H "Accept: application/json" -X GET "$URL")

expectedStr="{\"jsonrpc\":\"2.0\",\"id\":-1,\"result\":{\"code\":0,\"data\":\"\",\"log\":\"\",\"codespace\":\"\"}}"

if echo "$response" | grep -q "$expectedStr"; then
	echo "$response"
	echo "test task: ok"
else
	exit 1
fi
 
docker compose down 