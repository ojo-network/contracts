BINARY=secretcli
CHAINID_1="pulsar-2"
CONTRACT_PATH=./cosmwasm/artifacts/std_reference.wasm
RPC="https://rpc.testnet.secretsaturn.net"
NODE="--node $RPC"
TXFLAG="$NODE --chain-id $CHAINID_1 --gas-prices 1uscrt --gas-adjustment 1.3"

export DEMOWALLET=$($BINARY keys show relayer -a --keyring-backend=test --keyring-dir=./relayer-data/ ) && echo $DEMOWALLET;
$BINARY query wasm list-code $NODE

# deploy smart contract
$BINARY tx compute store $CONTRACT_PATH --from $DEMOWALLET --keyring-backend=test --keyring-dir=./cw-relayer/data/ $TXFLAG -y
sleep 5

$BINARY tx compute instantiate 19480 '{}' --label=newtest --from $DEMOWALLET --keyring-backend=test --keyring-dir=./cw-relayer/data/ -y
sleep 5

#instantiate contract
$BINARY tx compute instantiate 1 '{}' --label=newthings --from $DEMOWALLET --keyring-backend=test --keyring-dir=./cw-relayer/newrelayerdata/ --sign-mode=direct $TXFLAG -y
sleep 5

## query contract address
CONTRACT=$($BINARY query compute list-contract-by-code "19480" $NODE --output json | jq -r '.[0].contract_address')
echo $CONTRACT

#sample
ADD_RELAYERS='{"add_relayers": {"relayers": ["secret19uywplnd25gzxgc3t8pyqsl2tcse8heag2s6av"]}}'
$BINARY tx compute execute $CONTRACT "$ADD_RELAYERS" --keyring-dir=./relayer-data --keyring-backend=test --from $DEMOWALLET --dry-run -y
sleep 5

RELAY='{"force_relay": {"symbol_rates": [["ATOM","14"]], "resolve_time":"10", "request_id":"2"}}'
$BINARY tx compute execute $CONTRACT "$RELAY" --keyring-backend=test --keyring-dir=./cw-relayer/data/ --sign-mode=direct --from $DEMOWALLET $TXFLAG --offline
sleep 5

QUERY='{"get_ref": {"symbol": "ATOM"}}'
$BINARY query compute query $CONTRACT "$QUERY" $NODE --output json