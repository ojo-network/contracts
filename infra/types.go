package main

type Network struct {
	ChainID            string
	LocalRelayerBinary string
	LocalContractTar   string
	UserAddress        string
	ContractAddress    string
	NodeConfig         NodeConfig
}

type NodeConfig struct {
	InstanceName string
	InstanceID   string
}

type NodeSecretConfig struct {
}
