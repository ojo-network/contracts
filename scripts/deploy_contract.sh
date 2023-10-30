BINARY=injectived
CHAINID_1="injective-1"
CONTRACT_PATH=./cosmwasm/artifacts/std_reference-aarch64.wasm
DEMO_MNEMONIC_1="pony glide frown crisp unfold lawn cup loan trial govern usual matrix theory wash fresh address pioneer between meadow visa buffalo keep gallery swear"
RPC="http://0.0.0.0:26657"

NODE="--node $RPC"
TXFLAG="$NODE --chain-id $CHAINID_1 --gas-prices 0.25stake --keyring-backend test --gas auto --gas-adjustment 1.3 --broadcast-mode=async -y"
# network check
export DEMOWALLET=$($BINARY keys show demowallet1 -a --keyring-backend test --home ./data/$CHAINID_1) && echo $DEMOWALLET;
export RELAYWALLET=$($BINARY keys show rly1 -a --keyring-backend test --home ./data/$CHAINID_1) && echo $RELAYWALLET;
#$BINARY query wasm list-code $NODE

# deploy smart contract
$BINARY tx wasm store $CONTRACT_PATH --from $DEMOWALLET --home ./data/$CHAINID_1 $TXFLAG
sleep 2

#instantiate contract
$BINARY tx wasm instantiate 1 '{}' --label test --admin $DEMOWALLET --from $DEMOWALLET --home ./data/$CHAINID_1 $TXFLAG
sleep 2

# query contract address
CONTRACT=$($BINARY query wasm list-contract-by-code "1" $NODE --output json | jq -r '.contracts[-1]')
echo $CONTRACT

#sample transactions
ADD_RELAYERS='{"add_relayers": {"relayers": ["inj1p4szayprev5yml5f2l6uq39gyuzamq3juetup5"]}}'
$BINARY tx wasm execute $CONTRACT "$ADD_RELAYERS" --home ./data/$CHAINID_1 --from $DEMOWALLET $TXFLAG
sleep 2

RELAY='{"force_relay": {"symbol_rates": [["stake","30"]], "resolve_time":"10", "request_id":"1"}}'
$BINARY tx wasm execute $CONTRACT "$RELAY" --home ./data/$CHAINID_1 --from $RELAYWALLET $TXFLAG
sleep 2

QUERY='{"get_ref": {"symbol": "stake"}}'
$BINARY query wasm contract-state smart $CONTRACT "$QUERY" $NODE --output json
