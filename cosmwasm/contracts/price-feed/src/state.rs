use cosmwasm_std::{Addr, Uint64};
use cw_controllers::Admin;
use cw_storage_plus::{Item, Map};

// Administrator account
pub const ADMIN: Admin = Admin::new("admin");

// Used to store addresses of relayers and their state
pub const RELAYERS: Map<&Addr, bool> = Map::new("relayers");

// Used to store Median Status
pub const MEDIANSTATUS: Item<bool> = Item::new("medianstatus");

pub const TRIGGER_REQUEST: Item<bool> = Item::new("triggerequest");

pub const PINGCHECK: Map<&Addr, Uint64> = Map::new("pingcheck");

pub const LAST_RELAYER: Item<String> = Item::new("last_relayer");

pub const PING_THRESHOLD: Item<Uint64> = Item::new("ping_threshold");
