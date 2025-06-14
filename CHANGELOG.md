# CHANGELOG

## v0.4.0

### Bug Fixes

- (indexer) [#28](https://github.com/0xPellNetwork/pelldvs/pull/28) bugfix: do not return error while can't get operator from store in the in the loop in OperatorInfoProvider.GetOperatorsDVSStateAtBlock and GetOperatorsStateAtBlock
- (aggregator) [#29](https://github.com/0xPellNetwork/pelldvs/pull/29) refactor: improve aggregator client to handle client shutdown.
- (refactor) [#34](https://github.com/0xPellNetwork/pelldvs/pull/34) refactor: upgrade pelldvs-interactor version and improve e2e
- (fix) [#36](https://github.com/0xPellNetwork/pelldvs/pull/36) bugfix: remove overwrite server config through hardcode in flag defaults.

## v0.3.0

### Features

- (feat) [#10](https://github.com/0xPellNetwork/pelldvs/pull/10) feat(rpc): allow querying by dvs.height
- (feat) [#14](https://github.com/0xPellNetwork/pelldvs/pull/14) refactor: use interactor reader instead of aggregator interactor
- (feat) [#16](https://github.com/0xPellNetwork/pelldvs/pull/16) feat: add `client operator get-weight-for-group` cmd for getting weight for group
- (test) [#17](https://github.com/0xPellNetwork/pelldvs/pull/17) test: improve test for stake and operator weight in e2e

### Improvements

- (CI) [#7](https://github.com/0xPellNetwork/pelldvs/pull/7) ci: add changelog check in CI
- (CI) [#11](https://github.com/0xPellNetwork/pelldvs/pull/11) ci: retrieve E2E port config from GitHub variables
- (refactor) [#9](https://github.com/0xPellNetwork/pelldvs/pull/9) refactor: rename abci to avsi
- (refactor) [#8](https://github.com/0xPellNetwork/pelldvs/pull/8) refactor: update private validator interface
- (fix) [#12](https://github.com/0xPellNetwork/pelldvs/pull/12) fix: task failure when aggregator result does not meet threshold
- (refactor) [#15](https://github.com/0xPellNetwork/pelldvs/pull/15) refactor: make aggregator reactor asynchronously handle requests
- (feat) [#19](https://github.com/0xPellNetwork/pelldvs/pull/19) feat: DVS will save aggregated response with error message
- (refactor) [#24](https://github.com/0xPellNetwork/pelldvs/pull/24) refactor: security and types module

### Bug Fixes

- (dvs) [#6](https://github.com/0xPellNetwork/pelldvs/pull/6) fix: verify DVS request response digest length is 32
- (cmd) [#13](https://github.com/0xPellNetwork/pelldvs/pull/13) fix: gen-validator show correct keypair info

## v0.2.2

v0.2.2 includes a version release CI script fix.

## v0.2.0

This release marks the transition from a previously private repository to the public release. After a period of internal development, the repository is now archived and versioned as v0.2.0.
