BINARY=wasmd
CHAINID_1="wasm-test"

RELAY_MNEMONIC_1="alley afraid soup fall idea toss can goose become valve initial strong forward bright dish figure check leopard decide warfare hub unusual join cart"
RPC="http://0.0.0.0:26657"

NODE="--node $RPC"
TXFLAG="$NODE --chain-id $CHAINID_1 --gas-prices 0.25stake --keyring-backend test --broadcast-mode block --gas auto --gas-adjustment 1.3 -y"

# network check
export DEMOWALLET=$($BINARY keys show rly1 -a --keyring-backend test --home ./data/$CHAINID_1) && echo $DEMOWALLET;

## query contract address
CONTRACT=$($BINARY query wasm list-contract-by-code "1" $NODE --output json | jq -r '.contracts[-1]')
echo $CONTRACT

QUERY_CONTRACT=$($BINARY query wasm list-contract-by-code "2" $NODE --output json | jq -r '.contracts[-1]')
echo $QUERY_CONTRACT

REQUEST_RELAY='{"request": {"symbol": "ATOM", "resolve_time": "0","callback_sig":"","callback_data":"somename"}}'
for i in {1..50}
do
    $BINARY tx wasm execute $QUERY_CONTRACT "$REQUEST_RELAY" --from $DEMOWALLET --home ./data/$CHAINID_1 $TXFLAG
    $BINARY q wasm contract-state smart $QUERY_CONTRACT '"get_price"' --home ./data/$CHAINID_1
done
#
#CALLBACK='{"callback":{"symbol": "ATOM", "symbol_rate": "1000000", "resolve_time": "1625267345000", "callback_data": "test"}}'
#$BINARY tx wasm execute $QUERY_CONTRACT "$CALLBACK" --from $DEMOWALLET --home ./data/$CHAINID_1 $TXFLAG
#
#QUERY='{"get_total_request":{}}'
#$BINARY query wasm contract-state smart $CONTRACT "$QUERY" $NODE --output json

