package types

var ChainId string

func SetChainId(id string) {
	// todo validate chain id from user input
	ChainId = id
}
