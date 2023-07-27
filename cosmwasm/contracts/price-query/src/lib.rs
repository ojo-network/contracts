use cosmwasm_std::{
    entry_point, to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Reply, Response, StdError,
    StdResult, SubMsgResponse, Uint64,
};

use thiserror::Error;

use cosmwasm_schema::{cw_serde, QueryResponses};
use cw2::set_contract_version;
use cw_storage_plus::{Item, Map};

use price_feed_helper::helper::oracle_submessage;
use price_feed_helper::verify::*;
use price_feed_helper::HelperError::*;
use price_feed_helper::RequestRelay::*;

const CONTRACT_NAME: &str = "relay_contract";
const CONTRACT_VERSION: &str = "v1.0.0";
const CONFIG_KEY: &str = "config";
const REQUEST_KEY: &str = "request_id";
const PRICE_KEY: &str = "price";
const REPLY_ID: u64 = 1;

#[cw_serde]
pub struct InitMsg {
    pub contract_address: String,
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(Uint64)]
    GetPrice{
        symbol:String,
    },

    #[returns(String)]
    GetRequestId,
}

#[cw_serde]
pub enum ExecuteMsg {
    Request(RequestRateData),
    Callback(CallbackRateData),
}

#[cw_serde]
pub struct State {
    pub contract_address: String,
}

pub const CONFIG: Item<State> = Item::new(CONFIG_KEY);
pub const REQUEST: Item<String> = Item::new(REQUEST_KEY);
pub const PRICE: Map<String,Uint64> = Map::new(PRICE_KEY);

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InitMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    let state = State {
        contract_address: msg.contract_address,
    };

    CONFIG.save(deps.storage, &state)?;

    Ok(Response::new().add_attribute("action", "instantiate"))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Request(msg) => execute_request_relay(deps, env, info, msg),
        ExecuteMsg::Callback(msg) => execute_callback(deps, env, info, msg),
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetPrice{symbol} => to_binary(&query_request(deps,symbol)?),
        QueryMsg::GetRequestId => to_binary(&REQUEST.load(deps.storage)?),
    }
}

fn query_request(deps: Deps,symbol: String) -> StdResult<Uint64> {
    let price = PRICE.may_load(deps.storage,symbol)?.unwrap_or(Uint64::zero());
    Ok(price)
}

fn execute_request_relay(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: RequestRateData,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;

    let oracle_address = config.contract_address;
    let msg = oracle_submessage(
        oracle_address,
        msg.symbol,
        msg.resolve_time,
        msg.callback_data,
        REPLY_ID,
        String::from("callback"),
        RequestType::RequestRate,
    );

    Ok(Response::new()
        .add_submessage(msg)
        .add_attribute("action", "relay_message"))
}

fn execute_callback(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: CallbackRateData,
) -> Result<Response, ContractError> {
    // Implement your callback logic here
    let config = CONFIG.load(deps.storage)?;
    let oracle_address = config.contract_address;

    let prev_id = REQUEST.load(deps.storage)?;
    let request_id = msg.request_id;

    let check =
        is_relayer(&deps, &env, oracle_address, info.sender.to_string()).unwrap_or_default();

    if !check {
        return Err(ContractError::Custom(RelayerError::InvalidRelayer {
            relayer_address: info.sender.to_string(),
        }));
    }

    PRICE.save(deps.storage, msg.symbol.clone(), &msg.symbol_rate)?;

    Ok(Response::new()
        .add_attribute("action", "callback")
        .add_attribute("id", request_id)
        .add_attribute("symbol", msg.symbol)
        .add_attribute("symbol_rate", msg.symbol_rate)
        .add_attribute("last_updated", msg.last_updated)
        .add_attribute("is_verified", check.to_string())
        .add_attribute("prev_id", prev_id))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn reply(deps: DepsMut, _env: Env, reply: Reply) -> StdResult<Response> {
    match reply.id {
        REPLY_ID => process_reply(deps, _env, reply.result.into_result().unwrap()),
        _ => Err(StdError::generic_err("reply id is not 1")),
    }
}

pub fn process_reply(deps: DepsMut, _env: Env, reply: SubMsgResponse) -> StdResult<Response> {
    let id = price_feed_helper::helper::oracle_request_id_from_reply(&reply)?;
    REQUEST.save(deps.storage, &id)?;
    Ok(Response::new().add_attribute("request_id_returned", id))
}

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("{0}")]
    Custom(#[from] RelayerError),
}
