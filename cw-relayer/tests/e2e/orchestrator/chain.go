package orchestrator

type Chain struct {
	chainId      string
	val_mnemonic string
	address      string
}

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
