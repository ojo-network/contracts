use crate::errors::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, MigrateMsg, QueryMsg};
use crate::state::{
    RefData, RefStore, ReferenceData, WhitelistedRelayers, ADMIN, REFDATA, RELAYERS,
};
use cosmwasm_std::{
    entry_point, to_binary, Addr, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdError,
    StdResult, Uint256, Uint64,
};

const E9: Uint64 = Uint64::new(1_000_000_000u64);
const E18: Uint256 = Uint256::from_u128(1_000_000_000_000_000_000u128);

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    mut deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    _msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    // Set sender as admin
    let admin = info.sender;

    ADMIN.save(deps.storage, &admin.clone())?;

    Ok(Response::default())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::UpdateAdmin { admin } => {
            let admin = deps.api.addr_validate(&admin)?;
            let prev_admin = ADMIN.load(deps.storage)?;
            // check if sender is admin
            assert_admin(&prev_admin, &info.sender)?;

            // save new admin
            ADMIN.save(deps.storage, &admin)?;
            Ok(Response::default())
        }
        ExecuteMsg::AddRelayers { relayers } => execute_add_relayers(deps, info, relayers),
        ExecuteMsg::RemoveRelayers { relayers } => execute_remove_relayers(deps, info, relayers),
        ExecuteMsg::Relay {
            symbol_rates,
            resolve_time,
            request_id,
        } => execute_relay(deps, info, symbol_rates, resolve_time, request_id),
        ExecuteMsg::ForceRelay {
            symbol_rates,
            resolve_time,
            request_id,
        } => execute_force_relay(deps, info, symbol_rates, resolve_time, request_id),
    }
}

fn execute_add_relayers(
    deps: DepsMut,
    info: MessageInfo,
    relayers: Vec<String>,
) -> Result<Response, ContractError> {
    // Checks if sender is admin
    let admin = ADMIN.load(deps.storage)?;
    assert_admin(&admin, &info.sender)?;

    // Adds relayer
    for relayer in relayers {
        WhitelistedRelayers::save(deps.storage, &deps.api.addr_validate(&relayer)?)?;
    }

    Ok(Response::new().add_attribute("action", "add_relayers"))
}

fn execute_remove_relayers(
    deps: DepsMut,
    info: MessageInfo,
    relayers: Vec<String>,
) -> Result<Response, ContractError> {
    let admin = ADMIN.load(deps.storage)?;
    assert_admin(&admin, &info.sender)?;

    for relayer in relayers {
        WhitelistedRelayers::remove(deps.storage, &deps.api.addr_validate(&relayer)?)?;
    }

    Ok(Response::new().add_attribute("action", "remove_relayers"))
}

fn execute_relay(
    deps: DepsMut,
    info: MessageInfo,
    symbol_rates: Vec<(String, Uint64)>,
    resolve_time: Uint64,
    request_id: Uint64,
) -> Result<Response, ContractError> {
    // Checks if sender is a relayer
    let sender_addr = &info.sender;
    if !query_is_relayer(deps.as_ref(), sender_addr)? {
        return Err(ContractError::Unauthorized {
            msg: String::from("Sender is not a relayer"),
        });
    }

    // Saves price data
    for (symbol, rate) in symbol_rates {
        if let Some(existing_refdata) = RefStore::load(deps.storage, &symbol) {
            if existing_refdata.resolve_time >= resolve_time {
                continue;
            }
        }
        RefStore::save(
            deps.storage,
            &symbol,
            &RefData::new(rate, resolve_time, request_id),
        )?
    }

    Ok(Response::default().add_attribute("action", "execute_relay"))
}

fn execute_force_relay(
    deps: DepsMut,
    info: MessageInfo,
    symbol_rates: Vec<(String, Uint64)>,
    resolve_time: Uint64,
    request_id: Uint64,
) -> Result<Response, ContractError> {
    let sender_addr = &info.sender;

    if !query_is_relayer(deps.as_ref(), sender_addr)? {
        return Err(ContractError::Unauthorized {
            msg: String::from("Sender is not a relayer"),
        });
    }

    for (symbol, rate) in symbol_rates {
        RefStore::save(
            deps.storage,
            &symbol,
            &RefData::new(rate, resolve_time, request_id),
        )?;
    }

    Ok(Response::default().add_attribute("action", "execute_force_relay"))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Admin {} => to_binary(&ADMIN.load(deps.storage)?),
        QueryMsg::IsRelayer { relayer } => {
            to_binary(&query_is_relayer(deps, &deps.api.addr_validate(&relayer)?)?)
        }
        QueryMsg::GetRef { symbol } => to_binary(&query_ref(deps, &symbol)?),
        QueryMsg::GetReferenceData { symbol_pair } => {
            to_binary(&query_reference_data(deps, &symbol_pair)?)
        }
        QueryMsg::GetReferenceDataBulk { symbol_pairs } => {
            to_binary(&query_reference_data_bulk(deps, &symbol_pairs)?)
        }
    }
}

