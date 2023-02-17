package main

func reInitChain() string {
	return `
#!/bin/bash -xeu
BINARY=wasmd
CHAIN_DIR=./data
CHAINID_1="wasm-test"
VAL_MNEMONIC_1="copper push brief egg scan entry inform record adjust fossil boss egg comic alien upon aspect dry avoid interest fury window hint race symptom"
DEMO_MNEMONIC_1="pony glide frown crisp unfold lawn cup loan trial govern usual matrix theory wash fresh address pioneer between meadow visa buffalo keep gallery swear"
RELAY_MNEMONIC_1="alley afraid soup fall idea toss can goose become valve initial strong forward bright dish figure check leopard decide warfare hub unusual join cart"

# Stop if it is already running
if pgrep -x "$BINARY" >/dev/null; then
    echo "Terminating $BINARY..."
    killall $BINARY
fi

echo "Removing previous data..."
rm -rf $CHAIN_DIR/$CHAINID_1 &> /dev/null

# Add directories for both chains, exit if an error occurs
if ! mkdir -p $CHAIN_DIR/$CHAINID_1 2>/dev/null; then
    echo "Failed to create chain folder. Aborting..."
    exit 1
fi

echo "Initializing $CHAINID_1..."
$BINARY init test --home $CHAIN_DIR/$CHAINID_1 --chain-id=$CHAINID_1

echo "Adding genesis accounts..."
echo $VAL_MNEMONIC_1 | $BINARY keys add val1 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test
echo $DEMO_MNEMONIC_1 | $BINARY keys add demowallet1 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test
echo $RELAY_MNEMONIC_1 | $BINARY keys add rly1 --home $CHAIN_DIR/$CHAINID_1 --recover --keyring-backend=test

$BINARY add-genesis-account $($BINARY --home $CHAIN_DIR/$CHAINID_1 keys show val1 --keyring-backend test -a) 100000000000000000000000000stake --home $CHAIN_DIR/$CHAINID_1
$BINARY add-genesis-account $($BINARY --home $CHAIN_DIR/$CHAINID_1 keys show demowallet1 --keyring-backend test -a) 100000000000000000000000000stake  --home $CHAIN_DIR/$CHAINID_1
$BINARY add-genesis-account $($BINARY --home $CHAIN_DIR/$CHAINID_1 keys show rly1 --keyring-backend test -a) 100000000000000000000000000stake  --home $CHAIN_DIR/$CHAINID_1

sed -i -e 's/enable = false/enable = true/g' $CHAIN_DIR/$CHAINID_1/config/app.toml
sed -i -e 's/pruning = "default"/pruning = "everything"/g' $CHAIN_DIR/$CHAINID_1/config/app.toml

echo "Creating and collecting gentx..."
$BINARY gentx val1 1000000000000000000000stake  --home $CHAIN_DIR/$CHAINID_1 --chain-id $CHAINID_1 --keyring-backend test

$BINARY collect-gentxs --home $CHAIN_DIR/$CHAINID_1`
}

func deployContract() string {
	return `
#!/bin/bash -xeu
BINARY=wasmd
CHAINID_1="wasm-test"
CONTRACT_PATH=./std_reference.wasm
DEMO_MNEMONIC_1="pony glide frown crisp unfold lawn cup loan trial govern usual matrix theory wash fresh address pioneer between meadow visa buffalo keep gallery swear"
RPC="http://0.0.0.0:26657"

NODE="--node $RPC"
TXFLAG="$NODE --chain-id $CHAINID_1 --gas-prices 0.25stake --keyring-backend test --gas auto --gas-adjustment 1.3"
# network check
export DEMOWALLET=$($BINARY keys show demowallet1 -a --keyring-backend test --home ./data/$CHAINID_1) && echo $DEMOWALLET;
#$BINARY query wasm list-code $NODE

# deploy smart contract
$BINARY tx wasm store $CONTRACT_PATH --from $DEMOWALLET --home ./data/$CHAINID_1 $TXFLAG -y
sleep 5

#instantiate contract
$BINARY tx wasm instantiate 1 '{}' --label test --admin $DEMOWALLET --from $DEMOWALLET --home ./data/$CHAINID_1 $TXFLAG -y
sleep 5

# query contract address
CONTRACT=$($BINARY query wasm list-contract-by-code "1" $NODE --output json | jq -r '.contracts[-1]')
echo $CONTRACT

#sample transactions
ADD_RELAYERS='{"add_relayers": {"relayers": ["wasm1usr9g5a4s2qrwl63sdjtrs2qd4a7huh6qksawp"]}}'
$BINARY tx wasm execute $CONTRACT "$ADD_RELAYERS" --home ./data/$CHAINID_1 --from $DEMOWALLET $TXFLAG -y
sleep 5

RELAY='{"force_relay": {"symbol_rates": [["stake","30"]], "resolve_time":"10", "request_id":"1"}}'
$BINARY tx wasm execute $CONTRACT "$RELAY" --home ./data/$CHAINID_1 --from $DEMOWALLET $TXFLAG -y
sleep 5

QUERY='{"get_ref": {"symbol": "stake"}}'
$BINARY query wasm contract-state smart $CONTRACT "$QUERY" $NODE --output json
`
}
