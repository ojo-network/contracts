package tools

func StringToByte32(asset string) [32]byte {
	var byteArray [32]byte
	copy(byteArray[:], asset)

	return byteArray
}
