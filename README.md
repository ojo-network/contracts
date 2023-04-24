# Ojo's Price Feed Contracts

This monorepo is intended to contain Ojo's Price Feeding Contracts. It will be organized by:

- EVM Contracts
- Golang-based Relayer

The relayer will be a golang-based implementation to upload pricing information into the standard evm contract.

- In case of relayer restart modify median request id and request id in config.toml, according to the current request id