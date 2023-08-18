use cosmwasm_std::StdError;
// use cw_controllers::AdminError;
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized: sender is not an admin")]
    UnauthorizedAdmin {},

    #[error("Unauthorized: sender is not relayer")]
    UnauthorizedRelayer {},

    #[error("Median is disabled")]
    MedianDisabled {},
}
