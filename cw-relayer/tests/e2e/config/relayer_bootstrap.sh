#!/bin/sh

sed -i "s|\$CONTRACT_ADDRESS|$EVM_CONTRACT_ADDRESS|g" /usr/local/config.toml
sed -i "s|\$RELAYER_ADDRESS|$EVM_RELAYER_ADDRESS|g" /usr/local/config.toml
sed -i "s|\$PRIV_KEY|$EVM_PRIV_KEY|g" /usr/local/config.toml


#start relayer
echo -ne '\n' | cw-relayer /usr/local/config.toml > ./relayer-test.log 2>&1 &