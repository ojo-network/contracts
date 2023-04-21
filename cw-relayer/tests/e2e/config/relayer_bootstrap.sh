#!/bin/sh

#start relayer
echo -ne '\n' | cw-relayer ./config/relayer-config.toml > ./relayer-test.log 2>&1 &