fn query_is_relayer(deps: Deps, relayer: &Addr) -> StdResult<bool> {
    let status = RELAYERS.contains(deps.storage, relayer);
    Ok(status)
}

fn query_ref(deps: Deps, symbol: &str) -> StdResult<RefData> {
    if symbol == "USD" {
        Ok(RefData::new(E9, Uint64::MAX, Uint64::zero()))
    } else {
        let data = RefStore::load(deps.storage, symbol);
        data.ok_or(StdError::not_found("std_reference::state::RefData"))
    }
}

fn query_reference_data(deps: Deps, symbol_pair: &(String, String)) -> StdResult<ReferenceData> {
    let base = query_ref(deps, &symbol_pair.0)?;
    let quote = query_ref(deps, &symbol_pair.1)?;

    Ok(ReferenceData::new(
        Uint256::from(base.rate)
            .checked_mul(E18)?
            .checked_div(Uint256::from(quote.rate))?,
        base.resolve_time,
        quote.resolve_time,
    ))
}

fn query_reference_data_bulk(
    deps: Deps,
    symbol_pairs: &[(String, String)],
) -> StdResult<Vec<ReferenceData>> {
    symbol_pairs
        .iter()
        .map(|pair| query_reference_data(deps, pair))
        .collect()
}

fn assert_admin(config_admin: &Addr, account: &Addr) -> Result<(), ContractError> {
    if config_admin != account {
        return Err(ContractError::Admin {
            msg: String::from("ADMIN ONLY"),
        });
    }

    Ok(())
}

#[cfg(test)]
mod tests {

    use cosmwasm_std::testing::*;
    use cosmwasm_std::{
        from_binary, BlockInfo, ContractInfo, MessageInfo, OwnedDeps, QueryResponse, ReplyOn,
        SubMsg, Timestamp, TransactionInfo, WasmMsg,
    };

    use crate::msg::ExecuteMsg::{AddRelayers, Relay};

    use super::*;

    // This function will setup the contract for other tests
    fn setup(mut deps: DepsMut, sender: &str) {
        let info = mock_info(sender, &[]);
        let env = mock_env();
        instantiate(deps.branch(), env, info, InstantiateMsg {}).unwrap();
    }

    fn is_relayers(deps: Deps, relayers: Vec<Addr>) -> Vec<bool> {
        relayers
            .iter()
            .map(|r| query_is_relayer(deps, r))
            .collect::<StdResult<Vec<bool>>>()
            .unwrap()
    }

    // This function will setup the relayer for other tests
    fn setup_relayers(mut deps: DepsMut, sender: &str, relayers: Vec<String>) {
        setup(deps.branch(), sender);

        let info = mock_info(sender, &[]);
        let env = mock_env();
        let msg = AddRelayers {
            relayers: relayers.clone(),
        };
        execute(deps.branch(), env, info, msg).unwrap();
    }

    // This function will setup mock relays for other tests
    fn setup_relays(
        mut deps: DepsMut,
        sender: &str,
        relayers: Vec<String>,
        symbol_rates: Vec<(String, Uint64)>,
        resolve_time: Uint64,
        request_id: Uint64,
    ) {
        setup_relayers(deps.branch(), sender, relayers.clone());

        let info = mock_info(relayers[0].as_str(), &[]);
        let env = mock_env();

        let msg = Relay {
            symbol_rates,
            resolve_time,
            request_id,
        };
        execute(deps.branch(), env, info, msg).unwrap();
    }

    mod instantiate {
        use super::*;
        use crate::msg::QueryMsg::Admin;

