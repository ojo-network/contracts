package client

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	scrttypes "github.com/ojo-network/cw-relayer/relayer/dep/utils/types"
)

func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeTestEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	// register accountI
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)

	scrttypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	scrttypes.RegisterLegacyAminoCodec(encodingConfig.Amino)

	return encodingConfig
}
