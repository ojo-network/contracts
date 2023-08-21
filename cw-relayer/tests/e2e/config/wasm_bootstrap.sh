#!/bin/sh

CHAIN_DIR=/data
HOME=$CHAIN_DIR/$E2E_WASMD_CHAIN_ID

wasmd init val --chain-id=$E2E_WASMD_CHAIN_ID --home $HOME
echo $E2E_WASMD_VAL_MNEMONIC | wasmd keys add val --recover --keyring-backend=test --home $HOME

sed -i 's/"time_iota_ms": "1000"/"time_iota_ms": "10"/'$HOME/config/genesis.json

wasmd genesis add-genesis-account $(wasmd keys show val -a --keyring-backend=test --home $HOME) 1000000000000stake --home $HOME
wasmd genesis gentx val 500000000000stake --chain-id=$E2E_WASMD_CHAIN_ID --keyring-backend=test --home $HOME
wasmd genesis collect-gentxs --home $HOME


#start wasm chain
wasmd start --rpc.laddr tcp://0.0.0.0:26657 --grpc.address="0.0.0.0:8080" --home $HOME