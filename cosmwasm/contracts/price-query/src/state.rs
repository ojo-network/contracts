use cw_storage_plus::{Item, Map};
use cosmwasm_std::Uint64;
use cosmwasm_schema::cw_serde;

#[cw_serde]
pub struct State {
    pub contract_address: String,
}

const CONFIG_KEY: &str = "config";
const RATE_REQUEST_KEY: &str = "request_id";
const MEDIAN_REQUEST_KEY: &str = "median_request_id";
const DEVIATION_REQUEST_KEY: &str = "deviation_request_id";
const PRICE_KEY: &str = "price";
const MEDIAN_KEY: &str = "median";
const DEVIATION_KEY: &str = "deviation";

pub const CONFIG: Item<State> = Item::new(CONFIG_KEY);
pub const DEVIATION_REQUEST: Map<String,String> = Map::new(DEVIATION_REQUEST_KEY);
pub const RATE_REQUEST: Map<String,String> = Map::new(RATE_REQUEST_KEY);
pub const MEDIAN_REQUEST: Map<String,String> = Map::new(MEDIAN_REQUEST_KEY);

pub const RATE: Map<String,Uint64> = Map::new(PRICE_KEY);
pub const DEVIATION: Map<String,Vec<Uint64>> = Map::new(DEVIATION_KEY);
pub const MEDIAN: Map<String,Vec<Uint64>> = Map::new(MEDIAN_KEY);