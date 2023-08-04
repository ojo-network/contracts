pub mod state;
pub mod msg;

use state::*;
use msg::{QueryMsg,ExecuteMsg,InitMsg,Request};
use cosmwasm_std::{entry_point, to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Reply, Response, StdError, StdResult, SubMsgResponse, Uint64};

use thiserror::Error;
use cw2::set_contract_version;

use price_feed_helper::helper::oracle_submessage;
use price_feed_helper::verify::*;
use price_feed_helper::HelperError::*;
use price_feed_helper::RequestRelay::*;
use price_feed_helper::RequestRelay::RequestType::{RequestDeviation, RequestMedian, RequestRate};

const CONTRACT_NAME: &str = "relay_contract";
const CONTRACT_VERSION: &str = "v1.0.0";
const RATE_REPLY_ID: u64 = 1;
const MEDIAN_REPLY_ID: u64 = 2;
const DEVIATION_REPLY_ID: u64 = 3;


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
        QueryMsg::GetRateRequestId{symbol} => to_binary(&RATE_REQUEST.load(deps.storage,symbol)?),
        QueryMsg::GetDeviationRequestId{symbol} => to_binary(&DEVIATION_REQUEST.load(deps.storage,symbol)?),
        QueryMsg::GetMedianRequestId{symbol} => to_binary(&MEDIAN_REQUEST.load(deps.storage,symbol)?),
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
    // Implement callback logic
    let config = CONFIG.load(deps.storage)?;
    let oracle_address = config.contract_address;
    let id = RATE_REQUEST.load(deps.storage,msg.symbol.clone())?;

    // check if request id matches
    if id!=msg.request_id {
        return Err(ContractError::RequestIDMismatch(id, msg.request_id));
    }

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
        .add_attribute("id", id)
        .add_attribute("symbol", msg.symbol)
        .add_attribute("symbol_rate", msg.symbol_rate)
        .add_attribute("last_updated", msg.last_updated)
        .add_attribute("is_verified", check.to_string()))
}

fn execute_historic_deviation(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: CallbackRateDeviation,
) -> Result<Response, ContractError> {
    // Implement callback logic
    let config = CONFIG.load(deps.storage)?;
    let oracle_address = config.contract_address;
    let id = DEVIATION_REQUEST.load(deps.storage,msg.symbol.clone())?;

    // check if request id matches
    if id!=msg.request_id {
        return Err(ContractError::RequestIDMismatch(id, msg.request_id));
    }

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
        .add_attribute("id", id)
        .add_attribute("symbol", msg.symbol)
        .add_attribute("last_updated", msg.last_updated)
        .add_attribute("is_verified", check.to_string()))
}

fn execute_historic_median(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: CallbackRateMedian,
) -> Result<Response, ContractError> {
    // Implement callback logic
    let config = CONFIG.load(deps.storage)?;
    let oracle_address = config.contract_address;
    let id = MEDIAN_REQUEST.load(deps.storage,msg.symbol.clone())?;

    // check if request id matches
    if id!=msg.request_id {
        return Err(ContractError::RequestIDMismatch(id, msg.request_id));
    }

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
        .add_attribute("id", id)
        .add_attribute("symbol", msg.symbol)
        .add_attribute("last_updated", msg.last_updated)
        .add_attribute("is_verified", check.to_string()))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn reply(deps: DepsMut, _env: Env, reply: Reply) -> StdResult<Response> {
    match reply.id {
        RATE_REPLY_ID => process_rate_reply(deps, _env, reply.result.into_result().unwrap()),
        MEDIAN_REPLY_ID=> process_median_reply(deps, _env, reply.result.into_result().unwrap()),
        DEVIATION_REPLY_ID=> process_deviation_reply(deps, _env, reply.result.into_result().unwrap()),
        _ => Err(StdError::generic_err("reply id is not supported")),
    }
}

pub fn process_rate_reply(deps: DepsMut, _env: Env, reply: SubMsgResponse) -> StdResult<Response> {
    let reply = price_feed_helper::helper::id_and_symbol_from_reply(&reply)?;
    RATE_REQUEST.save(deps.storage,reply.symbol.clone(), &reply.request_id.clone())?;
    Ok(Response::new().add_attribute("request_id_returned",reply.request_id).add_attribute("symbol", reply.symbol))
}

pub fn process_median_reply(deps: DepsMut, _env: Env, reply: SubMsgResponse) -> StdResult<Response> {
    let reply = price_feed_helper::helper::id_and_symbol_from_reply(&reply)?;
    MEDIAN_REQUEST.save(deps.storage,reply.symbol.clone(), &reply.request_id.clone())?;
    Ok(Response::new().add_attribute("median_request_id_returned",reply.request_id).add_attribute("symbol", reply.symbol))
}


pub fn process_deviation_reply(deps: DepsMut, _env: Env, reply: SubMsgResponse) -> StdResult<Response> {
    let reply = price_feed_helper::helper::id_and_symbol_from_reply(&reply)?;
    DEVIATION_REQUEST.save(deps.storage,reply.symbol.clone(), &reply.request_id.clone())?;
    Ok(Response::new().add_attribute("deviation_request_id_returned", reply.request_id).add_attribute("symbol", reply.symbol))
}


#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("{0}")]
    Custom(#[from] RelayerError),

    #[error("Request id mismatches {0} {1}")]
    RequestIDMismatch(String, String),
}
