use cosmwasm_schema::cw_serde;
use serde::{Deserialize, Serialize};
use cosmwasm_std::{Addr, StdResult, Storage, Uint256, Uint64};
use secret_toolkit::serialization::Json;
use secret_toolkit::storage::{Item, Keymap};
use std::ops::Add;
use secret_toolkit::viewing_key::ViewingKeyStore;

// Administrator account
pub const ADMIN_KEY: &[u8] = b"admin";
pub static ADMIN: Item<Addr> = Item::new(ADMIN_KEY);

// Used to store addresses of relayers and their state
pub const RELAYER_KEY: &[u8] = b"relayer";
pub static RELAYERS: Keymap<Addr, bool> = Keymap::new(RELAYER_KEY);

pub struct RelayerStatus {}
impl RelayerStatus {
    pub fn save(store: &mut dyn Storage, relayer: &Addr, status: &bool) -> StdResult<()> {
        RELAYERS.insert(store, &relayer.clone(), status)
    }
}


// Used to store RefData
pub const REFDATA_KEY: &[u8] = b"refdata";
pub static REFDATA: Keymap<String, RefData> = Keymap::new(REFDATA_KEY);

#[derive(Serialize, Debug, Deserialize, Clone, PartialEq, Eq, Default, schemars::JsonSchema)]
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

pub struct RefStore {}
impl RefStore{
    pub fn load(store: &dyn Storage, symbol: &str) -> Option<RefData> {
        REFDATA
            .get(store, &String::from(symbol.clone()))
    }

    pub fn save(store: &mut dyn Storage, symbol: &str, data: &RefData)-> StdResult<()>{
        REFDATA.insert(store, &String::from(symbol.clone()),data )
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
