BINARY=wasmd
RPC="http://0.0.0.0:26657"
NODE="--node $RPC"
CHAINID_1="wasm-test"

QUERY_CONTRACT=$($BINARY query wasm list-contract-by-code "2" $NODE --output json | jq -r '.contracts[-1]')
echo $QUERY_CONTRACT


$BINARY q wasm contract-state smart $QUERY_CONTRACT '{"get_price":{"symbol": "ATOM"}}' --home ./data/$CHAINID_1

$BINARY q wasm contract-state smart $QUERY_CONTRACT '{"get_median":{"symbol": "ATOM"}}' --home ./data/$CHAINID_1

$BINARY q wasm contract-state smart $QUERY_CONTRACT '{"get_deviation":{"symbol": "ATOM"}}' --home ./data/$CHAINID_1