        #[test]
        fn can_instantiate() {
            let mut deps = mock_dependencies();
            let init_msg = InstantiateMsg {};
            let info = mock_info("owner", &[]);
            let env = mock_env();
            let res = instantiate(deps.as_mut(), env, info.clone(), init_msg).unwrap();
            assert_eq!(0, res.messages.len());
            let saved_admin = ADMIN.load(&deps.storage).unwrap();
            let is_admin = assert_admin(&saved_admin, &info.sender);

            assert_eq!(is_admin.is_err(), false);
        }
    }

    mod relay {
        use std::iter::zip;

        // use cw_controllers::AdminError;

        use crate::msg::ExecuteMsg::{AddRelayers, ForceRelay, Relay, RemoveRelayers, UpdateAdmin};

        use super::*;

        #[test]
        fn change_admin() {
            // Setup
            let mut deps = mock_dependencies();
            let init_msg = InstantiateMsg {};
            let info = mock_info("owner1", &[]);
            let env = mock_env();
            instantiate(deps.as_mut(), env.clone(), info, init_msg).unwrap();

            let msg = UpdateAdmin {
                admin: String::from("owner2"),
            };

            // Test authorized attempt to add relayers
            let info = mock_info("owner1", &[]);
            let env = mock_env();

            execute(deps.as_mut(), env, info, msg.clone()).unwrap();

            // Unauthorized attempt to change admin
            let info = mock_info("owner1", &[]);
            let env = mock_env();

            // would fail as current_owner=owner2
            let err = execute(deps.as_mut(), env, info, msg);
            assert_eq!(err.is_err(), true);

            assert_eq!(
                err.err().unwrap(),
                ContractError::Admin {
                    msg: String::from("ADMIN ONLY"),
                }
            )
        }

        #[test]
        fn add_relayers_by_owner() {
            // Setup
            let mut deps = mock_dependencies();
            let init_msg = InstantiateMsg {};
            let info = mock_info("owner", &[]);
            let env = mock_env();
            instantiate(deps.as_mut(), env.clone(), info, init_msg).unwrap();
            let relayers_to_add: Vec<String> = vec![
                String::from("relayer_1"),
                String::from("relayer_2"),
                String::from("relayer_3"),
            ];

            // Test authorized attempt to add relayers
            let info = mock_info("owner", &[]);
            let env = mock_env();
            let msg = AddRelayers {
                relayers: relayers_to_add.clone(),
            };
            execute(deps.as_mut(), env, info, msg).unwrap();

            assert_eq!(
                is_relayers(
                    deps.as_ref(),
                    relayers_to_add
                        .iter()
                        .map(Addr::unchecked)
                        .collect::<Vec<Addr>>(),
                ),
                [true, true, true]
            );
        }

        #[test]
        fn add_relayers_by_other() {
            // Setup
            let mut deps = mock_dependencies();
            let init_msg = InstantiateMsg {};
            let info = mock_info("owner", &[]);
            let env = mock_env();
            instantiate(deps.as_mut(), env.clone(), info, init_msg).unwrap();

            // Test unauthorized attempt to add relayer
            let info = mock_info("user", &[]);
            let env = mock_env();
            let msg = AddRelayers {
                relayers: vec![String::from("relayer_1")],
            };
            let err = execute(deps.as_mut(), env, info, msg).unwrap_err();
            assert_eq!(
                err,
                ContractError::Admin {
                    msg: String::from("ADMIN ONLY"),
                }
            )
        }

        #[test]
        fn remove_relayers_by_owner() {
            // Setup
            let mut deps = mock_dependencies();
            let relayers_list = vec![
                String::from("relayer_1"),
                String::from("relayer_2"),
                String::from("relayer_3"),
            ];
            setup_relayers(deps.as_mut(), "owner", relayers_list.clone());

            // Remove relayer
            let relayers_to_remove = relayers_list[..2].to_vec();
            let info = mock_info("owner", &[]);
            let env = mock_env();
            let msg = RemoveRelayers {
                relayers: relayers_to_remove,
            };
            execute(deps.as_mut(), env, info, msg).unwrap();

            assert_eq!(
                is_relayers(
                    deps.as_ref(),
                    relayers_list
                        .into_iter()
                        .map(Addr::unchecked)
                        .collect::<Vec<Addr>>(),
                ),
                [false, false, true]
            );
        }

