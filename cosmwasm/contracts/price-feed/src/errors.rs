use cosmwasm_std::StdError;
// use cw_controllers::AdminError;
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized: {msg}")]
    Admin { msg: String },

    #[error("Unauthorized: {msg}")]
    Unauthorized { msg: String },
}
