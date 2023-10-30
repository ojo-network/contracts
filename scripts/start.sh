#!/bin/bash

BINARY=injectived
CHAIN_DIR=./data
CHAINID_1="injective-1"

echo "Starting $CHAINID_1 in $CHAIN_DIR..."
echo "Creating log file at $CHAIN_DIR/$CHAINID_1.log"
$BINARY start --log-level trace --log-format json --home $CHAIN_DIR/$CHAINID_1 --pruning=everything --minimum-gas-prices=0.00001stake > $CHAIN_DIR/$CHAINID_1.log 2>&1 &
