use cosmwasm_std::{Uint64,Binary};
use cosmwasm_schema::{cw_serde, QueryResponses};
use price_feed_helper::RequestRelay::{CallbackRateData,CallbackRateMedian,CallbackRateDeviation};

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
    GetRateRequestId{
        symbol: String,
    },

    #[returns(String)]
    GetMedianRequestId{
        symbol: String,
    },

    #[returns(String)]
    GetDeviationRequestId{
        symbol: String,
    },
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

    // execute msg for each request callback type
    CallbackRate(CallbackRateData),
    CallbackMedian(CallbackRateMedian),
    CallbackDeviation(CallbackRateDeviation),
}
