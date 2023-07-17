use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{to_binary, Binary, DepsMut, Env, Event, SubMsgResponse, SubMsgResult, Uint128, Uint256, Uint64, StdResult};

pub mod RequestRelay {
    use cosmwasm_schema::cw_serde;
    use cosmwasm_schema::schemars::JsonSchema;
    use cosmwasm_std::{StdResult, Uint64, Binary, to_binary};

    #[cw_serde]
    pub enum RequestType{
        RequestRate,
        RequestMedian,
        RequestDeviation,
    }

    #[cw_serde]
    pub struct RequestRateData {
        pub symbol: String,
        pub resolve_time: Uint64,
        pub callback_sig:String,
        pub callback_data: Binary,
    }

    #[cw_serde]
    pub struct RequestDeviationData {
        pub symbol: String,
        pub resolve_time: Uint64,
        pub callback_sig:String,
        pub callback_data: Binary,
    }

    #[cw_serde]
    pub struct RequestMedianData {
        pub symbol: String,
        pub resolve_time: Uint64,
        pub callback_sig:String,
        pub callback_data: Binary,
    }

    #[cw_serde]
    pub struct CallbackRateData {
        pub request_id: String,
        pub symbol: String,
        pub symbol_rate: Uint64,
        pub last_updated: Uint64,
        pub callback_data: Binary,
    }

    #[cw_serde]
    pub struct CallbackRateMedian {
        pub request_id: String,
        pub symbol: String,
        pub symbol_rates: Vec<Uint64>,
        pub last_updated: Uint64,
        pub callback_data: Binary,
    }

    #[cw_serde]
    pub struct CallbackRateDeviation {
        pub request_id: String,
        pub symbol: String,
        pub symbol_rate: Uint64,
        pub last_updated: Uint64,
        pub callback_data: Binary,
    }

    #[cw_serde]
    pub enum OracleRequestMsg {
        RequestRate(RequestRateData),
        RequestDeviation(RequestDeviationData),
        RequestMedian(RequestMedianData)
    }

    impl OracleRequestMsg{
        fn to_binary(&self) -> StdResult<Binary> {
            to_binary(self)
        }
    }

    impl RequestRateData {
        pub fn into_binary(self) -> StdResult<Binary> {
            let msg = OracleRequestMsg::RequestRate(self);
            to_binary(&msg)
        }
    }

    impl RequestDeviationData {
        pub fn into_binary(self) -> StdResult<Binary> {
            let msg = OracleRequestMsg::RequestDeviation(self);
            to_binary(&msg)
        }
    }

    impl RequestMedianData {
        pub fn into_binary(self) -> StdResult<Binary> {
            let msg = OracleRequestMsg::RequestMedian(self);
            to_binary(&msg)
        }
    }
}

pub mod helper {
    use crate::RequestRelay::{RequestDeviationData, RequestMedianData, RequestRateData, RequestType};
    use cosmwasm_std::{
        to_binary, Binary, CosmosMsg, DepsMut, Env, Event, Response, StdError, StdResult, SubMsg,
        SubMsgResponse, SubMsgResult, Uint128, Uint256, Uint64, WasmMsg,
    };

    pub fn oracle_submessage(
        oracle_address: String,
        symbol: String,
        resolve_time: Uint64,
        callback_data: Binary,
        success_id: u64,
        callback_sig: String,
        msg_type: RequestType,
    ) -> SubMsg {
        let payload:Binary;
        match msg_type {
            RequestType::RequestRate => {
                payload = RequestRateData {
                    symbol,
                    resolve_time,
                    callback_sig,
                    callback_data,
                }
                    .into_binary()
                    .unwrap();
            }
            RequestType::RequestMedian => {
                payload = RequestMedianData {
                     symbol,
                    resolve_time,
                    callback_sig,
                    callback_data,
                }
                    .into_binary()
                    .unwrap();
            }
            RequestType::RequestDeviation=>{
                payload = RequestDeviationData {
                    symbol,
                    resolve_time,
                    callback_sig,
                    callback_data,
                }
                    .into_binary()
                    .unwrap();
            }
        }


        let msg = SubMsg::reply_on_success(
            CosmosMsg::Wasm(WasmMsg::Execute {
                contract_addr: oracle_address,
                funds: vec![],
                msg: payload,
            }),
            success_id,
        );

        return msg;
    }

    pub fn oracle_request_id_from_reply(reply: &SubMsgResponse) -> StdResult<String> {
        let event = reply
            .events
            .iter()
            .find(|event| event_contains_attr(event, "action", "demand_price"))
            .ok_or_else(|| StdError::generic_err("cannot find demand price event"))?;

        let request_id = event
            .attributes
            .iter()
            .cloned()
            .find(|attr| attr.key == "request_id")
            .ok_or_else(|| StdError::generic_err("cannot find `request_id` attribute"))?
            .value;

        Ok(request_id)
    }

    fn event_contains_attr(event: &Event, key: &str, value: &str) -> bool {
        event
            .attributes
            .iter()
            .any(|attr| attr.key == key && attr.value == value)
    }
}

pub mod verify {
    use cosmwasm_schema::cw_serde;
    use cosmwasm_std::{
        to_binary, Deps, DepsMut, Env, MessageInfo, QueryRequest, StdResult, WasmQuery,
    };
    use std::fmt::Binary;
    #[cw_serde]
    pub enum QueryMsg {
        IsRelayer { relayer: String },
    }

    pub fn is_relayer(
        deps: &DepsMut,
        _: &Env,
        contract_address: String,
        sender: String,
    ) -> StdResult<bool> {
        let is_relayer_query_msg = QueryMsg::IsRelayer {
            relayer: sender.into(),
        };

        deps.querier.query_wasm_smart(contract_address,&is_relayer_query_msg)
    }
}

pub mod Error{
    use thiserror::Error;

    #[derive(Error, Debug, PartialEq)]
    pub enum Error {
        #[error("Invalid Relayer address: {relayer_address}")]
        InvalidRelayer {relayer_address:String},
    }
}