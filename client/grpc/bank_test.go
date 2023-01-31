package client

import (
	"github.com/bnb-chain/gnfd-go-sdk/keys"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSendToken(t *testing.T) {
	km, err := keys.NewPrivateKeyManager("ab463aca3d2965233da3d1d6108aa521274c5ddc2369ff72970a52a451863fbf")
	assert.NoError(t, err)
	gnfdCli := NewGreenlandClientWithKeyManager("localhost:9090", "greenfield_9000-121", km)
	_, err = gnfdCli.SendToken("bnb", "0x76d244CE05c3De4BbC6fDd7F56379B145709ade9", 10, true)
	assert.NoError(t, err)
}
