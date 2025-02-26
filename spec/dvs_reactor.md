# DVS Reactor Spec

## Data Structure

DVSReactor is a data structure that includes key components such as ProxyApp, Aggregator Client, and dvsRequestIndexer.

```
type DVSReactor struct {
    config            config.PellConfig
    ProxyApp          proxy.AppConns
    dvsState          *DVSState
    logger            log.Logger
    aggregator        aggTypes.Aggregator
    dvsRequestIndexer requestindex.DvsRequestIndexer
    dvsReader         reader.DVSReader
}
```

## OnRequest

The key function of DVSReactor is `OnRequest`. After the user requests the `request_dvs` interface through RPC, they will enter the processing flow of the `OnRequest` function.

The specific processing of the `OnRequest` function is:

1. **SaveDVSRequestResult (First time)**: Add a `DVSRequestResult` structure in the `DVSRequestIndexer` and save it to the local KV database. At this time, the `DVSRequestResult` data must be empty.

2. **GetOperatorsDVSStateAtBlock**: Call the `GetOperatorsDVSStateAtBlock` function of `DVSReader` to query the Operator ID and the staking weight information of the Operator ID under the corresponding ChainID and Group Number at the block height where the request was initiated.

```
// GetOperatorsDVSStateAtBlock return data type 
type OperatorDVSState struct {
    OperatorID      OperatorID
    OperatorAddress common.Address
    OperatorInfo    OperatorInfo
    StakePerGroup   map[GroupNumber]StakeAmount
    BlockNumber     uint32
}
```

3. **ProcessDVSRequest**: Call the `ProcessDVSRequest` function of the Application that implements the AVSI interface as a parameter, and the specific processing logic will be customized by the Application according to the task content.

```
// Initiate a call to Application's ProcessDVSRequest 
responseProcessDVSRequest, err := dvs.ProxyApp.Dvs().ProcessDVSRequest(context.Background(), &avsitypes.RequestProcessDVSRequest{
    Request:  &request,
    Operator: operators,
})
```

4. **SaveDVSRequestResult (Second time)**: After receiving the result of `ProcessDVSRequest`, call `DVSRequestIndexer` again to save the result data. At this time, the `DVSRequestResult` data is not empty.

5. **SignMessage**: The Operator calls the `SignMessage` function of the `privValidator` module to sign the result of `ProcessDVSRequest`.

```
// Sign the result of ProcessDVSRequest
dvs.SignMessage(responseProcessDVSRequest.ResponseDigest)
```

6. **CollectResponseSignature**: Call the `CollectResponseSignature` function through the Aggregator Client and send your signature along with the `DVSRequestResult` data to the Aggregator server level for processing. The Aggregator server level will collect `DVSRequestResultWithSignature` from different Operators until the number of collected signatures exceeds the threshold, and then return a response to the `CollectResponseSignature` request.

```
// Send DVSRequestResult along with signature to Aggregator for processing 
dvs.aggregator.CollectResponseSignature(&responseWithSignature, validatedResponseCh)
```

7. **SaveDVSRequestResult (Third time)**: Compared to the second call, this time `ProcessDVSRequest` contains data related to aggregate signatures, including a list of Operators who have not signed.

8. **ProcessDVSResponse**: After aggregating the signature, call the `ProcessDVSResponse` function of the Application through the AVSI interface, and hand over the final result to the application layer for processing.

```
// Process the result after Aggregator's aggregated signature
dvs.ProxyApp.Dvs().ProcessDVSResponse(context.Background(), postReq)
```

9. **SaveDVSRequestResult (Fourth time)**: Compared to the third call, this time's `ProcessDVSRequest` contains the response returned by `ProcessDVSResponse`.

---

## State Change of DVSRequestResult Data

`DVSRequestResult` is a structure data that will be stored in the local database. In the processing flow of `OnRequest`, a total of four `SaveDVSRequestResult` function calls were made.

Each time the `SaveDVSRequestResult` function is called, new fields are written, and different fields reflect the different stages of `OnRequest` function execution:

1. **DVSRequest**: The imported parameter of the `OnRequest` function. This field is written to the database when `SaveDVSRequestResult` is called for the first time.

2. **ResponseProcessDVSRequest**: After calling `ProcessDVSRequest` through the AVSI interface, the result of `ResponseProcessDVSRequest` is obtained, and this field is entered into the database when `SaveDVSRequestResult` is called for the second time.

3. **DVSResponse**: After calling the `CollectResponseSignature` call of Aggregator, get the result of `DVSResponse`, and enter this field into the database when calling `SaveDVSRequestResult` for the third time.

4. **ResponseProcessDVSResponse**: After calling `ProcessDVSResponse` through the AVSI interface, the result of `ResponseProcessDVSResponse` is obtained, and this field is entered into the database when `SaveDVSRequestResult` is called for the fourth time.

```
message DVSRequestResult {
  DVSRequest                  dvs_request                   = 1;
  ResponseProcessDVSRequest   response_process_dvs_request  = 2;
  DVSResponse                 dvs_response                  = 3;
  ResponseProcessDVSResponse  response_process_dvs_response = 4;
}
```
