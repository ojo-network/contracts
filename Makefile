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

.PHONY: test-e2e compile-contract-arm compile-contract