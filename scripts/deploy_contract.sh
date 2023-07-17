BINARY=wasmd
CHAINID_1="wasm-test"

CONTRACT_PATH=./cosmwasm/artifacts/ojo_price_feeds.wasm
QUERY_CONTRACT_PATH=./cosmwasm/artifacts/price_query.wasm

DEMO_MNEMONIC_1="pony glide frown crisp unfold lawn cup loan trial govern usual matrix theory wash fresh address pioneer between meadow visa buffalo keep gallery swear"
RPC="http://0.0.0.0:26657"
NODE="--node $RPC"
TXFLAG="$NODE --chain-id $CHAINID_1 --gas-prices 0.25stake --keyring-backend test --gas auto --gas-adjustment 1.3 --broadcast-mode block -y"

# network check
export DEMOWALLET=$($BINARY keys show demowallet1 -a --keyring-backend test --home ./data/$CHAINID_1) && echo $DEMOWALLET;

# deploy smart contract
$BINARY tx wasm store $CONTRACT_PATH --from $DEMOWALLET --home ./data/$CHAINID_1 $TXFLAG

#instantiate contract
$BINARY tx wasm instantiate 1 '{"ping_threshold":"10800"}' --label test --admin $DEMOWALLET --from $DEMOWALLET --home ./data/$CHAINID_1 $TXFLAG

## query contract address
CONTRACT=$($BINARY query wasm list-contract-by-code "1" $NODE --output json | jq -r '.contracts[-1]')
echo $CONTRACT

#sample transactions
ADD_RELAYERS='{"add_relayers": {"relayers": ["wasm1usr9g5a4s2qrwl63sdjtrs2qd4a7huh6qksawp"]}}'
$BINARY tx wasm execute $CONTRACT "$ADD_RELAYERS" --home ./data/$CHAINID_1 --from $DEMOWALLET $TXFLAG

CHANGE_TRIGGER='{"change_trigger": {"trigger": true}}'
$BINARY tx wasm execute $CONTRACT "$CHANGE_TRIGGER" --home ./data/$CHAINID_1 --from $DEMOWALLET $TXFLAG

# deploy query smart contract
$BINARY tx wasm store $QUERY_CONTRACT_PATH --from $DEMOWALLET --home ./data/$CHAINID_1 $TXFLAG

$BINARY tx wasm instantiate 2 '{"contract_address":"'$CONTRACT'"}' --label query --admin $DEMOWALLET --from $DEMOWALLET --home ./data/$CHAINID_1 $TXFLAG

QUERY_CONTRACT=$($BINARY query wasm list-contract-by-code "2" $NODE --output json | jq -r '.contracts[-1]')
echo $QUERY_CONTRACT