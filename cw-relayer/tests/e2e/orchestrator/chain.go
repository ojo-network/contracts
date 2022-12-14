package orchestrator

// Chain defines structure that contains chain id for wasm chain, validator mnemonic and address.
type Chain struct {
	chainId      string
	val_mnemonic string
	address      string
}

// NewChain returns instance of Chain with set chain id, a random validator mnemonic and address.
func NewChain(chainId string) *Chain {
	mnemonic, _ := createMnemonic()

	address, err := createMemoryKeyFromMnemonic(mnemonic)
	if err != nil {
		panic(err)
	}

	return &Chain{
		chainId:      chainId,
		val_mnemonic: mnemonic,
		address:      address.String(),
	}
}
