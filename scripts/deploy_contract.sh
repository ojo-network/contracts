BINARY=secretcli
CHAINID_1="pulsar-3"
CONTRACT_PATH=./cosmwasm/artifacts/std_reference.wasm
RPC="https://rpc.pulsar3.scrttestnet.com:443"
NODE="--node $RPC"
TXFLAG="$NODE --chain-id $CHAINID_1 --gas-prices 1uscrt --gas-adjustment 1.3 --broadcast-mode block --gas auto"

export DEMOWALLET=$($BINARY keys show relayer2 -a --keyring-backend=test ) && echo $DEMOWALLET;

# deploy smart contract
$BINARY tx compute store $CONTRACT_PATH --from $DEMOWALLET --keyring-backend=test $TXFLAG -y

#instantiate contract
$BINARY tx compute instantiate 1 '{}' --label=test --from $DEMOWALLET --keyring-backend=test -y

# query contract address
CONTRACT=$($BINARY query compute list-contract-by-code "1" $NODE --output json | jq -r '.[0].contract_address')
echo $CONTRACT

#adding relayer
ADD_RELAYERS='{"add_relayers": {"relayers": ["secret19uywplnd25gzxgc3t8pyqsl2tcse8heag2s6av"]}}'
$BINARY tx compute execute $CONTRACT "$ADD_RELAYERS" --from $DEMOWALLET --keyring-backend=test $TXFLAG -y

#sample price tx
RELAY='{"force_relay": {"symbol_rates": [["ATOM","14"]], "resolve_time":"10", "request_id":"2"}}'
$BINARY tx compute execute $CONTRACT "$RELAY" --keyring-backend=test  --sign-mode=direct --from $DEMOWALLET $TXFLAG

QUERY='{"get_ref": {"symbol": "ATOM"}}'
$BINARY query compute query $CONTRACT "$QUERY" $NODE --output json