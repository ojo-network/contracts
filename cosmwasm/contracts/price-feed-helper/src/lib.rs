use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{to_binary, Binary, DepsMut, Env, Event, SubMsgResponse, SubMsgResult, Uint128, Uint256, Uint64, StdResult};

pub mod RequestRelay {
    use cosmwasm_schema::cw_serde;
    use cosmwasm_std::{StdResult,Uint64,Binary};

    #[cw_serde]
    pub struct RequestRateData {
        pub symbol: String,
        pub resolve_time: Uint64,
        pub callback_data: Binary,
    }

    #[cw_serde]
    pub struct RequestDeviationData{
        pub symbol: String,
        pub resolve_time: Uint64,
        pub callback_data: Binary,
    }

    #[cw_serde]
    pub struct RequestMedianData{
        pub symbol: String,
        pub resolve_time: Uint64,
        pub callback_data:Binary,
    }
}

    #[cw_serde]
    pub struct RateDataCallback {
        pub request_id: String,
        pub symbol: String,
        pub symbol_rate: Uint64,
        pub last_updated: Uint64,
        pub callback_data: Binary,
    }

    #[cw_serde]
    pub struct RateMedianCallback {
        pub request_id: String,
        pub symbol: String,
        pub symbol_rate: Vec<Uint64>,
        pub last_updated: Uint64,
        pub callback_data: Binary,
    }

    #[cw_serde]
    pub struct RateDeviationCallback {
        pub request_id: String,
        pub symbol: String,
        pub symbol_rate: Uint64,
        pub last_updated: Uint64,
        pub callback_data: Binary,
    }

    #[cw_serde]
    pub enum OracleMsg {
        RequestRelay(RequestRelayData),
        Callback(CallbackData),
    }

    impl RequestRelayData {
        pub fn into_binary(self) -> StdResult<Binary> {
            let msg = OracleMsg::RequestRelay(self);
            to_binary(&msg)
        }
    }
}

pub mod helper {
    use crate::RequestRelay::RequestRelayData;
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
    ) -> SubMsg {
        let payload = RequestRelayData {
            symbol: symbol,
            resolve_time,
            callback_data,
        }
        .into_binary()
        .unwrap();

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

    pub fn request_id_from_reply(reply: &SubMsgResponse) -> StdResult<String> {
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

    pub fn verify_relayer(
        deps: &DepsMut,
        env: &Env,
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