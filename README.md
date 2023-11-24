# Ojo's Price Feed Contracts

This monorepo is intended to contain Ojo's Price Feeding Contracts. It will be organized by:

- CosmWasm Contracts
- Golang-based Relayer

The Contracts in this repo are a fork of Band Protocol's Cosmwasm Contracts, with a few key changes outlined in the [Readme](./cosmwasm/README.md).

The relayer will be a golang-based implementation to upload pricing information into the standard cosmwasm contract.

- In case of relayer restart modify median request id and request id in config.toml, according to the current request id

## Contract Deployments

| Chain             | Contract Address                                                   | Admin Address                               | Relayer Address                                | Relayer Release Binary          |
| -----------------:| ------------------------------------------------------------------:| -------------------------------------------:| ----------------------------------------------:| -------------------------------:|
| Archway Mainnet   | archway1c4nn5dl24fls0gspcuakuk8qtgpsmygxvq3jf5ua9tfxe78vg60s6jzndv |                                             | archway1ner6kc63xl903wrv2n8p9mtun79gegjl6wtgen | cw-relayer-v0.1.7-alpha2        |
| Comdex Mainnet    | comdex1d9w3v30r26xuckatahmeelut92jl6uc8suzmlpp0dm08drwlskjscvg83a  |                                             | comdex1nmps9c8j4zfzu572nv9fdlr3y40wm4ae88s8wy  | cw-relayer-v0.1.7-alpha1        |
| Comdex Testnet    | comdex1wn625s4jcmvk0szpl85rj5azkfc6suyvf75q6vrddscjdphtve8s6z5fs0  |                                             |                                                |                                 |
| Injective Mainnet | inj1m6lerugeyrfp78eetrtw04ykleu3vsx686dw9y                         | inj1cx44k04dp4rpdlkpdc532d6qcc03j2h6f2g78y  | inj1cx44k04dp4rpdlkpdc532d6qcc03j2h6f2g78y     |                                 |
| Injective Testnet | inj1v0h36f859f7uhqn5l64uccn7gsg60n4eteux2u                         | inj1u6c4qcjdlzg8c7455h0mrpftcps8lpqayrxept  | inj1u6c4qcjdlzg8c7455h0mrpftcps8lpqayrxept     |                                 |
| Juno Mainnet      | juno1yqm8q56hjv8sd4r37wdhkkdt3wu45gc5ptrjmd9k0nhvavl0354qwcf249    |                                             | juno1rkhrfuq7k2k68k0hctrmv8efyxul6tgn8hny6y    | cw-relayer-v0.1.5-alpha4        |
| Juno Testnet      | juno1v7na2m53cm3rkxxpe0k62yra3nkpeq7dy5zr528795fgymxrlz2ql4jsqp    |                                             |                                                |                                 |
| Neutron Mainnet   | neutron123lcmhhl79x9ghf2dallkvvxuqwl7fmp9g6avjzt2zyrvddme7eqsvtr07 |                                             | neutron1xhmxhxekescrqrv54dp30vx3ngcc97pe9mq2k4 | cw-relayer-v0.1.7-alpha1        |
| Neutron Testnet   | neutron1y3jqjzwv6x369xqr67d2ry05kk8l3ajc56d49tprjfpfvl5z0d0s5zcdr5 |                                             | neutron1xhmxhxekescrqrv54dp30vx3ngcc97pe9mq2k4 | cw-relayer-v0.1.5               |
| Osmosis Mainnet   | osmo1996xgzduvstz7j3ut0r0mursag58hg25269r5lhswk7awkd52tjset60yn    | osmo1z0fzccr6g6l957s5xe5h8p46qgrd9k3uqf98u4 | osmo1z3v35nvrhj70xx45708k2fg6cmy8ypng5rzcfn    | sdk47-v1.7.0                    |
| Osmosis Testnet   | osmo1amnnzxadxe5r4g8tezywea78fycuef68vlttneyvz0g7h9js9yxsmzk6cw    | osmo1z0fzccr6g6l957s5xe5h8p46qgrd9k3uqf98u4 | osmo1z0fzccr6g6l957s5xe5h8p46qgrd9k3uqf98u4    | sdk47-v1.7.0                    |
| Secret Mainnet    | secret179haa5uxgmsanuna2dapquwz3twylys394gz8e                      |                                             | secret1cm4wctlnkaszv7s5ccxuw50wuj3lg8lqn567sz  | cw-relayer-secret-v0.0.4-alpha1 |
| Secret Testnet    | secret1u6mtm6yk9z95w4d34dzj443h495yt3zqn8ny3p                      |                                             |                                                |                                 |
| Sei Testnet       | sei1gjava555e8k566y5nxkyfv2rjcmwknhp2ew8fvcqk3d25f7n3sds0mw434     | sei1kj8ndvywaeq42rw2r42v6ewnq23sewyh88m2ae  | sei1kj8ndvywaeq42rw2r42v6ewnq23sewyh88m2ae     |                                 |
| Stargaze Mainnet  | stars1e4fx86gr7l7wu64kz3tay6fqzz4mmml8r0welpfexeyuav0ztm3q0j5ulq   |                                             |                                                |                                 |
| Stargaze Testnet  | stars1yrm90l6z89eldsfhduxaygjq0rstqcuu3uu3slpq83c2lldqf6jsfy8py2   |                                             | stars1ksxvztdwplktkugn0kqxlttvqhrhlspxth8qh4   | cw-relayer-v0.1.7-alpha2        |
