COMETBFT_BUILD_OPTIONS += badgerdb,boltdb,cleveldb,rocksdb

include ../../common.mk

check-env-gh-token:
	@if [ -z "$${GITHUB_TOKEN}" ] ; then \
		echo "Error: GITHUB_TOKEN is not set in environment"; \
		exit 1; \
	else \
		echo "GITHUB_TOKEN is set."; \
	fi

all: generator runner

node:
	go build -race $(BUILD_FLAGS) -tags '$(BUILD_TAGS)' -o build/node ./node

generator:
	go build -o build/generator ./generator

runner:
	go build -o build/runner ./runner

build-pelle2e:
	go build -o build/pelle2e ./pelle2e/cmd/pelle2e

docker-emulator-up:
	@docker compose up emulator -d

docker-emulator-down:
	@docker compose down emulator

docker-emulator-logs:
	@docker compose logs -f emulator

docker-emulator-logs-no-follow:
	@docker compose logs emulator

docker-emulator-shell:
	@docker compose exec -it emulator bash

docker-emulator-rerun:
	@make docker-emulator-down
	@make docker-emulator-up
	@make docker-emulator-logs

docker-aggregator-up:
	@docker compose up aggregator -d

docker-aggregator-down:
	@docker compose down aggregator

docker-aggregator-logs:
	@docker compose logs -f aggregator

docker-aggregator-logs-no-follow:
	@docker compose logs aggregator

docker-aggregator-shell:
	@docker compose exec -it aggregator bash

docker-aggregator-rerun:
	@make docker-aggregator-down
	@make docker-aggregator-up
	@make docker-aggregator-logs

docker-operator-up:
	@docker compose up operator -d

docker-operator-down:
	@docker compose down operator

docker-operator-logs:
	@docker compose logs -f operator

docker-operator-logs-no-follow:
	@docker compose logs operator

docker-operator-shell:
	@docker compose exec -it operator bash

docker-operator-rerun:
	@make docker-operator-down
	@make docker-operator-up
	@make docker-operator-logs

docker-test:
	docker compose run --rm test

docker-test-pelle2e:
	docker compose run --rm test-pelle2e

docker-all-up:
	docker compose up -d

docker-all-down:
	docker compose down -v

docker-up:
	docker compose up -d

docker-down:
	docker compose down -v

docker-up-operator:
	make docker-operator-up

docker-logs-operator:
	make docker-operator-logs

docker-logs-aggregator:
	make docker-aggregator-logs

docker-logs-emulator:
	make docker-emulator-logs

docker-contracts-up:
	docker compose up hardhat -d

docker-contracts-start:
	docker compose start

docker-contracts-down:
	docker compose down

docker-contracts-stop:
	docker compose stop

docker-contracts-logs-eth:
	docker compose logs -f eth

docker-contracts-logs-hardhat:
	docker compose logs -f hardhat

docker-contracts-shell-eth:
	docker compose exec eth sh

docker-contracts-shell-hardhat:
	docker compose exec hardhat bash

docker-build-all: check-env-gh-token
	docker compose -f docker-compose.build.yml build

docker-build-contract: check-env-gh-token
	docker compose -f docker-compose.build.yml build hardhat

docker-build-pelldvs: check-env-gh-token
	docker compose -f docker-compose.build.yml build pelldvs

docker-build-pelle2e: check-env-gh-token
	docker compose -f docker-compose.build.yml build pelle2e

docker-operator-show-key:
	docker compose exec operator bash -c "cat ~/.pelldvs/keys/operator.ecdsa.key.json | jq -r '.address'"

docker-operator-show-interactor-config:
	docker compose exec operator bash -c "cat ~/.pelldvs/config/interactor_config.json | jq -r"

.PHONY: all node generator runner
