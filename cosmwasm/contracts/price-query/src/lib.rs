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
use price_feed_helper::RequestRelay::RequestType::{RequestDeviation, RequestMedian, RequestRate};

const CONTRACT_NAME: &str = "relay_contract";
const CONTRACT_VERSION: &str = "v1.0.0";
const CONFIG_KEY: &str = "config";
const RATE_REQUEST_KEY: &str = "request_id";
const MEDIAN_REQUEST_KEY: &str = "median_request_id";
const DEVIATION_REQUEST_KEY: &str = "deviation_request_id";
const PRICE_KEY: &str = "price";
const MEDIAN_KEY: &str = "median";
const DEVIATION_KEY: &str = "deviation";

const RATE_REPLY_ID: u64 = 1;
const MEDIAN_REPLY_ID: u64 = 2;
const DEVIATION_REPLY_ID: u64 = 3;

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

    #[returns(Vec<Uint64>)]
    GetMedian{
        symbol:String,
    },

    #[returns(Vec<Uint64>)]
    GetDeviation{
        symbol:String,
    },

    #[returns(String)]
    GetRateRequestId,

    #[returns(String)]
    GetMedianRequestID,

    #[returns(String)]
    GetDeviationRequestID,
}

#[cw_serde]
pub struct Request {
    pub symbol: String,
    pub callback_data:Binary
}

#[cw_serde]
pub enum ExecuteMsg {
    RequestRate(Request),
    RequestMedian(Request),
    RequestDeviation(Request),

    // having an execute msg for each request callback type
    CallbackRate(CallbackRateData),
    CallbackMedian(CallbackRateMedian),
    CallbackDeviation(CallbackRateDeviation),
}

#[cw_serde]
pub struct State {
    pub contract_address: String,
}

pub const CONFIG: Item<State> = Item::new(CONFIG_KEY);
pub const DEVIATION_REQUEST: Item<String> = Item::new(DEVIATION_REQUEST_KEY);
pub const RATE_REQUEST: Item<String> = Item::new(RATE_REQUEST_KEY);
pub const MEDIAN_REQUEST: Item<String> = Item::new(MEDIAN_REQUEST_KEY);

pub const RATE: Map<String,Uint64> = Map::new(PRICE_KEY);
pub const DEVIATION: Map<String,Vec<Uint64>> = Map::new(DEVIATION_KEY);
pub const MEDIAN: Map<String,Vec<Uint64>> = Map::new(MEDIAN_KEY);

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
        ExecuteMsg::RequestRate(msg) => execute_request_relay(deps, env, info, msg, RATE_REPLY_ID,String::from("callback_rate"),RequestRate),
        ExecuteMsg::RequestMedian(msg) => execute_request_relay(deps, env, info, msg,MEDIAN_REPLY_ID,String::from("callback_median"),RequestMedian),
        ExecuteMsg::RequestDeviation(msg) => execute_request_relay(deps, env, info, msg,DEVIATION_REPLY_ID,String::from("callback_deviation"),RequestDeviation),

        ExecuteMsg::CallbackRate(msg) => execute_rate_callback(deps, env, info, msg),
        ExecuteMsg::CallbackMedian(msg) => execute_historic_median(deps, env, info, msg),
        ExecuteMsg::CallbackDeviation(msg) => execute_historic_deviation(deps, env, info, msg),
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::GetPrice{symbol} => to_binary(&RATE.load(deps.storage,symbol)?),
        QueryMsg::GetMedian {symbol} => to_binary(&MEDIAN.load(deps.storage,symbol)?),
        QueryMsg::GetDeviation{symbol} => to_binary(&DEVIATION.load(deps.storage,symbol)?),
        QueryMsg::GetRateRequestId => to_binary(&RATE_REQUEST.load(deps.storage)?),
        QueryMsg::GetDeviationRequestID => to_binary(&DEVIATION_REQUEST.load(deps.storage)?),
        QueryMsg::GetMedianRequestID => to_binary(&MEDIAN_REQUEST.load(deps.storage)?),
    }
}

fn execute_request_relay(
    deps: DepsMut,
    env: Env,
    _info: MessageInfo,
    msg:Request,
    reply_id: u64,
    callback_sig: String,
    request_type: RequestType,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;

    let oracle_address = config.contract_address;
    let msg = oracle_submessage(
        oracle_address,
        msg.symbol,
        env.block.time.seconds().into(),
        msg.callback_data,
        reply_id,
        callback_sig,
        request_type,
    );

    Ok(Response::new()
        .add_submessage(msg)
        .add_attribute("action", "relay_message"))
}

