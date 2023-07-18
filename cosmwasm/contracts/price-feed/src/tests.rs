#[cfg(test)]
mod tests {
    use crate::contract::*;
    use crate::errors::*;
    use crate::helpers::EventType;
    use crate::msg::ExecuteMsg::*;
    use crate::msg::InstantiateMsg;
    use crate::state::{ADMIN, LAST_RELAYER, PINGCHECK, TRIGGER_REQUEST};
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{Addr, Binary, Deps, DepsMut, StdResult, Uint64};

    fn default_init_msg() -> InstantiateMsg {
        InstantiateMsg {
            ping_threshold: Uint64::from(60 as u64),
        }
    }
    // This function will setup the contract for other tests
    fn setup(mut deps: DepsMut, sender: &str) {
        let info = mock_info(sender, &[]);
        let env = mock_env();
        instantiate(deps.branch(), env, info, default_init_msg()).unwrap();
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

    fn execute_ping(mut deps: DepsMut, relayers: Vec<String>) {
        for relayer in relayers {
            let info = mock_info(relayer.as_str(), &[]);
            let env = mock_env();
            let msg = RelayerPing {};
            let res = execute(deps.branch(), env, info, msg);
            assert!(res.is_ok());
        }
    }

    mod instantiate {
        use super::*;

        #[test]
        fn can_instantiate() {
            let mut deps = mock_dependencies();
            let init_msg = default_init_msg();
            let info = mock_info("owner", &[]);
            let env = mock_env();
            let res = instantiate(deps.as_mut(), env, info.clone(), init_msg).unwrap();
            assert_eq!(0, res.messages.len());
            assert_eq!(ADMIN.is_admin(deps.as_ref(), &info.sender).unwrap(), true);
        }
    }

    mod relay {
        use cw_controllers::AdminError;

        use crate::msg::ExecuteMsg::{AddRelayers, RemoveRelayers};

        use super::*;

        #[test]
        fn add_relayers_by_owner() {
            // Setup
            let mut deps = mock_dependencies();
            let init_msg = default_init_msg();
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
            let init_msg = default_init_msg();
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
                    0: AdminError::NotAdmin {}
                }
            );
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
                    0: AdminError::NotAdmin {}
                }
            );
        }
    }

    #[test]
    fn test_select_relayer() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("owner", &[]);
        setup(deps.as_mut(), &info.sender.as_str());

        // add relayuers
        let relayers_to_add: Vec<String> = vec![String::from("relayer_1")];

        let msg = AddRelayers {
            relayers: relayers_to_add.clone(),
        };

        execute(deps.as_mut(), env.clone(), info, msg).unwrap();

        // ping updates for the relayer 1
        let blocktime = env.block.time;
        let relayer_info = mock_info("relayer_1", &[]);
        let ping_msg = RelayerPing {};

        execute(
            deps.as_mut(),
            env.clone(),
            relayer_info.clone(),
            ping_msg.clone(),
        )
        .unwrap();

        // pingcheck
        let check = PINGCHECK.load(&deps.storage, &relayer_info.sender).unwrap();
        assert_eq!(check.u64(), blocktime.seconds());

        let res = select_relayer(deps.as_ref(), blocktime.seconds()).unwrap();
        assert_eq!(relayer_info.sender, res);

        // check again to see if the relayer is the same
        let res = select_relayer(deps.as_ref(), blocktime.seconds()).unwrap();
        assert_eq!(relayer_info.sender, res);
    }

    #[test]
    fn test_execute_demand_price() {
        // Create a mock API instance and storage
        let mut deps = mock_dependencies();
        let env = mock_env();
        let sender_info = mock_info("sender", &[]);
        setup(deps.as_mut(), "owner");

        // Set the trigger request status to enabled in storage
        TRIGGER_REQUEST.save(&mut deps.storage, &true).unwrap();

        // Verify the storage changes
        let status = TRIGGER_REQUEST.load(&deps.storage).unwrap();
        assert!(status);

        let relayers_to_add: Vec<String> = vec![
            String::from("relayer_1"),
            String::from("relayer_2"),
            String::from("relayer_3"),
        ];

        setup_relayers(deps.as_mut(), "owner", relayers_to_add.clone());
        execute_ping(deps.as_mut(), relayers_to_add.clone());

        // pingcheck
        let check = PINGCHECK
            .load(&deps.storage, &Addr::unchecked("relayer_1"))
            .unwrap();
        assert_eq!(check.u64(), env.block.time.seconds());

        // setup request msg
        let symbol = "TEST".to_string();
        let resolve_time = Uint64::from(10 as u64);
        let callback_sig = "callback_signature".to_string();
        let callback_data = Binary::from(b"callback_data".to_vec());

        let mut expected_request_id = "sender_".to_owned();
        expected_request_id.push_str(env.clone().block.time.seconds().to_string().as_str());

        let relay_demand = RequestRate {
            symbol: symbol.clone(),
            resolve_time: resolve_time.clone(),
            callback_sig: callback_sig.clone(),
            callback_data: callback_data.clone(),
        };

        let res1 = execute(
            deps.as_mut(),
            env.clone(),
            sender_info.clone(),
            relay_demand.clone(),
        );
        assert!(res1.is_ok());

        let next_relayer = LAST_RELAYER.load(&deps.storage).unwrap();
        assert_eq!(next_relayer, "relayer_1");

        // next relayer should change
        let res2 = execute(
            deps.as_mut(),
            env.clone(),
            sender_info.clone(),
            relay_demand.clone(),
        );
        assert!(res2.is_ok());

        let next_relayer = LAST_RELAYER.load(&deps.storage).unwrap();
        assert_eq!(next_relayer, "relayer_2");

        // Verify the emitted event
        let found = res1.unwrap().events.iter().any(|event| {
            // check total lenght of events
            let total_events = event.attributes.len();
            assert_eq!(total_events, 8);

            let attrs = &event.attributes;
            attrs
                .iter()
                .all(|attr| match (attr.key.as_str(), attr.value.as_str()) {
                    ("request_id", value) => value == expected_request_id,
                    ("relayer_address", value) => value == "relayer_1",
                    ("event_contract_address", value) => value == "sender",
                    ("symbol", value) => value == symbol,
                    ("resolve_time", value) => value == "10",
                    ("callback_data", value) => value == callback_data.to_string(),
                    ("callback_signature", value) => value == callback_sig,
                    ("request_type", value) => value == EventType::RequestRate.to_string(),
                    _ => {
                        println!("Unexpected attribute: {}", attr.key.as_str());
                        false
                    }
                })
        });
        assert!(found);
    }

    mod query {
        use cosmwasm_std::testing::{mock_dependencies, mock_env};
        use cosmwasm_std::from_binary;

        use crate::msg::QueryMsg::{LastPing, PingThreshold};

        use super::*;
        #[test]
        fn attempt_query_config() {
            // Setup
            let mut deps = mock_dependencies();
            setup(deps.as_mut(), "owner");

            // Test if query_config results are correct
            assert_eq!(
                ADMIN
                    .is_admin(deps.as_ref(), &Addr::unchecked("owner"))
                    .unwrap(),
                true
            );
        }

        #[test]
        fn attempt_query_is_relayer() {
            let mut deps = mock_dependencies();
            let relayer = String::from("relayer");
            setup_relayers(deps.as_mut(), "owner", vec![relayer.clone()]);

            // Test if is_relayer results are correct
            assert!(
                query_is_relayer(deps.as_ref(), &Addr::unchecked(relayer.clone())).unwrap()
            );
            assert_eq!(
                query_is_relayer(deps.as_ref(), &Addr::unchecked("not_a_relayer")).unwrap(),
                false
            );
        }

        #[test]
        fn attempt_query_ping_threshold() {
            let mut deps = mock_dependencies();
            let env= mock_env();
            let init_msg= default_init_msg();
            setup(deps.as_mut(), "owner");

            let res = query(deps.as_ref(),env,PingThreshold {});
            assert!(res.is_ok());

            let threshold:Uint64=from_binary(&res.unwrap()).unwrap();
            assert_eq!(threshold, init_msg.ping_threshold);  
        }

        #[test]
        fn attempt_query_last_ping() {
            let mut deps = mock_dependencies();
            let relayer = String::from("relayer");
            let env = mock_env();


            let blocktime = env.block.time;
            setup_relayers(deps.as_mut(), "owner", vec![relayer.clone()]);
            execute_ping(deps.as_mut(),vec![relayer.clone()]);

            // Test if is_relayer results are correct
            let ping=query(deps.as_ref(),env,LastPing { relayer: relayer.clone().to_string()});
            assert!(ping.is_ok());

            let pingtime:Uint64=from_binary(&ping.unwrap()).unwrap();
            assert_eq!(pingtime.u64(), blocktime.seconds());
        }
    }
}