        #[test]
        fn remove_relayers_by_other() {
            // Setup
            let mut deps = mock_dependencies();
            let relayers = vec![String::from("relayer_1")];
            setup_relayers(deps.as_mut(), "owner", relayers.clone());

            // Test unauthorized attempt to remove relayer
            let info = mock_info("user", &[]);
            let env = mock_env();
            let msg = RemoveRelayers { relayers };
            let err = execute(deps.as_mut(), env, info, msg).unwrap_err();

            assert_eq!(
                err,
                ContractError::Admin {
                    msg: String::from("ADMIN ONLY"),
                }
            );
        }

        #[test]
        fn attempt_relay_by_relayer() {
            // Setup
            let mut deps = mock_dependencies();
            let relayer = String::from("relayer");
            setup_relayers(deps.as_mut(), "owner", vec![relayer.clone()]);

            // Test authorized attempt to relay data
            let info = mock_info(relayer.as_str(), &[]);
            let env = mock_env();
            let symbols = vec!["AAA", "BBB", "CCC"]
                .into_iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>();
            let rates = [1000, 2000, 3000]
                .iter()
                .map(|r| Uint64::new(*r))
                .collect::<Vec<Uint64>>();

            let msg = Relay {
                symbol_rates: zip(symbols.clone(), rates.clone())
                    .collect::<Vec<(String, Uint64)>>(),
                resolve_time: Uint64::from(100u64),
                request_id: Uint64::one(),
            };
            execute(deps.as_mut(), env, info, msg).unwrap();

            // Check if relay was successful
            let reference_datas = query_reference_data_bulk(
                deps.as_ref(),
                &symbols
                    .clone()
                    .iter()
                    .map(|s| (s.clone(), String::from("USD")))
                    .collect::<Vec<(String, String)>>(),
            )
            .unwrap();
            let retrieved_rates = reference_datas
                .clone()
                .into_iter()
                .map(|rd| rd.rate / Uint256::from(E9))
                .collect::<Vec<Uint256>>();
            assert_eq!(
                retrieved_rates,
                rates
                    .iter()
                    .map(|r| Uint256::from(*r))
                    .collect::<Vec<Uint256>>()
            );
        }

        #[test]
        fn attempt_relay_by_relayer_with_invalid_resolve_time() {
            // Setup
            let mut deps = mock_dependencies();
            let relayer = String::from("relayer");
            setup_relayers(deps.as_mut(), "owner", vec![relayer.clone()]);

            // Relay initial set of data
            let info = mock_info(relayer.as_str(), &[]);
            let env = mock_env();
            let symbols = vec!["AAA", "BBB", "CCC"]
                .into_iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>();
            let rates = [1000, 2000, 3000]
                .iter()
                .map(|r| Uint64::new(*r))
                .collect::<Vec<Uint64>>();
            let msg = Relay {
                symbol_rates: zip(symbols.clone(), rates.clone())
                    .collect::<Vec<(String, Uint64)>>(),
                resolve_time: Uint64::from(100u64),
                request_id: Uint64::one(),
            };
            execute(deps.as_mut(), env, info, msg).unwrap();

            // Test attempt to relay with invalid resolve times
            let info = mock_info(relayer.as_str(), &[]);
            let env = mock_env();
            let old_rates = [1, 2, 3]
                .iter()
                .map(|r| Uint64::new(*r))
                .collect::<Vec<Uint64>>();
            let msg = Relay {
                symbol_rates: zip(symbols.clone(), old_rates).collect::<Vec<(String, Uint64)>>(),
                resolve_time: Uint64::from(90u64),
                request_id: Uint64::one(),
            };
            execute(deps.as_mut(), env, info, msg).unwrap();

            // Check if relay was successful
            let reference_datas = query_reference_data_bulk(
                deps.as_ref(),
                &symbols
                    .clone()
                    .iter()
                    .map(|s| (s.clone(), String::from("USD")))
                    .collect::<Vec<(String, String)>>(),
            )
            .unwrap();
            let retrieved_rates = reference_datas
                .clone()
                .into_iter()
                .map(|rd| rd.rate / Uint256::from(E9))
                .collect::<Vec<Uint256>>();
            assert_eq!(
                retrieved_rates,
                rates
                    .iter()
                    .map(|r| Uint256::from(*r))
                    .collect::<Vec<Uint256>>()
            );
        }