fn execute_rate_callback(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: CallbackRateData,
) -> Result<Response, ContractError> {
    // Implement your callback logic here
    let config = CONFIG.load(deps.storage)?;
    let oracle_address = config.contract_address;

    let prev_id = RATE_REQUEST.load(deps.storage)?;
    let request_id = msg.request_id;

    let check =
        is_relayer(&deps, &env, oracle_address, info.sender.to_string()).unwrap_or_default();

    if !check {
        return Err(ContractError::Custom(RelayerError::InvalidRelayer {
            relayer_address: info.sender.to_string(),
        }));
    }

    RATE.save(deps.storage, msg.symbol.clone(), &msg.symbol_rate)?;

    Ok(Response::new()
        .add_attribute("action", "rate callback")
        .add_attribute("id", request_id)
        .add_attribute("symbol", msg.symbol)
        .add_attribute("symbol_rate", msg.symbol_rate)
        .add_attribute("last_updated", msg.last_updated)
        .add_attribute("is_verified", check.to_string())
        .add_attribute("prev_id", prev_id))
}

fn execute_historic_deviation(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: CallbackRateDeviation,
) -> Result<Response, ContractError> {
    // Implement your callback logic here
    let config = CONFIG.load(deps.storage)?;
    let oracle_address = config.contract_address;

    let prev_id = DEVIATION_REQUEST.load(deps.storage)?;
    let request_id = msg.request_id;

    let check =
        is_relayer(&deps, &env, oracle_address, info.sender.to_string()).unwrap_or_default();

    if !check {
        return Err(ContractError::Custom(RelayerError::InvalidRelayer {
            relayer_address: info.sender.to_string(),
        }));
    }

    DEVIATION.save(deps.storage, msg.symbol.clone(), &msg.symbol_rates)?;

    Ok(Response::new()
        .add_attribute("action", "deviation callback")
        .add_attribute("id", request_id)
        .add_attribute("symbol", msg.symbol)
        .add_attribute("last_updated", msg.last_updated)
        .add_attribute("is_verified", check.to_string())
        .add_attribute("prev_id", prev_id))
}

fn execute_historic_median(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: CallbackRateMedian,
) -> Result<Response, ContractError> {
    // Implement your callback logic here
    let config = CONFIG.load(deps.storage)?;
    let oracle_address = config.contract_address;

    let prev_id = MEDIAN_REQUEST.load(deps.storage)?;
    let request_id = msg.request_id;

    let check =
        is_relayer(&deps, &env, oracle_address, info.sender.to_string()).unwrap_or_default();

    if !check {
        return Err(ContractError::Custom(RelayerError::InvalidRelayer {
            relayer_address: info.sender.to_string(),
        }));
    }

    MEDIAN.save(deps.storage, msg.symbol.clone(), &msg.symbol_rates)?;

    Ok(Response::new()
        .add_attribute("action", "median callback")
        .add_attribute("id", request_id)
        .add_attribute("symbol", msg.symbol)
        .add_attribute("last_updated", msg.last_updated)
        .add_attribute("is_verified", check.to_string())
        .add_attribute("prev_id", prev_id))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn reply(deps: DepsMut, _env: Env, reply: Reply) -> StdResult<Response> {
    match reply.id {
        RATE_REPLY_ID => process_rate_reply(deps, _env, reply.result.into_result().unwrap()),
        MEDIAN_REPLY_ID=> process_median_reply(deps, _env, reply.result.into_result().unwrap()),
        DEVIATION_REPLY_ID=> process_deviation_reply(deps, _env, reply.result.into_result().unwrap()),
        _ => Err(StdError::generic_err("reply id is not 1")),
    }
}

pub fn process_rate_reply(deps: DepsMut, _env: Env, reply: SubMsgResponse) -> StdResult<Response> {
    let id = price_feed_helper::helper::oracle_request_id_from_reply(&reply)?;
    RATE_REQUEST.save(deps.storage, &id)?;
    Ok(Response::new().add_attribute("request_id_returned", id))
}

pub fn process_median_reply(deps: DepsMut, _env: Env, reply: SubMsgResponse) -> StdResult<Response> {
    let id = price_feed_helper::helper::oracle_request_id_from_reply(&reply)?;
    MEDIAN_REQUEST.save(deps.storage, &id)?;
    Ok(Response::new().add_attribute("median request id returned", id))
}


pub fn process_deviation_reply(deps: DepsMut, _env: Env, reply: SubMsgResponse) -> StdResult<Response> {
    let id = price_feed_helper::helper::oracle_request_id_from_reply(&reply)?;
    DEVIATION_REQUEST.save(deps.storage, &id)?;
    Ok(Response::new().add_attribute("deviation request id returned", id))
}


#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("{0}")]
    Custom(#[from] RelayerError),
}
