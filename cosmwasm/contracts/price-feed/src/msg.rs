use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Addr, Binary, Uint128, Uint256, Uint64};

use crate::state::{RefData, RefMedianData, ReferenceData};

#[cw_serde]
pub struct InstantiateMsg {
    pub ping_threshold: Uint64,
}

#[cw_serde]
pub struct MigrateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
    // Updates the contract config
    UpdateAdmin {
        // Address of the new owner
        admin: String,
    },

    // Whitelists addresses into relayer set
    AddRelayers {
        // Addresses of the to-be relayers
        relayers: Vec<String>,
    },
    // Removes addresses from relayer set
    RemoveRelayers {
        // Addresses to revoke the relayer rights
        relayers: Vec<String>,
    },

    // Relays a vector of symbols and their corresponding rates
    RequestRate {
        symbol: String,
        resolve_time: Uint64,
        callback_data: Binary,
    },

    RequestMedian {
        symbol: String,
        resolve_time: Uint64,
        callback_data: Binary,
    },

    RequestDeviation {
        symbol: String,
        resolve_time: Uint64,
        callback_data: Binary,
    },

    RelayerPing {},

    ChangeTrigger {
        trigger: bool,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    // Returns admin accounts
    #[returns(cw_controllers::AdminResponse)]
    Admin {},
    //return median status
    #[returns(bool)]
    MedianStatus {},

    #[returns(Uint256)]
    PingThreshold {},

    // Queries if given a address is a relayer
    #[returns(bool)]
    IsRelayer {
        // Address to check relayer status
        relayer: String,
    },
    #[returns(RefData)]
    // Returns the RefData of a given symbol
    GetRef {
        // Symbol to query
        symbol: String,
    },
    #[returns(ReferenceData)]
    // Returns the ReferenceData of a given asset pairing
    GetReferenceData {
        // Symbol pair to query where:
        // symbol_pair := (base_symbol, quote_symbol)
        // e.g. BTC/USD ≡ ("BTC", "USD")
        symbol_pair: (String, String),
    },
    #[returns(Vec < ReferenceData >)]
    // Returns the ReferenceDatas of the given asset pairings
    GetReferenceDataBulk {
        // Vector of Symbol pair to query
        // e.g. <BTC/USD ETH/USD, OJO/BTC> ≡ <("BTC", "USD"), ("ETH", "USD"), ("OJO", "BTC")>
        symbol_pairs: Vec<(String, String)>,
    },
    #[returns(RefMedianData)]
    // Returns the RefMedianData of a given symbol
    GetMedianRef {
        // Symbol to query
        symbol: String,
    },

    #[returns(Vec < RefMedianData >)]
    // Returns the RefMedianData of the given symbols
    GetMedianRefDataBulk {
        // Vector of Symbols to query
        symbols: Vec<String>,
    },
    #[returns(RefData)]
    // Returns the deviation RefData of a given symbol
    GetDeviationRef {
        // Symbol to query
        symbol: String,
    },

    #[returns(Vec < RefData >)]
    // Returns the deviation RefData of the given symbols
    GetDeviationRefBulk {
        // Vector of Symbols to query
        symbols: Vec<String>,
    },
}
