package util

import (
	"github.com/bnb-chain/gnfd-go-sdk/common"
	"github.com/bnb-chain/gnfd-go-sdk/types"
)

func ValidateAmount(amount int64) error {
	//TODO check if enough balance
	return nil
}

func ValidateToken(token string) error {
	for _, c := range common.SupportedCoins {
		if c == token {
			return nil
		}
	}
	return types.TokenNotSupportError
}
