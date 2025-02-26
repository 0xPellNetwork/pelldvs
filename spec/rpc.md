# RPC Spec

This document defines the JSON-RPC interface specification for PellDVS.

- [RPC Spec](#rpc-spec)
  - [Health](#health)
    - [Node heartbeat](#node-heartbeat)
  - [RequestDVS](#requestdvs)
    - [Parameters](#parameters)
    - [Request](#request)
    - [Response](#response)
  - [RequestDVSAsync](#requestdvsasync)
    - [Parameters](#parameters-1)
    - [Request](#request-1)
    - [Response](#response-1)
  - [QueryRequest](#queryrequest)
    - [Parameters](#parameters-2)
    - [Request](#request-2)
    - [Response](#response-2)
  - [SearchRequest](#searchrequest)
    - [Parameters](#parameters-3)
    - [Request](#request-3)
    - [Response](#response-3)


## Health

### Node heartbeat

#### Parameters
None

#### Request

**HTTP**
```
curl http://127.0.0.1:26657/health
```

**JSON-RPC**
```
curl -X POST http://localhost:26657 \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "health"
  }'
```

#### Response
```
{
  "jsonrpc": "2.0",
  "id": -1,
  "result": {}
}
```

---

## RequestDVS

When a new task is detected from the contract, the application layer will initiate a `RequestDVS` call to inform PellDVS that there are new tasks to be processed.

### Parameters

- **data** ([]byte) : The request data encoded
- **height** (int64) : The block height required for security
- **chainID** (int64) : The chain ID
- **groupNumbers** ([]uint32) : The encoded group numbers
- **groupThresholdPercentages** ([]uint32) : The group threshold

### Request

**HTTP**
```
curl \
  -H 'Accept: application/json' \
  -X GET \
  'http://localhost:26657/request_dvs?data=%22111111111112=12345678901234567890123456789017%22&height=111&chainid=1337'
```

**JSON-RPC**
```
curl -X POST http://localhost:26657 -d '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "request_dvs",
  "params": {
    "data": "11111111111112=12345678901234567890123456789017",
    "height": 111,
    "chainid": 1337
  }
}'
```

#### Response
```
{
  "jsonrpc": "2.0",
  "id": -1,
  "result": {
    "code": 0,
    "data": "",
    "log": "",
    "codespace": ""
  }
}
```

---

## RequestDVSAsync

The asynchronous interface of `RequestDVS` will return the hash value of the task instead of the execution result, and subsequent queries need to be based on the hash value.

### Parameters

- **data** ([]byte) : The request data encoded
- **height** (int64) : The block height required for security
- **chainID** (int64) : The chain ID
- **groupNumbers** ([]uint32) : The encoded group numbers
- **groupThresholdPercentages** ([]uint32) : The group threshold

### Request

**HTTP**
```
curl \
  -H 'Accept: application/json' \
  -X GET \
  'http://localhost:26657/request_dvs_async?data=%22111111111112=12345678901234567890123456789017%22&height=111&chainid=1337'
```

**JSON-RPC**
```
curl -X POST http://localhost:26657 -d '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "request_dvs",
  "params": {
    "data": "11111111111112=12345678901234567890123456789017",
    "height": 111,
    "chainid": 1337
  }
}'
```

#### Response
```
{
  "jsonrpc": "2.0",
  "id": -1,
  "result": {
    "hash": "C7B7DD51C31DA8E27D28856A5DA8FE964393E56B47696F28DA7B3CA7AAD9BE51"
  }
}
```

---

## QueryRequest

Query DVS Request information based on the hash value.

### Parameters

- **hash** (string) : `RequestDVSAsync` parameter hash value

### Request

**HTTP**
```
curl \
  -H 'Accept: application/json' \
  -X GET \
  'http://localhost:26657/query_request?hash=C7B7DD51C31DA8E27D28856A5DA8FE964393E56B47696F28DA7B3CA7AAD9BE51'
```

**JSON-RPC**
```
curl -X POST http://localhost:26657 -d '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "query_request",
  "params": {
    "hash": "C7B7DD51C31DA8E27D28856A5DA8FE964393E56B47696F28DA7B3CA7AAD9BE51"
  }
}'
```

#### Response
```
{
  "jsonrpc": "2.0",
  "id": -1,
  "result": { }
}
```

---

## SearchRequest

Query task results based on the event content of the task.

### Parameters

- **query** (string) : A query string, key-value form
- **prove** (bool) : Not useful for now
- **pagePtr** (*int) : Paging parameter, return which page
- **perPagePtr** ([]*int) : Paging parameters, how many items per page

### Request

**HTTP**
```
curl \
  -H 'Accept: application/json' \
  -X GET \
  'http://operator:26657/search_request?query="SecondEventType.SecondEventKey='\''SecondEventValue'\''"'
```

**JSON-RPC**
```
curl -X POST http://localhost:26657 -H 'Content-Type: application/json' -d '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "search_request",
  "params": {
    "query": "SecondEventType.SecondEventKey=\'SecondEventValue\'"
  }
}'
```

#### Response
```
{
  "jsonrpc": "2.0",
  "id": -1,
  "result": {
    "dvs_requests": [
      {
        "dvs_request": {
          "data": "MTExMTExMTExMTEyPTEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE3",
          "height": "111",
          "chain_id": "1337"
        },
        "dvs_response": {
          "data": "MTExMTExMTExMTEy",
          "hash": "MTY2YmY0N2JmZmY4OWNiYjI3MDRiNmRiNjkxNWU2MTZhMDJhYWQ2MTVkNDY2NWM4OWRhYjQ3NWE4MThhYTgyNA=="
        }
      }
    ],
    "total_count": "1"
  }
}
```
