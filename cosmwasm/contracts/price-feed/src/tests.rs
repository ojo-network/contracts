
#[cfg(test)]
mod tests {
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{Addr, Uint256,Uint64,DepsMut,Deps,StdResult,Binary};
    use cw_controllers::*;
    use crate::state::{ADMIN, PINGCHECK};
    use crate::contract::*;
    use crate::msg::ExecuteMsg::*;
    use crate::errors::*;
    use crate::msg::InstantiateMsg;

    use super::*;

    fn defualt_init_msg() -> InstantiateMsg{
        InstantiateMsg{
            ping_threshold:Uint64::from(60 as u64)
        }
    }
    // This function will setup the contract for other tests
    fn setup(mut deps: DepsMut, sender: &str) {
        let info = mock_info(sender, &[]);
        let env = mock_env();
        instantiate(deps.branch(), env, info, defualt_init_msg()).unwrap();
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
    fn setup_request_relays(
        mut deps: DepsMut,
        sender: &str,
        relayers: Vec<String>,
        symbol:String,
        resolve_time: Uint64,
        callback_sig: String,
        callback_data: Binary,
    ) {
        setup_relayers(deps.branch(), sender, relayers.clone());

        let info = mock_info(relayers[0].as_str(), &[]);
        let env = mock_env();

        let msg = RequestRate  {
            symbol,
            resolve_time,
            callback_sig,
            callback_data,
        };
        execute(deps.branch(), env, info, msg).unwrap();
    }

    mod instantiate {
        use super::*;

        #[test]
        fn can_instantiate() {
            let mut deps = mock_dependencies();
            let init_msg = defualt_init_msg();
            let info = mock_info("owner", &[]);
            let env = mock_env();
            let res = instantiate(deps.as_mut(), env, info.clone(), init_msg).unwrap();
            assert_eq!(0, res.messages.len());
            assert_eq!(ADMIN.is_admin(deps.as_ref(), &info.sender).unwrap(), true);
        }
    }

    mod relay {
        use cw_controllers::AdminError;
        use std::iter::zip;

        use crate::msg::ExecuteMsg::{
            AddRelayers, RemoveRelayers,
        };

        use super::*;

        #[test]
        fn add_relayers_by_owner() {
            // Setup
            let mut deps = mock_dependencies();
            let init_msg = defualt_init_msg();
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
            let init_msg = defualt_init_msg();
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
        let mut env= mock_env();
        let info = mock_info("owner", &[]);
        setup(deps.as_mut(),&info.sender.as_str());
        
        // add relayuers
        let relayers_to_add: Vec<String> = vec![
            String::from("relayer_1"),
        ];

        let msg = AddRelayers {
            relayers: relayers_to_add.clone(),
        };

        execute(deps.as_mut(), env.clone(), info, msg).unwrap();

        // ping updates for the relayer 1
        let blocktime = env.block.time;
        let relayer_info=mock_info("relayer_1", &[]);
        let ping_msg= RelayerPing {};

        execute(deps.as_mut(), env, relayer_info.clone(), ping_msg).unwrap();

        // pingcheck
        let check= PINGCHECK.load(&deps.storage, &relayer_info.sender).unwrap();
        assert_eq!(check.u64(),blocktime.seconds());
                
        let res = select_relayer(deps.as_ref(),blocktime.seconds()).unwrap();
        
        assert_eq!(relayer_info.sender,res);
    
    }



    // mod query {
    //     use std::iter::zip;
    //     use std::ops::Mul;
    //
    //     use cosmwasm_std::from_binary;
    //
    //     use crate::msg::QueryMsg::{GetRef, GetReferenceData, GetReferenceDataBulk};
    //
    //     use super::*;
    //
    //     #[test]
    //     fn attempt_query_config() {
    //         // Setup
    //         let mut deps = mock_dependencies();
    //         setup(deps.as_mut(), "owner");
    //
    //         // Test if query_config results are correct
    //         assert_eq!(
    //             ADMIN
    //                 .is_admin(deps.as_ref(), &Addr::unchecked("owner"))
    //                 .unwrap(),
    //             true
    //         );
    //     }
    //
    //     #[test]
    //     fn attempt_query_is_relayer() {
    //         let mut deps = mock_dependencies();
    //         let relayer = String::from("relayer");
    //         setup_relayers(deps.as_mut(), "owner", vec![relayer.clone()]);
    //
    //         // Test if is_relayer results are correct
    //         assert_eq!(
    //             query_is_relayer(deps.as_ref(), &Addr::unchecked(relayer.clone())).unwrap(),
    //             true
    //         );
    //         assert_eq!(
    //             query_is_relayer(deps.as_ref(), &Addr::unchecked("not_a_relayer")).unwrap(),
    //             false
    //         );
    //     }
    //
    //     #[test]
    //     fn attempt_query_get_ref() {
    //         // Setup
    //         let mut deps = mock_dependencies();
    //         let relayer = String::from("relayer");
    //         let symbol = vec![String::from("AAA")];
    //         let rate = vec![Uint64::new(1000)];
    //         setup_relays(
    //             deps.as_mut(),
    //             "owner",
    //             vec![relayer.clone()],
    //             zip(symbol.clone(), rate.clone()).collect(),
    //             Uint64::from(100u64),
    //             Uint64::one(),
    //         );
    //
    //         // Test if get_ref results are correct
    //         let env = mock_env();
    //         let msg = GetRef {
    //             symbol: symbol[0].to_owned(),
    //         };
    //         let binary_res = query(deps.as_ref(), env, msg).unwrap();
    //         assert_eq!(
    //             from_binary::<RefData>(&binary_res).unwrap(),
    //             RefData::new(rate[0], Uint64::from(100u64), Uint64::one())
    //         );
    //
    //         // Test invalid symbol
    //         let env = mock_env();
    //         let msg = GetRef {
    //             symbol: String::from("DNE"),
    //         };
    //         let err = query(deps.as_ref(), env, msg).unwrap_err();
    //         assert_eq!(err, StdError::not_found("std_reference::state::RefData"));
    //     }
    //
    //     #[test]
    //     fn attempt_query_get_reference_data() {
    //         // Setup
    //         let mut deps = mock_dependencies();
    //         let relayer = String::from("relayer");
    //         let symbol = vec![String::from("AAA")];
    //         let rate = vec![Uint64::new(1000)];
    //         setup_relays(
    //             deps.as_mut(),
    //             "owner",
    //             vec![relayer.clone()],
    //             zip(symbol.clone(), rate.clone()).collect(),
    //             Uint64::from(100u64),
    //             Uint64::one(),
    //         );
    //
    //         // Test if get_reference_data results are correct
    //         let env = mock_env();
    //         let msg = GetReferenceData {
    //             symbol_pair: (symbol[0].clone(), String::from("USD")),
    //         };
    //         let binary_res = query(deps.as_ref(), env, msg).unwrap();
    //         assert_eq!(
    //             from_binary::<ReferenceData>(&binary_res).unwrap(),
    //             ReferenceData::new(
    //                 Uint256::from(rate[0]).mul(Uint256::from(E9)),
    //                 Uint64::from(100u64),
    //                 Uint64::MAX,
    //             )
    //         );
    //
    //         // Test invalid symbol
    //         let env = mock_env();
    //         let msg = GetReferenceData {
    //             symbol_pair: (String::from("DNE"), String::from("USD")),
    //         };
    //         let err = query(deps.as_ref(), env, msg).unwrap_err();
    //         assert_eq!(err, StdError::not_found("std_reference::state::RefData"));
    //         // Test invalid symbols
    //         let env = mock_env();
    //         let msg = GetReferenceData {
    //             symbol_pair: (String::from("DNE1"), String::from("DNE2")),
    //         };
    //         let err = query(deps.as_ref(), env, msg).unwrap_err();
    //         assert_eq!(err, StdError::not_found("std_reference::state::RefData"));
    //     }
    //
    //     #[test]
    //     fn attempt_query_get_reference_data_bulk() {
    //         // Setup
    //         let mut deps = mock_dependencies();
    //         let relayer = String::from("relayer");
    //         let symbols = vec!["AAA", "BBB", "CCC"]
    //             .into_iter()
    //             .map(|s| s.to_string())
    //             .collect::<Vec<String>>();
    //         let rates = [1000, 2000, 3000]
    //             .iter()
    //             .map(|r| Uint64::new(*r))
    //             .collect::<Vec<Uint64>>();
    //         setup_relays(
    //             deps.as_mut(),
    //             "owner",
    //             vec![relayer.clone()],
    //             zip(symbols.clone(), rates.clone()).collect(),
    //             Uint64::from(100u64),
    //             Uint64::one(),
    //         );
    //
    //         // Test if get_reference_data results are correct
    //         let env = mock_env();
    //         let msg = GetReferenceDataBulk {
    //             symbol_pairs: symbols
    //                 .clone()
    //                 .iter()
    //                 .map(|s| (s.clone(), String::from("USD")))
    //                 .collect::<Vec<(String, String)>>(),
    //         };
    //         let binary_res = query(deps.as_ref(), env, msg).unwrap();
    //         let expected_res = rates
    //             .iter()
    //             .map(|r| {
    //                 ReferenceData::new(
    //                     Uint256::from(*r).mul(Uint256::from(E9)),
    //                     Uint64::from(100u64),
    //                     Uint64::MAX,
    //                 )
    //             })
    //             .collect::<Vec<ReferenceData>>();
    //         assert_eq!(
    //             from_binary::<Vec<ReferenceData>>(&binary_res).unwrap(),
    //             expected_res
    //         );
    //
    //         // Test invalid symbols
    //         let env = mock_env();
    //         let msg = GetReferenceDataBulk {
    //             symbol_pairs: vec![
    //                 (String::from("AAA"), String::from("USD")),
    //                 (String::from("DNE1"), String::from("USD")),
    //                 (String::from("DNE2"), String::from("USD")),
    //             ],
    //         };
    //         let err = query(deps.as_ref(), env, msg).unwrap_err();
    //         assert_eq!(err, StdError::not_found("std_reference::state::RefData"));
    //
    //         // Test invalid symbols
    //         let env = mock_env();
    //         let msg = GetReferenceDataBulk {
    //             symbol_pairs: vec![
    //                 (String::from("AAA"), String::from("DNE1")),
    //                 (String::from("DNE2"), String::from("DNE2")),
    //                 (String::from("BBB"), String::from("DNE1")),
    //             ],
    //         };
    //         let err = query(deps.as_ref(), env, msg).unwrap_err();
    //         assert_eq!(err, StdError::not_found("std_reference::state::RefData"));
    //     }
    // }
}
