use crate::helpers::{generate_oracle_event, EventType};
use semver::Version;

use cosmwasm_std::{
    entry_point, to_binary, Addr, Binary, Deps, DepsMut, Env, MessageInfo, Order, Response,
    StdError, StdResult, Timestamp, Uint64,
};
use cw2::{get_contract_version, set_contract_version};
use cw_storage_plus::Bound;

use crate::msg::ExecuteMsg::*;
use crate::msg::{ExecuteMsg, InstantiateMsg, MigrateMsg, QueryMsg};

use crate::state::{
    ADMIN, LAST_RELAYER, MEDIANSTATUS, PINGCHECK, PING_THRESHOLD, RELAYERS, TRIGGER_REQUEST,
};

use crate::errors::ContractError;
use crate::errors::ContractError::*;

// Version Info
const CONTRACT_NAME: &str = "ojo-price-feeds";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    mut deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    // Set contract version
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    // Set sender as admin
    ADMIN.set(deps.branch(), Some(info.sender))?;
    MEDIANSTATUS.save(deps.storage, &true)?;
    PING_THRESHOLD.save(deps.storage, &msg.ping_threshold)?;

    Ok(Response::default())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    let block_time = env.block.time;
    match msg {
        UpdateAdmin { admin } => {
            let admin = deps.api.addr_validate(&admin)?;
            Ok(ADMIN.execute_update_admin(deps, info, Some(admin))?)
        }
        AddRelayers { relayers } => execute_add_relayers(deps, info, relayers),
        RemoveRelayers { relayers } => execute_remove_relayers(deps, info, relayers),
        ChangeTrigger { trigger } => {
            ADMIN.assert_admin(deps.as_ref(), &info.sender)?;
            TRIGGER_REQUEST.save(deps.storage, &trigger)?;
            Ok(Response::new().add_attribute("change_trigger", trigger.to_string()))
        }
        RequestRate {
            symbol,
            resolve_time,
            callback_sig,
            callback_data,
        } => execute_request_price(
            deps,
            info,
            block_time,
            symbol,
            resolve_time,
            callback_sig,
            callback_data,
            EventType::RequestRate,
        ),
        RequestDeviation {
            symbol,
            resolve_time,
            callback_sig,
            callback_data,
        } => execute_request_price(
            deps,
            info,
            block_time,
            symbol,
            resolve_time,
            callback_sig,
            callback_data,
            EventType::RequestDeviation,
        ),
        RequestMedian {
            symbol,
            resolve_time,
            callback_sig,
            callback_data,
        } => execute_request_price(
            deps,
            info,
            block_time,
            symbol,
            resolve_time,
            callback_sig,
            callback_data,
            EventType::RequestMedian,
        ),
        RelayerPing {} => {
            let check = query_is_relayer(deps.as_ref(), &info.sender).unwrap_or(false);
            if !check {
                return Err(ContractError::UnauthorizedRelayer {
                    msg: info.sender.to_string(),
                });
            }
            let relayer = deps.api.addr_validate(&info.sender.to_string())?;
            PINGCHECK.save(deps.storage, &relayer, &Uint64::new(block_time.seconds()))?;

            Ok(Response::new()
                .add_attribute("ping", relayer.to_string())
                .add_attribute("timestamp", block_time.to_string()))
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn migrate(deps: DepsMut, _env: Env, _msg: MigrateMsg) -> Result<Response, ContractError> {
    fn from_semver(err: semver::Error) -> StdError {
        StdError::generic_err(format!("Semver: {}", err))
    }

    // New contract version
    let version: Version = CONTRACT_VERSION.parse().map_err(from_semver)?;

    // Current contract version
    let stored_info = get_contract_version(deps.storage)?;

    // Stored contract version
    let stored_version: Version = stored_info.version.parse().map_err(from_semver)?;

    // Check contract type
    if CONTRACT_NAME != stored_info.contract {
        return Err(ContractError::CannotMigrate {
            previous_contract: stored_info.contract,
        });
    }

    // Check new contract version is equal or newer
    if stored_version > version {
        return Err(ContractError::CannotMigrateVersion {
            previous_version: stored_info.version,
        });
    }

    Ok(Response::default())
}

fn execute_add_relayers(
    deps: DepsMut,
    info: MessageInfo,
    relayers: Vec<String>,
) -> Result<Response, ContractError> {
    // Checks if sender is admin
    ADMIN.assert_admin(deps.as_ref(), &info.sender)?;

    // Adds relayer
    for relayer in relayers {
        RELAYERS.save(deps.storage, &deps.api.addr_validate(&relayer)?, &true)?;
    }

    Ok(Response::new().add_attribute("action", "add_relayers"))
}

fn execute_remove_relayers(
    deps: DepsMut,
    info: MessageInfo,
    relayers: Vec<String>,
) -> Result<Response, ContractError> {
    // Checks if sender is admin
    ADMIN.assert_admin(deps.as_ref(), &info.sender)?;

    for relayer in relayers {
        RELAYERS.remove(deps.storage, &deps.api.addr_validate(&relayer)?);
    }

    Ok(Response::new().add_attribute("action", "remove_relayers"))
}

fn execute_request_price(
    deps: DepsMut,
    info: MessageInfo,
    blocktime: Timestamp,
    symbol: String,
    resolve_time: Uint64,
    callback_sig: String,
    callback_data: Binary,
    event_type: EventType,
) -> Result<Response, ContractError> {
    let status = TRIGGER_REQUEST.load(deps.storage).unwrap_or_default();
    if !status {
        return Err(TriggerRequestDisabled {});
    }

    let contract_address = deps.api.addr_validate(&info.sender.to_string()).unwrap();
    let mut request_id = contract_address.clone().to_string();
    request_id.push('_');
    request_id.push_str(blocktime.seconds().to_string().as_str());

    let next_relayer = select_relayer(deps.as_ref(), blocktime.seconds())?;

    LAST_RELAYER.save(deps.storage, &next_relayer.clone().into_string())?;

    let event = generate_oracle_event(
        next_relayer.to_string(),
        info.sender.clone().to_string(),
        symbol.clone(),
        resolve_time.clone(),
        callback_data.clone(),
        request_id.clone(),
        callback_sig,
        event_type,
    );

    Ok(Response::new()
        .add_attribute("action", "request_price")
        .add_attribute("request_id", request_id)
        .add_event(event))
}

pub fn select_relayer(deps: Deps, blocktime: u64) -> Result<Addr, ContractError> {
    // Get the last selected relayer

    let last_relayer = LAST_RELAYER.may_load(deps.storage)?;
    let addr = match &last_relayer {
        Some(addr_str) => deps.api.addr_validate(addr_str)?,
        None => Addr::unchecked(""),
    };

    let start = if last_relayer.is_some() {
        Some(Bound::exclusive(&addr))
    } else {
        Some(Bound::inclusive(&addr))
    };

    let threshold = PING_THRESHOLD.load(deps.storage)?;

    // Find the next available relayer
    let mut iter = RELAYERS.range(deps.storage, start, None, Order::Ascending);
    let mut next_relayer = None;
    while let Some(result) = iter.next() {
        let (key, _) = result?;
        let last_ping = PINGCHECK.load(deps.storage, &key).unwrap_or_default();
        if blocktime - last_ping.u64() <= threshold.u64() {
            next_relayer = Some(key.clone());
            break;
        }
    }

    // If no next relayer was found, start from the beginning
    if next_relayer.is_none() {
        let mut iter = RELAYERS.range(
            deps.storage,
            Some(Bound::inclusive(&Addr::unchecked(""))),
            None,
            Order::Ascending,
        );
        while let Some(result) = iter.next() {
            let (key, _) = result?;
            let last_ping = PINGCHECK.load(deps.storage, &key).unwrap_or_default();
            if blocktime - last_ping.u64() <= threshold.u64() {
                next_relayer = Some(key.clone());
                break;
            }
        }
    }

    // If no available relayer was found, return an error
    next_relayer.ok_or(ContractError::RelayerNotFound {})
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Admin {} => to_binary(&ADMIN.query_admin(deps)?),
        QueryMsg::MedianStatus {} => to_binary(&MEDIANSTATUS.load(deps.storage)?),
        QueryMsg::IsRelayer { relayer } => {
            to_binary(&query_is_relayer(deps, &deps.api.addr_validate(&relayer)?)?)
        }
        QueryMsg::LastPing { relayer } => {
            to_binary(&PINGCHECK.load(deps.storage, &deps.api.addr_validate(&relayer)?)?)
        }
        QueryMsg::PingThreshold {} => to_binary(&PING_THRESHOLD.load(deps.storage)?),
    }
}

pub fn query_is_relayer(deps: Deps, relayer: &Addr) -> StdResult<bool> {
    Ok(RELAYERS.may_load(deps.storage, relayer)?.is_some())
}
