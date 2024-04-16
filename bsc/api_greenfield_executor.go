package bsc

import (
	"context"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/bnb-chain/greenfield-go-sdk/bsctypes"
	bsccommon "github.com/bnb-chain/greenfield-go-sdk/common"
)

type IGreenfieldExecutorClient interface {
	Execute(ctx context.Context, message *bsctypes.ExecutorMessages) (*common.Hash, error)
}

func (c *Client) Execute(ctx context.Context, message *bsctypes.ExecutorMessages) (*common.Hash, error) {
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.ExecutorABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("execute", message.MsgTypes, message.MsgBytes)
	if err != nil {
		log.Fatalf("Failed to pack data for Execute: %v", err)
	}

	contractAddress := common.HexToAddress(c.GetDeployment().GreenfieldExecutor)
	tx, err := c.SendTx(ctx, 0, &contractAddress, message.RelayFee, nil, packedData)
	if err != nil {
		log.Fatalf("Failed to call contract: %v", err)
	}

	return tx, nil
}
