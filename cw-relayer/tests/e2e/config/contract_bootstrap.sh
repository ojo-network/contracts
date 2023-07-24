#!/bin/sh

BINARY=wasmd
CONTRACT_PATH=config/ojo_price_feeds.wasm
QUERY_CONTRACT_PATH=config/price_query.wasm
RPC="http://0.0.0.0:26657"
HOME=/data/$E2E_WASMD_CHAIN_ID


NODE="--node $RPC"
TXFLAG="$NODE --chain-id $E2E_WASMD_CHAIN_ID --gas-prices 0.25stake -b=block --keyring-backend test --gas auto -y --gas-adjustment 1.3"

export wallet=$(wasmd keys show val -a --keyring-backend test --home $HOME) && echo $wallet;

# deploy smart contract
wasmd tx wasm store $CONTRACT_PATH --from $wallet --home $HOME $TXFLAG

wasmd tx wasm store $QUERY_CONTRACT_PATH --from $wallet --home $HOME $TXFLAG

#instantiate contract
wasmd tx wasm instantiate 1 '{"ping_threshold":"10800"}'  --label test --admin $wallet --from $wallet --home $HOME $TXFLAG

# get contract address for oracle
CONTRACT=$($BINARY query wasm list-contract-by-code "1" $NODE --output json | jq -r '.contracts[-1]')
echo $CONTRACT

# deploy price query sample contract
wasmd tx wasm instantiate 2  '{"contract_address":"'$CONTRACT'"}'  --label test --admin $wallet --from $wallet --home $HOME $TXFLAG

# start requesting prices
QUERY_CONTRACT=$($BINARY query wasm list-contract-by-code "2" $NODE --output json | jq -r '.contracts[-1]')
echo $QUERY_CONTRACT


REQUEST_RELAY='{"request": {"symbol": "ATOM", "resolve_time": "0","callback_sig":"","callback_data":"somename"}}'
for i in {1..100000}
do
    $BINARY tx wasm execute $QUERY_CONTRACT "$REQUEST_RELAY" --from user --home $HOME $TXFLAG > ./query-contract-test.log 2>&1 &
done


