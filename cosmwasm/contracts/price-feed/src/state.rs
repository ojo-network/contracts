use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, StdResult, Storage, Uint256, Uint64};
use secret_toolkit::serialization::{Bincode2, Json};
use secret_toolkit::storage::{Item, Keymap, Keyset, KeysetBuilder, WithoutIter};
use serde::{Deserialize, Serialize};

// Administrator account
pub const ADMIN_KEY: &[u8] = b"admin";
pub static ADMIN: Item<Addr> = Item::new(ADMIN_KEY);

// Used to store addresses of relayers and their state
pub const RELAYER_KEY: &[u8] = b"relayer";
// build without iter
pub static RELAYERS: Keyset<Addr, Bincode2, WithoutIter> =
    KeysetBuilder::new(RELAYER_KEY).without_iter().build();

pub struct WhitelistedRelayers {}
impl WhitelistedRelayers {
    pub fn save(store: &mut dyn Storage, relayer: &Addr) -> StdResult<()> {
        RELAYERS.insert(store, &relayer.clone())
    }

    pub fn remove(store: &mut dyn Storage, relayer: &Addr) -> StdResult<()> {
        RELAYERS.remove(store, &relayer.clone())
    }
}

// Used to store RefData
pub const REFDATA_KEY: &[u8] = b"refdata";
pub static REFDATA: Keymap<String, RefData> = Keymap::new(REFDATA_KEY);

// Used to store Deviation data
pub const DEVIATIONDATA_KEY: &[u8] = b"deviationdata";
pub static DEVIATIONDATA: Keymap<String, RefData> = Keymap::new(DEVIATIONDATA_KEY);

// Stores Median status
pub const MEDIANSTATUS_KEY: &[u8]= b"medianstatus";
pub static MEDIANSTATUS: Item<bool>= Item::new(MEDIANSTATUS_KEY);


// Used to store Median data
pub const MEDIANDATA_KEY: &[u8] = b"mediandata";
pub static MEDIANDATA: Keymap<String, RefMedianData> = Keymap::new(MEDIANDATA_KEY);

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
impl RefStore {
    pub fn load(store: &dyn Storage, symbol: &str) -> Option<RefData> {
        REFDATA.get(store, &String::from(symbol.clone()))
    }

    pub fn save(store: &mut dyn Storage, symbol: &str, data: &RefData) -> StdResult<()> {
        REFDATA.insert(store, &String::from(symbol.clone()), data)
    }
}

// store deviation data
pub struct RefDeviationStore {}
impl RefDeviationStore {
    pub fn load(store: &dyn Storage, symbol: &str) -> Option<RefData> {
        DEVIATIONDATA.get(store, &String::from(symbol.clone()))
    }

    pub fn save(store: &mut dyn Storage, symbol: &str, data: &RefData) -> StdResult<()> {
        DEVIATIONDATA.insert(store, &String::from(symbol.clone()), data)
    }
}


#[derive(Serialize, Debug, Deserialize, Clone, PartialEq, Eq, Default, schemars::JsonSchema)]
pub struct RefMedianData {
    // Median Rates of an asset relative to USD
    pub rates: Vec<Uint64>,
    // The resolve time of the request ID
    pub resolve_time: Uint64,
    // The request ID where the rate was derived from
    pub request_id: Uint64,
}

impl RefMedianData{
    pub fn new(rates: Vec<Uint64>, resolve_time: Uint64, request_id: Uint64) -> Self {
        RefMedianData {
            rates,
            resolve_time,
            request_id,
        }
    }
}

// stores median data
pub struct RefMedianStore{}
impl RefMedianStore {
    pub fn load(store: &dyn Storage, symbol: &str) -> Option<RefMedianData> {
        MEDIANDATA.get(store, &String::from(symbol.clone()))
    }

    pub fn save(store: &mut dyn Storage, symbol: &str, data: &RefMedianData) -> StdResult<()> {
        MEDIANDATA.insert(store, &String::from(symbol.clone()), data)
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