        #[test]
        fn attempt_relay_by_relayer_with_partially_invalid_resolve_time() {
            // Setup
            let mut deps = mock_dependencies();
            let relayer = String::from("relayer");
            setup_relayers(deps.as_mut(), "owner", vec![relayer.clone()]);

            // Relay initial set of data
            let info = mock_info(relayer.as_str(), &[]);
            let env = mock_env();
            let symbols = vec!["AAA", "BBB", "CCC"]
                .into_iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>();
            let rates = [1000, 2000, 3000]
                .iter()
                .map(|r| Uint64::new(*r))
                .collect::<Vec<Uint64>>();
            let msg = Relay {
                symbol_rates: zip(symbols.clone(), rates).collect::<Vec<(String, Uint64)>>(),
                resolve_time: Uint64::from(10u64),
                request_id: Uint64::one(),
            };
            execute(deps.as_mut(), env, info, msg).unwrap();

            // Only relay one symbol
            let info = mock_info(relayer.as_str(), &[]);
            let env = mock_env();
            let msg = Relay {
                symbol_rates: vec![(String::from("AAA"), Uint64::new(99999))],
                resolve_time: Uint64::from(20u64),
                request_id: Uint64::from(3u64),
            };
            execute(deps.as_mut(), env, info, msg).unwrap();

            // Test attempt to relay with partially invalid resolve times
            let info = mock_info(relayer.as_str(), &[]);
            let env = mock_env();
            let update_rates = [1, 2, 3]
                .iter()
                .map(|r| Uint64::new(*r))
                .collect::<Vec<Uint64>>();
            let msg = Relay {
                symbol_rates: zip(symbols.clone(), update_rates).collect::<Vec<(String, Uint64)>>(),
                resolve_time: Uint64::from(15u64),
                request_id: Uint64::from(2u64),
            };
            execute(deps.as_mut(), env, info, msg).unwrap();

            // Check if relay was successful
            let reference_datas = query_reference_data_bulk(
                deps.as_ref(),
                &symbols
                    .clone()
                    .iter()
                    .map(|s| (s.clone(), String::from("USD")))
                    .collect::<Vec<(String, String)>>(),
            )
            .unwrap();
            let retrieved_rates = reference_datas
                .clone()
                .into_iter()
                .map(|rd| rd.rate / Uint256::from(E9))
                .collect::<Vec<Uint256>>();
            let expected_rates = vec![99999, 2, 3]
                .iter()
                .map(|r| Uint256::from(*r as u128))
                .collect::<Vec<Uint256>>();
            assert_eq!(retrieved_rates, expected_rates);
        }

        #[test]
        fn attempt_relay_by_others() {
            // Setup
            let mut deps = mock_dependencies();
            setup(deps.as_mut(), "owner");

            // Test unauthorized attempt to relay data
            let info = mock_info("user", &[]);
            let env = mock_env();
            let symbols = vec!["AAA", "BBB", "CCC"]
                .into_iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>();
            let rates = [1000, 2000, 3000]
                .iter()
                .map(|r| Uint64::new(*r))
                .collect::<Vec<Uint64>>();
            let msg = Relay {
                symbol_rates: zip(symbols.clone(), rates.clone())
                    .collect::<Vec<(String, Uint64)>>(),
                resolve_time: Uint64::zero(),
                request_id: Uint64::zero(),
            };
            let err = execute(deps.as_mut(), env, info, msg).unwrap_err();
            assert_eq!(
                err,
                ContractError::Unauthorized {
                    msg: String::from("Sender is not a relayer")
                }
            );
        }

        #[test]
        fn attempt_force_relay_by_relayer() {
            // Setup
            let mut deps = mock_dependencies();
            let relayer = String::from("relayer");
            setup_relayers(deps.as_mut(), "owner", vec![relayer.clone()]);

            // Test authorized attempt to relay data
            let info = mock_info(relayer.as_str(), &[]);
            let env = mock_env();
            let symbols = vec!["AAA", "BBB", "CCC"]
                .into_iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>();
            let rates = [1000, 2000, 3000]
                .iter()
                .map(|r| Uint64::new(*r))
                .collect::<Vec<Uint64>>();
            let msg = ForceRelay {
                symbol_rates: zip(symbols.clone(), rates.clone())
                    .collect::<Vec<(String, Uint64)>>(),
                resolve_time: Uint64::from(100u64),
                request_id: Uint64::from(2u64),
            };
            execute(deps.as_mut(), env, info, msg).unwrap();

            // Test attempt to force relay
            let info = mock_info(relayer.as_str(), &[]);
            let env = mock_env();
            let forced_rates = [1, 2, 3]
                .iter()
                .map(|r| Uint64::new(*r))
                .collect::<Vec<Uint64>>();
            let msg = ForceRelay {
                symbol_rates: zip(symbols.clone(), forced_rates.clone())
                    .collect::<Vec<(String, Uint64)>>(),
                resolve_time: Uint64::from(90u64),
                request_id: Uint64::one(),
            };
            execute(deps.as_mut(), env, info, msg).unwrap();

            // Check if forced relay was successful
            let reference_datas = query_reference_data_bulk(
                deps.as_ref(),
                &symbols
                    .clone()
                    .iter()
                    .map(|s| (s.clone(), String::from("USD")))
                    .collect::<Vec<(String, String)>>(),
            )
            .unwrap();
            let retrieved_rates = reference_datas
                .into_iter()
                .map(|rd| rd.rate / Uint256::from(E9))
                .collect::<Vec<Uint256>>();
            assert_eq!(
                retrieved_rates,
                forced_rates
                    .iter()
                    .map(|r| Uint256::from(*r))
                    .collect::<Vec<Uint256>>()
            );
        }

