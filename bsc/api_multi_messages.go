package bsc

import (
	"context"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/bnb-chain/greenfield-go-sdk/bsctypes"
	bsccommon "github.com/bnb-chain/greenfield-go-sdk/common"
)

type IMultiMessageClient interface {
	SendMessages(ctx context.Context, message *bsctypes.MultiMessage) (*common.Hash, error)
}

func (c *Client) SendMessages(ctx context.Context, message *bsctypes.MultiMessage) (*common.Hash, error) {
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.MultiMessageABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("sendMessages", message.Targets, message.Data, message.Values)
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	sum := new(big.Int)
	for _, value := range message.Values {
		sum.Add(sum, value)
	}

	contractAddress := common.HexToAddress(c.GetDeployment().MultiMessage)
	tx, err := c.SendTx(ctx, 0, &contractAddress, sum, nil, packedData)
	if err != nil {
		log.Fatalf("Failed to call contract: %v", err)
	}

	return tx, nil
}
