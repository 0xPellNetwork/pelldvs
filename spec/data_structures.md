
# Data Structures Spec

Here we describe the data structures used in the PellDVS and the rules for validating them.

- [Data Structures](#data-structures)
  - [Data types used by RPC interfaces](#data-types-used-by-rpc-interfaces)
    - [RequestDVS](#requestdvs)
    - [ResultRequest](#resultrequest)
  - [Data types used by DVS Reactor](#data-types-used-by-dvs-reactor)
    - [OnRequest](#onrequest)
    - [DVSRequest](#dvsrequest)
    - [DVSRequestResult](#dvsrequestresult)
    - [SaveDVSRequestResult](#savedvsrequestresult)
  - [Data types used by AVSI](#data-types-used-by-avsi)
    - [ProcessDVSRequest](#processdvsrequest)
    - [ResponseProcessDVSRequest](#responseprocessdvsrequest)
    - [ProcessDVSResponse](#processdvsresponse)
    - [RequestProcessDVSResponse](#requestprocessdvsresponse)
    - [DVSResponse](#dvsresponse)
    - [ResponseProcessDVSResponse](#responseprocessdvsresponse)
  - [Data types used by Aggregator](#data-types-used-by-aggregator)
    - [CollectResponseSignature](#collectresponsesignature)
    - [ValidatedResponse](#validatedresponse)


## Data types used by RPC interfaces

### RequestDVS

RequestDVS is the outermost entry point. Users initiate requests for task processing by calling the RPC of RequestDVS.

- Imported parameter types:

```
func (env *Environment) RequestDVS(ctx *rpctypes.Context,
    data []byte,
    height int64,
    chainID int64,
    groupNumbers []uint32,
    groupThresholdPercentages []uint32,
) (*ctypes.ResultRequest, error) 
```

- Return value type:

```
type ResultRequest struct {
    Code      uint32         `json:"code"`
    Data      bytes.HexBytes `json:"data"`
    Log       string         `json:"log"`
    Codespace string         `json:"codespace"`
}
```

## Data types used by DVS Reactor

### OnRequest

OnRequest is the core process of DVS Reactor and PellDVS. The key steps are executed in OnRequest.

- Imported parameter types:

```
message DVSRequest {
  bytes                 data                         = 1;
  int64                 height                       = 2;
  int64                 chain_id                     = 3;
  repeated uint32       group_numbers                = 4;
  repeated uint32       group_threshold_percentages  = 5;
}
```

- Return value type:

```
message DVSRequestResult {
  DVSRequest                  dvs_request                   = 1;
  ResponseProcessDVSRequest   response_process_dvs_request  = 2;
  DVSResponse                 dvs_response                  = 3;
  ResponseProcessDVSResponse  response_process_dvs_response = 4;
}
```

### DVSRequestResult

DVSRequestResult is a data structure that records each DVSRequest to the local database.

- Type definition (consistent with the imported parameter types of OnRequest)

```
message DVSRequestResult {
  DVSRequest                  dvs_request                   = 1;
  ResponseProcessDVSRequest   response_process_dvs_request  = 2;
  DVSResponse                 dvs_response                  = 3;
  ResponseProcessDVSResponse  response_process_dvs_response = 4;
}
```

- Execute function

SaveDVSRequestResult is responsible for storing incoming DVSRequestResult to the local database.

```
func (dvs *DVSReactor) SaveDVSRequestResult(res *avsitypes.DVSRequestResult, first bool) error
```

SaveDVSRequestResult function calls DVS Request Indexer for processing, and the function imported parameter type of DVS Request Indexer is also DVSRequestResult:

```
func (dvsReqIdx *DvsRequestIndex) Index(result *avsi.DVSRequestResult) error
```

## Data types used by AVSI

### ProcessDVSRequest

ProcessDVSRequest is used by Application to handle DVSRequest requests.

- Imported parameter types:

DVSRequest is an imported parameter of the OnRequest function.

```
message RequestProcessDVSRequest {
  DVSRequest        request = 1;    // Parameter of the OnRequest function
  repeated Operator operator = 2;
}

message Operator {
  bytes id          = 1;  // [32]byte
  bytes address     = 2;  // [20]byte
  string meta_uri   = 3;
  string socket     = 4;
  int64 stake       = 5;
  OperatorPubkeys pubkeys = 6;
}
```

- Return value type:

```
message ResponseProcessDVSRequest {
  uint32          code            = 1;
  bytes           data            = 2;
  string          log             = 3;  // nondeterministic
  string          info            = 4;  // nondeterministic
  repeated Event  events          = 5
      [(gogoproto.nullable) = false, (gogoproto.jsontag) = "events,omitempty"];  // nondeterministic
  string          codespace       = 6;
  bytes           response        = 7;
  bytes           response_digest = 8;
}
```

### ProcessDVSResponse

ProcessDVSRequest is used for the final result after the application processes the aggregated signature.

- Imported parameter types:

DVSRequest is an imported parameter of the OnRequest function.

```
message RequestProcessDVSResponse {
  DVSRequest                 dvs_request =1;   // Parameter of the OnRequest function
  DVSResponse                dvs_response =2;
}

message DVSResponse {
  bytes                          data                             = 1;
  string                         error                            = 2;
  bytes                          hash                             = 3;
  repeated bytes                 non_signers_pubkeys_g1           = 4;
  repeated bytes                 group_apks_g1                    = 5;
  bytes                          signers_apk_g2                   = 6;
  bytes                          signers_agg_sig_g1               = 7;
  repeated uint32                non_signer_group_bitmap_indices  = 8;
  repeated uint32                group_apk_indices                = 9;
  repeated uint32                total_stake_indices              = 10;
  repeated NonSignerStakeIndice  non_signer_stake_indices         = 11;
}
```

- Return value type:

```
message ResponseProcessDVSResponse {
  uint32          code            = 1;
  bytes           data            = 2;
  string          log             = 3;  // nondeterministic
  string          info            = 4;  // nondeterministic
  repeated Event  events          = 5
      [(gogoproto.nullable) = false, (gogoproto.jsontag) = "events,omitempty"];  // nondeterministic
  string          codespace       = 6;
}
```

## Data types used by Aggregator

### CollectResponseSignature

CollectResponseSignature is the core function of Aggregator, which is used to collect the task results and signature data of Operator, and generate the aggregated signature data according to the signature of Operator.

- Imported parameter types:

RequestData is an imported parameter of the OnRequest function.

```
type ResponseWithSignature struct {
    Data        []byte
    Digest      [32]byte
    Signature   *bls.Signature
    OperatorID  [32]byte
    RequestData avsiTypes.DVSRequest    // Parameter of the OnRequest function
}
```

- Return value type:

```
type ValidatedResponse struct {
    Data                        []byte
    Err                         error
    Hash                        []byte
    NonSignersPubkeysG1         []*bls.G1Point
    GroupApksG1                 []*bls.G1Point
    SignersApkG2                *bls.G2Point
    SignersAggSigG1             *bls.Signature
    NonSignerGroupBitmapIndices []uint32
    GroupApkIndices             []uint32
    TotalStakeIndices           []uint32
    NonSignerStakeIndices       [][]uint32
}
```