        #[test]
        fn attempt_force_relay_by_other() {
            // Setup
            let mut deps = mock_dependencies();
            setup(deps.as_mut(), "owner");

            // Test unauthorized attempt to relay data
            let info = mock_info("user", &[]);
            let env = mock_env();
            let symbols = vec!["AAA", "BBB", "CCC"]
                .into_iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>();
            let rates = [1000, 2000, 3000]
                .iter()
                .map(|r| Uint64::new(*r))
                .collect::<Vec<Uint64>>();
            let msg = ForceRelay {
                symbol_rates: zip(symbols.clone(), rates.clone())
                    .collect::<Vec<(String, Uint64)>>(),
                resolve_time: Uint64::zero(),
                request_id: Uint64::zero(),
            };
            let err = execute(deps.as_mut(), env, info, msg).unwrap_err();
            assert_eq!(
                err,
                ContractError::Unauthorized {
                    msg: String::from("Sender is not a relayer")
                }
            );
        }
    }

    mod query {
        use std::iter::zip;
        use std::ops::Mul;

        use cosmwasm_std::from_binary;
        use cosmwasm_std::OverflowOperation::Add;

        use crate::msg::QueryMsg::{GetRef, GetReferenceData, GetReferenceDataBulk};

        use super::*;

        #[test]
        fn attempt_query_config() {
            // Setup
            let mut deps = mock_dependencies();
            setup(deps.as_mut(), "owner");

            // Test if query_config results are correct

            let admin = ADMIN.load(&deps.storage).unwrap();
            let is_admin = assert_admin(&admin, &Addr::unchecked("owner"));

            assert_eq!(is_admin.is_err(), false,);

            let is_admin = assert_admin(&admin, &Addr::unchecked("notowner"));

            assert_eq!(is_admin.is_err(), true,);

            assert_eq!(
                is_admin.err().unwrap(),
                ContractError::Admin {
                    msg: String::from("ADMIN ONLY"),
                }
            )
        }

        #[test]
        fn attempt_query_is_relayer() {
            let mut deps = mock_dependencies();
            let relayer = String::from("relayer");
            setup_relayers(deps.as_mut(), "owner", vec![relayer.clone()]);

            // Test if is_relayer results are correct
            assert_eq!(
                query_is_relayer(deps.as_ref(), &Addr::unchecked(relayer.clone())).unwrap(),
                true
            );
            assert_eq!(
                query_is_relayer(deps.as_ref(), &Addr::unchecked("not_a_relayer")).unwrap(),
                false
            );
        }

        #[test]
        fn attempt_query_get_ref() {
            // Setup
            let mut deps = mock_dependencies();
            let relayer = String::from("relayer");
            let symbol = vec![String::from("AAA")];
            let rate = vec![Uint64::new(1000)];
            setup_relays(
                deps.as_mut(),
                "owner",
                vec![relayer.clone()],
                zip(symbol.clone(), rate.clone()).collect(),
                Uint64::from(100u64),
                Uint64::one(),
            );

            // Test if get_ref results are correct
            let env = mock_env();
            let msg = GetRef {
                symbol: symbol[0].to_owned(),
            };
            let binary_res = query(deps.as_ref(), env, msg).unwrap();
            assert_eq!(
                from_binary::<RefData>(&binary_res).unwrap(),
                RefData::new(rate[0], Uint64::from(100u64), Uint64::one())
            );

            // Test invalid symbol
            let env = mock_env();
            let msg = GetRef {
                symbol: String::from("DNE"),
            };
            let err = query(deps.as_ref(), env, msg).unwrap_err();
            assert_eq!(err, StdError::not_found("std_reference::state::RefData"));
        }

