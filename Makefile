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
	cd evm && yarn hardhat test

compile-contract:
	cd evm && yarn hardhat compile && yarn hardhat export-abi

lint-contract:
	cd evm && solhint 'contracts/**/*.sol'

update-abi:
	cd evm && npx hardhat export-abi
	cd cw-relayer && rm -r ./relayer/client/oracle.go && abigen --abi /Users/aniketdixit/GolandProjects/contracts/evm/abi/contracts/Oracle.sol/PriceFeed.json --pkg client --type Oracle --out ./relayer/client/oracle.go

start-relayer:
	cd cw-relayer && ${MAKE} start

.PHONY: test-e2e