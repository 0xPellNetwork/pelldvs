
# PellDVS Spec

This is the PellDVS Specification.

If you find discrepancies between the spec and the code that
do not have an associated issue or pull request on github,
please submit them to our [bug bounty](https://github.com/0xPellNetwork/pelldvs#security)!

## Contents

PellDVS is a framework used in the Restaking ecosystem to provide interaction between security layer and application layer nodes. (The current code repository) includes the Operator node itself, the Aggregator node, and the interactor interaction command line.

### Data Structures

- [Data Structures Spec](./data_structures.md)

### RPC

- [RPC SPEC](./rpc.md): Specification of the PellDVS remote procedure call interface.

### Software

Software
- [AVSI Spec](./avsi.md): Details about the interaction between the security layer and the application layer through the AVSI interface.
- [DVS Reactor Spec](./dvs_reactor.md): Details on how DVS Reactor handles requests, interacts with Aggreators, and returns responses.
- [Aggregator Spec](./aggregator.md): The Aggerator server level aggregates the task results with node signatures.