use cosmwasm_std::{Binary, Event, Uint64};

pub fn generate_oracle_event(
    relayer_address: String,
    contract_address: String,
    symbol: String,
    resolve_time: Uint64,
    callback_data: Binary,
    request_id: String,
) -> Event {
    Event::new("price-feed")
        .add_attribute("relayer_address", relayer_address)
        .add_attribute("event_contract_address", contract_address)
        .add_attribute("symbol", symbol.clone())
        .add_attribute("resolve_time", resolve_time)
        .add_attribute("callback_data", callback_data.to_string())
        .add_attribute("request_id", request_id.clone())
}
