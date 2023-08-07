package client

import (
	"github.com/CosmWasm/wasmd/app"
	wasmparams "github.com/CosmWasm/wasmd/app/params"
	"github.com/cosmos/cosmos-sdk/std"
)

// MakeEncodingConfig encoding config to support wasm msgs
func MakeEncodingConfig() wasmparams.EncodingConfig {
	encodingConfig := wasmparams.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	app.ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	app.ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
