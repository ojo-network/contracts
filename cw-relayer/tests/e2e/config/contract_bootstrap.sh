#!/bin/sh

BINARY=wasmd
CONTRACT_PATH=config/std_reference.wasm
QUERY_CONTRACT_PATH=config/price_query.wasm
RPC="http://0.0.0.0:26657"
HOME=/data/$E2E_WASMD_CHAIN_ID


NODE="--node $RPC"
TXFLAG="$NODE --chain-id $E2E_WASMD_CHAIN_ID --gas-prices 0.25stake -b=block --keyring-backend test --gas auto -y --gas-adjustment 1.3"

export wallet=$(wasmd keys show val -a --keyring-backend test --home $HOME) && echo $wallet;

# deploy smart contract
wasmd tx wasm store $CONTRACT_PATH --from $wallet --home $HOME $TXFLAG
sleep 5

wasmd tx wasm store $QUERY_CONTRACT_PATH --from $wallet --home $HOME $TXFLAG
sleep 5

#instantiate contract
wasmd tx wasm instantiate 1 '{"ping_threshold":"10800"}'  --label test --admin $wallet --from $wallet --home $HOME $TXFLAG
sleep 5

# get contract address for oracle
CONTRACT=$($BINARY query wasm list-contract-by-code "1" $NODE --output json | jq -r '.contracts[-1]')
echo $CONTRACT

# deploy price query sample contract
wasmd tx wasm instantiate 2  '{"contract_address":"'$CONTRACT'"}'  --label test --admin $wallet --from $wallet --home $HOME $TXFLAG
sleep 5

# start requesting prices
QUERY_CONTRACT=$($BINARY query wasm list-contract-by-code "2" $NODE --output json | jq -r '.contracts[-1]')
echo $QUERY_CONTRACT


REQUEST_RELAY='{"request": {"symbol": "TEST-0", "resolve_time": "0","callback_sig":"callback","callback_data":"callback_data_test"}}'
for i in {1..100}
do
    $BINARY tx wasm execute $QUERY_CONTRACT "$REQUEST_RELAY" --from user --home $HOME $TXFLAG
done


