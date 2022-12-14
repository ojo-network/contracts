#!/bin/bash

BINARY=wasmd
CHAIN_DIR=./data
CHAINID_1="wasm-test"
GRPCPORT_1=8090
GRPCWEB_1=8091

echo "Starting $CHAINID_1 in $CHAIN_DIR..."
echo "Creating log file at $CHAIN_DIR/$CHAINID_1.log"
$BINARY start --log_level trace --log_format json --home $CHAIN_DIR/$CHAINID_1 --pruning=nothing --grpc.address="0.0.0.0:$GRPCPORT_1" --grpc-web.address="0.0.0.0:$GRPCWEB_1" --minimum-gas-prices=0.00001stake > $CHAIN_DIR/$CHAINID_1.log 2>&1 &
