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
	Location     NodeLocation
}

type NodeLocation struct {
	Region string
	Zone   string
}

type NodeSecretConfig struct {
}
