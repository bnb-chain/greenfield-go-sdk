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

type IBasicClient interface {
	GetChainID() int
	GetDeployment() *bsctypes.Deployment
	GetMinAckRelayFee() (relayFee *big.Int, minAckRelayFee *big.Int, err error)
}

func (c *Client) GetChainID() int {
	return c.chainID
}

func (c *Client) GetDeployment() *bsctypes.Deployment {
	return c.deployment
}

func (c *Client) GetMinAckRelayFee() (relayFee *big.Int, minAckRelayFee *big.Int, err error) {
	var (
		ok1 bool
		ok2 bool
	)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.MultiMessageABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("getRelayFees")
	if err != nil {
		log.Fatalf("Failed to pack data for sendMessages: %v", err)
	}

	contractAddress := common.HexToAddress(c.GetDeployment().CrossChain)
	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: packedData,
	}

	resp, err := c.chainClient.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Fatalf("Failed to call contract: %v", err)
	}

	result, err := parsedABI.Unpack("getRelayFees", resp)
	if err != nil {
		log.Fatalf("Failed to unpack returned data: %v", err)
	}

	if len(result) != 2 {
		log.Fatalf("Expected two return values from getRelayFees")
	}

	relayFee, ok1 = result[0].(*big.Int)
	minAckRelayFee, ok2 = result[1].(*big.Int)
	if !ok1 || !ok2 {
		log.Fatalf("Type assertion failed for one or both return values")
	}

	return relayFee, minAckRelayFee, nil
}
