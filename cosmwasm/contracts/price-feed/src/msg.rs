use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Binary, Uint256, Uint64};

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
        callback_sig: String,
        callback_data: Binary,
    },

    RequestMedian {
        symbol: String,
        resolve_time: Uint64,
        callback_sig: String,
        callback_data: Binary,
    },

    RequestDeviation {
        symbol: String,
        resolve_time: Uint64,
        callback_sig: String,
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

    #[returns(Uint256)]
    LastPing { relayer: String },
}
