#!/bin/sh

BINARY=wasmd
CONTRACT_PATH=config/std_reference.wasm
RPC="http://0.0.0.0:26657"
HOME=/data/$E2E_WASMD_CHAIN_ID


NODE="--node $RPC"
TXFLAG="$NODE --chain-id $E2E_WASMD_CHAIN_ID --gas-prices 0.25stake --keyring-backend test --gas auto --gas-adjustment 1.3"

export wallet=$(wasmd keys show val -a --keyring-backend  test --home $HOME) && echo $wallet;

# deploy smart contract
wasmd tx wasm store $CONTRACT_PATH --from $wallet --home $HOME $TXFLAG -y
sleep 5

#instantiate contract
wasmd tx wasm instantiate 1 '{}' --label test --admin $wallet --from $wallet --home $HOME $TXFLAG -y
sleep 5
