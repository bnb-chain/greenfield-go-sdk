package bsc

import (
	"context"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/bnb-chain/greenfield-go-sdk/bsctypes"
	bsccommon "github.com/bnb-chain/greenfield-go-sdk/common"
)

type IBasicClient interface {
	SendTx(ctx context.Context, nonce uint64, toAddr *common.Address, amount *big.Int, gasPrice *big.Int, data []byte) (*common.Hash, error)
	GetDeployment() *bsctypes.Deployment
	GetMinAckRelayFee(ctx context.Context) (relayFee *big.Int, minAckRelayFee *big.Int, err error)
	GetCallbackGasPrice(ctx context.Context) (gasPrice *big.Int, err error)
	CheckTxStatus(ctx context.Context, tx *common.Hash) (bool, error)
}

func (c *Client) GetDeployment() *bsctypes.Deployment {
	return c.deployment
}

func (c *Client) SendTx(ctx context.Context, nonce uint64, toAddr *common.Address, amount *big.Int, gasPrice *big.Int, data []byte) (*common.Hash, error) {
	if nonce == 0 {
		n, err := c.chainClient.PendingNonceAt(ctx, *c.defaultAccount.GetAddress())
		if err != nil {
			return nil, err
		}
		nonce = n
	}
	gasLimit := uint64(5e6) // Assuming the gas limit is static, adjust as necessary
	var err error
	if gasPrice == nil {
		gasPrice, err = c.chainClient.SuggestGasPrice(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Create the transaction using NewTx instead of the deprecated NewTransaction
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       toAddr,
		Value:    amount,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	})

	chainId, err := c.chainClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainId), c.defaultAccount.GetKeyManager().GetPrivateKey())
	if err != nil {
		return nil, err
	}
	err = c.chainClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, err
	}
	hash := signedTx.Hash()
	return &hash, nil
}

func (c *Client) GetMinAckRelayFee(ctx context.Context) (relayFee *big.Int, minAckRelayFee *big.Int, err error) {
	var (
		ok1 bool
		ok2 bool
	)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.CrossChainABI))
	if err != nil {
		log.Fatalf("failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("getRelayFees")
	if err != nil {
		log.Fatalf("failed to pack data for getRelayFees: %v", err)
	}

	contractAddress := common.HexToAddress(c.GetDeployment().CrossChain)
	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: packedData,
	}

	resp, err := c.chainClient.CallContract(ctx, msg, nil)
	if err != nil {
		log.Fatalf("failed to call contract: %v", err)
	}

	result, err := parsedABI.Unpack("getRelayFees", resp)
	if err != nil {
		log.Fatalf("failed to unpack returned data: %v", err)
	}

	if len(result) != 2 {
		log.Fatalf("expected two return values from getRelayFees")
	}

	relayFee, ok1 = result[0].(*big.Int)
	minAckRelayFee, ok2 = result[1].(*big.Int)
	if !ok1 || !ok2 {
		log.Fatalf("type assertion failed for one or both return values")
	}

	return relayFee, minAckRelayFee, nil
}

func (c *Client) GetCallbackGasPrice(ctx context.Context) (gasPrice *big.Int, err error) {
	var (
		ok bool
	)
	parsedABI, err := abi.JSON(strings.NewReader(bsccommon.CrossChainABI))
	if err != nil {
		log.Fatalf("failed to parse contract ABI: %v", err)
	}

	packedData, err := parsedABI.Pack("callbackGasPrice")
	if err != nil {
		log.Fatalf("failed to pack data for callbackGasPrice: %v", err)
	}

	contractAddress := common.HexToAddress(c.GetDeployment().CrossChain)
	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: packedData,
	}

	resp, err := c.chainClient.CallContract(ctx, msg, nil)
	if err != nil {
		log.Fatalf("failed to call contract: %v", err)
	}

	result, err := parsedABI.Unpack("callbackGasPrice", resp)
	if err != nil {
		log.Fatalf("failed to unpack returned data: %v", err)
	}

	if len(result) != 1 {
		log.Fatalf("expected one return values from callbackGasPrice")
	}

	gasPrice, ok = result[0].(*big.Int)
	if !ok {
		log.Fatalf("type assertion failed for one or both return values")
	}

	return gasPrice, nil
}

func (c *Client) CheckTxStatus(ctx context.Context, tx *common.Hash) (bool, error) {
	var success bool

	receipt, err := c.chainClient.TransactionReceipt(ctx, *tx)
	if err != nil {
		log.Fatal(err)
	}
	if receipt.Status == 1 {
		success = true
	}

	return success, nil
}
