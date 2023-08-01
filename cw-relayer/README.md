# Ojo Cosmwasm Relayer

This repo contains the cosmwasm relayer used to take prices from Ojo & Post them to an existing CW on-chain contract.

### Relayer
- Relayer subscribes to event rpc to listen to ojo network for new blocks
- When new prices are stored in ojo abci end blocker and ojo.oracle.v1.EventSetFxRate event is triggered, the relayer then queries the prices of from query_rpc and creates a wasm based msg to relay these prices (rates, medians) and price deviation of assets  to the contract
- Relay triggers 3 msgs (2 if median is disabled in the config) 
  - ```MsgRelay``` to relay rates (spot) of the assets
  - ```MsgRelayHistoricalMedian``` to relay historical medians of the assets
  - ```MsgRelayHistoricalDeviation``` to relay prices deviations of the assets

each msg above has a forced version which ignores the resolve duration present in the contract. It is used when msgs is failed to be broadcasted and number of missed attempts exceeds missed threshold  

#### Median Duration
- median duration determines how frequently median prices are posted to the contract
- if median duration is set to 0, then median prices are not posted to the contract

### Links to other supported implementations
- [Secret Network](https://github.com/ojo-network/contracts/tree/secret)
- [Evm](https://github.com/ojo-network/contracts/tree/evm)