docker run --rm -v "$(pwd)/cosmwasm":/code \
--mount type=volume,source="$(basename "$(pwd)")_cache",target=/code/target \
--mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
-e RUST_BACKTRACE=full \
cosmwasm/workspace-optimizer:0.12.7
