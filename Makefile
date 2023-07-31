install:
	@echo "installing wasmd"
	./scripts/install.sh

init: kill-dev
	@echo "init wasm chain"
	./scripts/init.sh

localnet:
	@echo "Starting up test network"
	./scripts/init.sh
	./scripts/start.sh

deploy:
	@echo "Deploying ref contract"
	./scripts/deploy_contract.sh

trigger-event:
	@echo "Deploying ref contract"
	./scripts/trigger_event.sh

restart:
	${MAKE} kill-dev
	${MAKE} localnet
	sleep 5
	${MAKE} deploy


kill-dev:
	@echo "Killing wasmd and removing previous data"
	-@rm -rf ./data
	-@killall -9 wasmd 2>/dev/null

kill:
	@echo "Killing wasmd"
	-@killall -9 wasmd 2>/dev/null

test-unit-relayer:
	@echo "Testing Relayer"
	cd cw-relayer && ${MAKE} test-unit

test-unit-contract:
	@echo "Testing contract"
	cd cosmwasm && cargo test

compile-contract:
	cosmwasm/scripts/build_artifacts.sh

compile-contract-arm:
	cosmwasm/scripts/build_artifacts_arm.sh

build-relayer:
	cd cw-relayer && ${MAKE} build

lint-relayer:
	cd cw-relayer &&  golangci-lint run ./...

format-contract:
	cd cosmwasm && cargo fmt

start-relayer:
	cd cw-relayer && ${MAKE} start

test-e2e:
	@echo "Running e2e tests"
	${MAKE} compile-contract
	cp -f cosmwasm/artifacts/ojo_price_feeds.wasm cw-relayer/tests/e2e/config/ojo_price_feeds.wasm
	cp -f cosmwasm/artifacts/price_query.wasm cw-relayer/tests/e2e/config/price_query.wasm
	cd cw-relayer && ${MAKE} test-e2e
	rm cw-relayer/tests/e2e/config/*.wasm

test-e2e-arm:
	@echo "Running e2e tests"
	${MAKE} compile-contract-arm
	cp -f cosmwasm/artifacts/std_reference-aarch64.wasm cw-relayer/tests/e2e/config/std_reference.wasm
	cd cw-relayer && ${MAKE} test-e2e
	rm cw-relayer/tests/e2e/config/std_reference.wasm

.PHONY: test-e2e test-e2e-arm compile-contract compile-contract-arm