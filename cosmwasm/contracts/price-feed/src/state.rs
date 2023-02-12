use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, Uint256, Uint64};
use cw_controllers::Admin;
use cw_storage_plus::{Item, Map};

// Administrator account
pub const ADMIN: Admin = Admin::new("admin");

// Used to store addresses of relayers and their state
pub const RELAYERS: Map<&Addr, bool> = Map::new("relayers");

// Used to store RefData
pub const REFDATA: Map<&str, RefData> = Map::new("refdata");

// Used to store Median data
pub const MEDIANREFDATA: Map<&str, RefMedianData> = Map::new("medianrefdata");

// Used to store Median Status
pub const MEDIANSTATUS: Item<bool> = Item::new("medianstatus");

// Used to store Deviation data
pub const DEVIATIONDATA: Map<&str, RefData> = Map::new("deviationdata");

#[cw_serde]
pub struct RefData {
    // Rate of an asset relative to USD
    pub rate: Uint64,
    // The resolve time of the request ID
    pub resolve_time: Uint64,
    // The request ID where the rate was derived from
    pub request_id: Uint64,
}

impl RefData {
    pub fn new(rate: Uint64, resolve_time: Uint64, request_id: Uint64) -> Self {
        RefData {
            rate,
            resolve_time,
            request_id,
        }
    }
}

#[cw_serde]
pub struct RefMedianData {
    // Rate of an asset relative to USD
    pub rates: Vec<Uint64>,
    // The resolve time of the request ID
    pub resolve_time: Uint64,
    // The request ID where the rate was derived from
    pub request_id: Uint64,
}

impl RefMedianData {
    pub fn new(rates: Vec<Uint64>, resolve_time: Uint64, request_id: Uint64) -> Self {
        RefMedianData {
            rates,
            resolve_time,
            request_id,
        }
    }
}

#[cw_serde]
pub struct ReferenceData {
    // Pair rate e.g. rate of BTC/USD
    pub rate: Uint256,
    // Unix time of when the base asset was last updated. e.g. Last update time of BTC in Unix time
    pub last_updated_base: Uint64,
    // Unix time of when the quote asset was last updated. e.g. Last update time of USD in Unix time
    pub last_updated_quote: Uint64,
}

impl ReferenceData {
    pub fn new(rate: Uint256, last_updated_base: Uint64, last_updated_quote: Uint64) -> Self {
        ReferenceData {
            rate,
            last_updated_base,
            last_updated_quote,
        }
    }
}
