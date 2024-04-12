package bsc

import (
	"context"
	"log"
	"math/big"
	"strings"

	"github.com/bnb-chain/greenfield-go-sdk/bsctypes"
	bsccommon "github.com/bnb-chain/greenfield-go-sdk/common"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type IMultiMessageClient interface {
	SendMessages(ctx context.Context, message *bsctypes.MultiMessage) (bool, error)
}

func (c *Client) SendMessages(ctx context.Context, message *bsctypes.MultiMessage) (bool, error) {
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
	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: packedData,
	}

	if sum != nil {
		msg.Value = sum
	}

	resp, err := c.chainClient.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Fatalf("Failed to call contract: %v", err)
	}

	result, err := parsedABI.Unpack("sendMessages", resp)
	if err != nil {
		log.Fatalf("Failed to unpack returned data: %v", err)
	}

	if len(result) > 0 {
		if value, ok := result[0].(bool); ok {
			return value, nil
		} else {
			return false, err
		}
	}

	return false, nil
}
