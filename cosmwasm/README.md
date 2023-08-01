# Ojo Networks's Cosmwasm Price Feed Contract

## Overview

This repository contains the CosmWasm code for Ojo Network's standard price data.
This repo is a fork of [Band's Standard Reference Contract](https://github.com/bandprotocol/band-std-reference-contracts-cosmwasm). The main planned changes are:

- Defaulting all asset prices to be denominated in USD.
- Integrated E2E Tests.
- Building a golang-based relayer in the repo.

## Build

### Contract

To compile all contracts, run the following script in the repo root: `/scripts/build_artifacts.sh` or the command below:

```
docker run --rm -v "$(pwd)":/code \
  --mount type=volume,source="$(basename "$(pwd)")_cache",target=/code/target \
  --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
  cosmwasm/workspace-optimizer:0.12.7
```

The optimized wasm code and its checksums can be found in the `/artifacts` directory

### Schema

To generate the JSON schema files for the contract call, queries and query responses, run the following script in the
repo root: `/scripts/build_schemas.sh` or run `cargo schema` in the smart contract directory.

## Usage

To query the prices from Ojo Network, the contract looking to use the price values should
query Ojo Network's `std_reference` contract.

### QueryMsg

The query messages used to retrieve price data for price data are as follows:

```rust
pub enum QueryMsg {
    GetReferenceData {
        // Symbol pair to query where:
        // symbol_pair := (base_symbol, quote_symbol)
        // e.g. BTC/USD ≡ ("BTC", "USD")
        symbol_pair: (String, String),
    },
    GetReferenceDataBulk {
        // Vector of Symbol pair to query
        // e.g. <BTC/USD ETH/USD, OJO/BTC> ≡ <("BTC", "USD"), ("ETH", "USD"), ("OJO", "BTC")>
        symbol_pairs: Vec<(String, String)>,
    },

    // Returns the RefMedianData of a given symbol
    GetMedianRef {
      // Symbol to query
      symbol: String,
    },

    // Returns the RefMedianData of the given symbols
    GetMedianRefDataBulk {
      // Vector of Symbols to query
      symbols: Vec<String>,
    },

    // Returns the deviation RefData of a given symbol
    GetDeviationRef {
      // Symbol to query
      symbol: String,
    },

    // Returns the deviation RefData of the given symbols
    GetDeviationRefBulk {
      // Vector of Symbols to query
      symbols: Vec<String>,
    },
}
```

### ReferenceData

`ReferenceData` is the struct that is returned when querying with `GetReferenceData` or `GetReferenceDataBulk` where the
bulk variant returns `Vec<ReferenceData>`

`ReferenceData` is defined as:

```rust
pub struct ReferenceData {
    // Pair rate e.g. rate of BTC/USD
    pub rate: Uint256,
    // Unix time of when the base asset was last updated. e.g. Last update time of BTC in Unix time
    pub last_updated_base: Uint64,
    // Unix time of when the quote asset was last updated. e.g. Last update time of USD in Unix time
    pub last_updated_quote: Uint64,
}
```

### Examples

#### Single Query

For example, if we wanted to query the price of `BTC/USD`, the demo function below shows how this can be done.

```rust
fn demo(
    std_ref_addr: Addr,
    symbol_pair: (String, String),
) -> StdResult<ReferenceData> {
    deps.querier.query_wasm_smart(
        &std_ref_addr,
        &QueryMsg::GetReferenceData {
            symbol_pair,
        },
    )
}
```

Where the result from `demo(std_ref_addr, ("BTC", "USD"))` would yield:

```
ReferenceData(23131270000000000000000, 1659588229, 1659589497)
```

and the results can be interpreted as:

- BTC/USD
    - `rate = 23131.27 BTC/USD`
    - `lastUpdatedBase = 1659588229`
    - `lastUpdatedQuote = 1659589497`

#### Bulk Query

```rust
fn demo(
    std_ref_addr: Addr,
    symbol_pairs: Vec<String>,
) -> StdResult<ReferenceData> {
    deps.querier.query_wasm_smart(
        &std_ref_addr,
        &QueryMsg::GetReferenceDataBulk {
            symbol_pairs,
        },
    )
}
```

Where the result from `demo(std_ref_addr, [("BTC", "USD"), ("ETH", "BTC")])` would yield:

```
[
    ReferenceData(23131270000000000000000, 1659588229, 1659589497),
    ReferenceData(71601775432131482, 1659588229, 1659588229)
]
```

and the results can be interpreted as:

- BTC/USD
    - `rate = 23131.27 BTC/USD`
    - `lastUpdatedBase = 1659588229`
    - `lastUpdatedQuote = 1659589497`
- ETH/BTC
    - `rate = 0.07160177543213148 ETH/BTC`
    - `lastUpdatedBase = 1659588229`
    - `lastUpdatedQuote = 1659588229`


### RefMedianData

`RefMedianData` is the struct that is returned when querying with `GetMedianRef` or `GetMedianRefDataBulk` where the
bulk variant returns `Vec<RefMedianData>`

`RefMedianData` is defined as:

```rust
pub struct RefMedianData {
  // Median Rates of an asset relative to USD
  pub rates: Vec<Uint64>,
  // The resolve time of the request ID
  pub resolve_time: Uint64,
  // The request ID where the rate was derived from
  pub request_id: Uint64,
}
```

### DeviationData

`RefDeviationData` is the struct that is returned when querying with `GetDeviationRef` or `GetDeviationRefBulk` where the
bulk variant returns `Vec<RefDeviationData>`

`RefDeviationData` is defined as:

```rust
pub struct RefDeviationData {
  // Deviation Rates of an asset relative to USD
  pub rates: Vec<Uint64>,
  // The resolve time of the request ID
  pub resolve_time: Uint64,
  // The request ID where the rate was derived from
  pub request_id: Uint64,
}
```

### Median status
Median status tells us when posting medians to contract is allowed, would error otherwise, if median status is set as false
then queries would also fail with err "MEDIAN DISABLED"
