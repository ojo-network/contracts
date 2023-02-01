package main

type Network struct {
	ChainID               string
	LocalRelayerBinary    string
	RelayerHomeFolderName string
	NodeConfig            NodeConfig
}

type NodeLocation struct {
	Region string
	Zone   string
}

type NodeConfig struct {
	MachineType string
	DiskType    string
	DiskSizeGB  int
	Location    NodeLocation
}

type NodeSecretConfig struct {
}
