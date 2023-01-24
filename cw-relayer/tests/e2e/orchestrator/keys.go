package orchestrator

//func createMnemonic() (string, error) {
//	entropySeed, err := bip39.NewEntropy(256)
//	if err != nil {
//		return "", err
//	}
//
//	mnemonic, err := bip39.NewMnemonic(entropySeed)
//	if err != nil {
//		return "", err
//	}
//
//	return mnemonic, nil
//}
//
//func createMemoryKeyFromMnemonic(mnemonic string) (sdk.AccAddress, error) {
//	config := sdk.GetConfig()
//	config.SetBech32PrefixForAccount("wasm", "wasm"+sdk.PrefixPublic)
//	config.Seal()
//
//	kb, err := keyring.New("testnet", keyring.BackendMemory, "", nil)
//	if err != nil {
//		return nil, err
//	}
//
//	keyringAlgos, _ := kb.SupportedAlgorithms()
//	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), keyringAlgos)
//	if err != nil {
//		return nil, err
//	}
//
//	account, err := kb.NewAccount("", mnemonic, "", sdk.FullFundraiserPath, algo)
//	if err != nil {
//		return nil, err
//	}
//
//	return account.GetAddress(), nil
//}