        #[test]
        fn attempt_query_get_reference_data() {
            // Setup
            let mut deps = mock_dependencies();
            let relayer = String::from("relayer");
            let symbol = vec![String::from("AAA")];
            let rate = vec![Uint64::new(1000)];
            setup_relays(
                deps.as_mut(),
                "owner",
                vec![relayer.clone()],
                zip(symbol.clone(), rate.clone()).collect(),
                Uint64::from(100u64),
                Uint64::one(),
            );

            // Test if get_reference_data results are correct
            let env = mock_env();
            let msg = GetReferenceData {
                symbol_pair: (symbol[0].clone(), String::from("USD")),
            };
            let binary_res = query(deps.as_ref(), env, msg).unwrap();
            assert_eq!(
                from_binary::<ReferenceData>(&binary_res).unwrap(),
                ReferenceData::new(
                    Uint256::from(rate[0]).mul(Uint256::from(E9)),
                    Uint64::from(100u64),
                    Uint64::MAX,
                )
            );

            // Test invalid symbol
            let env = mock_env();
            let msg = GetReferenceData {
                symbol_pair: (String::from("DNE"), String::from("USD")),
            };
            let err = query(deps.as_ref(), env, msg).unwrap_err();
            assert_eq!(err, StdError::not_found("std_reference::state::RefData"));
            // Test invalid symbols
            let env = mock_env();
            let msg = GetReferenceData {
                symbol_pair: (String::from("DNE1"), String::from("DNE2")),
            };
            let err = query(deps.as_ref(), env, msg).unwrap_err();
            assert_eq!(err, StdError::not_found("std_reference::state::RefData"));
        }

        #[test]
        fn attempt_query_get_reference_data_bulk() {
            // Setup
            let mut deps = mock_dependencies();
            let relayer = String::from("relayer");
            let symbols = vec!["AAA", "BBB", "CCC"]
                .into_iter()
                .map(|s| s.to_string())
                .collect::<Vec<String>>();
            let rates = [1000, 2000, 3000]
                .iter()
                .map(|r| Uint64::new(*r))
                .collect::<Vec<Uint64>>();
            setup_relays(
                deps.as_mut(),
                "owner",
                vec![relayer.clone()],
                zip(symbols.clone(), rates.clone()).collect(),
                Uint64::from(100u64),
                Uint64::one(),
            );

            // Test if get_reference_data results are correct
            let env = mock_env();
            let msg = GetReferenceDataBulk {
                symbol_pairs: symbols
                    .clone()
                    .iter()
                    .map(|s| (s.clone(), String::from("USD")))
                    .collect::<Vec<(String, String)>>(),
            };
            let binary_res = query(deps.as_ref(), env, msg).unwrap();
            let expected_res = rates
                .iter()
                .map(|r| {
                    ReferenceData::new(
                        Uint256::from(*r).mul(Uint256::from(E9)),
                        Uint64::from(100u64),
                        Uint64::MAX,
                    )
                })
                .collect::<Vec<ReferenceData>>();
            assert_eq!(
                from_binary::<Vec<ReferenceData>>(&binary_res).unwrap(),
                expected_res
            );

            // Test invalid symbols
            let env = mock_env();
            let msg = GetReferenceDataBulk {
                symbol_pairs: vec![
                    (String::from("AAA"), String::from("USD")),
                    (String::from("DNE1"), String::from("USD")),
                    (String::from("DNE2"), String::from("USD")),
                ],
            };
            let err = query(deps.as_ref(), env, msg).unwrap_err();
            assert_eq!(err, StdError::not_found("std_reference::state::RefData"));

            // Test invalid symbols
            let env = mock_env();
            let msg = GetReferenceDataBulk {
                symbol_pairs: vec![
                    (String::from("AAA"), String::from("DNE1")),
                    (String::from("DNE2"), String::from("DNE2")),
                    (String::from("BBB"), String::from("DNE1")),
                ],
            };
            let err = query(deps.as_ref(), env, msg).unwrap_err();
            assert_eq!(err, StdError::not_found("std_reference::state::RefData"));
        }
    }
}
