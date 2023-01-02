package orchestrator

import (
	wasmparams "github.com/CosmWasm/wasmd/app/params"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
)

func createMnemonic() (string, error) {
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropySeed)
	if err != nil {
		return "", err
	}

	return mnemonic, nil
}

func createMemoryKeyFromMnemonic(mnemonic string) (sdk.AccAddress, error) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("wasm", "wasm"+sdk.PrefixPublic)
	config.Seal()

	encodingConfig := wasmparams.MakeEncodingConfig()

	kb, err := keyring.New("testnet", keyring.BackendMemory, "", nil, encodingConfig.Marshaler)
	if err != nil {
		return nil, err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
	if err != nil {
		return nil, err
	}

	account, err := kb.NewAccount("", mnemonic, "", sdk.FullFundraiserPath, algo)
	if err != nil {
		return nil, err
	}
	return account.GetAddress()
}
