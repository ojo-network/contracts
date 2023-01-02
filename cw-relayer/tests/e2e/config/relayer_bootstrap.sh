#!/bin/sh

sed -i "s|\$RELAYER_ADDRESS|$E2E_WASMD_VAL_ADDRESSS|g" config/relayer-config.toml
sed -i "s|\$CHAIN_ID|$E2E_WASMD_CHAIN_ID|g" config/relayer-config.toml
sed -i "s|\$CONTRACT_ADDRESS|$1|g" config/relayer-config.toml

#start relayer
echo -ne '\n' | cw-relayer config/relayer-config.toml > ./relayer-test.log 2>&1 &
