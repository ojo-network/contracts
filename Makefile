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

start-relayer:
	cd cw-relayer && ${MAKE} start

test-e2e:
	@echo "Running e2e tests"
	${MAKE} compile-contract
	cp -f cosmwasm/artifacts/std_reference.wasm cw-relayer/tests/e2e/config/std_reference.wasm
	cd cw-relayer && ${MAKE} test-e2e
	rm cw-relayer/tests/e2e/config/std_reference.wasm

test-e2e-arm:
	@echo "Running e2e tests"
	${MAKE} compile-contract-arm
	cp -f cosmwasm/artifacts/std_reference-aarch64.wasm cw-relayer/tests/e2e/config/std_reference.wasm
	cd cw-relayer && ${MAKE} test-e2e
	rm cw-relayer/tests/e2e/config/std_reference.wasm

.PHONY: test-e2e test-e2e-arm compile-contract compile-contract-arm