use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Binary, Event, Uint64};

#[cw_serde]
pub enum EventType {
    RequestRate,
    RequestMedian,
    RequestDeviation,
}

impl ToString for EventType {
    fn to_string(&self) -> String {
        match self {
            EventType::RequestRate => "request_rate".to_string(),
            EventType::RequestMedian => "request_median".to_string(),
            EventType::RequestDeviation => "request_deviation".to_string(),
        }
    }
}

pub fn generate_oracle_event(
    relayer_address: String,
    contract_address: String,
    symbol: String,
    resolve_time: Uint64,
    callback_data: Binary,
    request_id: String,
    callback_sig: String,
    event_type: EventType,
) -> Event {
    let event = Event::new("price-feed");
    event
        .add_attribute("request_type", event_type.to_string())
        .add_attribute("relayer_address", relayer_address)
        .add_attribute("event_contract_address", contract_address)
        .add_attribute("symbol", symbol.clone())
        .add_attribute("resolve_time", resolve_time)
        .add_attribute("callback_signature", callback_sig)
        .add_attribute("callback_data", callback_data.to_string())
        .add_attribute("request_id", request_id.clone())
}
