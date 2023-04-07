package e2e

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	types2 "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// GrpcAddress = "gnfd-testnet-fullnode-cosmos-us.nodereal.io:443"
	GrpcAddress = "localhost:9090"
	ChainID     = "greenfield_9000-121"
)

// ParseValidatorMnemonic read the validator mnemonic from file
func ParseValidatorMnemonic(i int) string {
	return ParseMnemonicFromFile(fmt.Sprintf("../../greenfield/deployment/localup/.local/validator%d/info", i))
}

func ParseMnemonicFromFile(fileName string) string {
	fileName = filepath.Clean(fileName)
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	// #nosec
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	var line string
	for scanner.Scan() {
		if scanner.Text() != "" {
			line = scanner.Text()
		}
	}
	return line
}

func Test_Basic(t *testing.T) {
	mnemonic := ParseValidatorMnemonic(0)
	account, err := types.NewAccountFromMnemonic("test", mnemonic)
	assert.NoError(t, err)
	cli, err := client.New(ChainID, GrpcAddress, account, client.Option{GrpcDialOption: grpc.WithTransportCredentials(insecure.NewCredentials())})
	assert.NoError(t, err)
	ctx := context.Background()
	_, _, err = cli.GetNodeInfo(ctx)
	assert.NoError(t, err)

	latestBlock, err := cli.GetLatestBlock(ctx)
	assert.NoError(t, err)
	fmt.Println(latestBlock.String())

	heightBefore := latestBlock.Header.Height
	err = cli.WaitForBlockHeight(ctx, heightBefore+10)
	assert.NoError(t, err)
	height, err := cli.GetLatestBlockHeight(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, height, heightBefore+10)

	syncing, err := cli.GetSyncing(ctx)
	assert.NoError(t, err)
	assert.False(t, syncing)

	blockByHeight, err := cli.GetBlockByHeight(ctx, heightBefore)
	assert.NoError(t, err)
	assert.Equal(t, blockByHeight.GetHeader(), latestBlock.GetHeader())
}

func Test_Account(t *testing.T) {
	mnemonic := ParseValidatorMnemonic(0)
	account, err := types.NewAccountFromMnemonic("test", mnemonic)
	assert.NoError(t, err)
	cli, err := client.New(ChainID, GrpcAddress, account, client.Option{GrpcDialOption: grpc.WithTransportCredentials(insecure.NewCredentials())})
	assert.NoError(t, err)
	ctx := context.Background()

	balance, err := cli.GetAccountBalance(ctx, account.GetAddress().String())
	assert.NoError(t, err)
	t.Logf("Balance: %s", balance.String())

	account1, err := types.NewAccount("test2")
	assert.NoError(t, err)
	transfer, err := cli.Transfer(ctx, account1.GetAddress().String(), 1, nil)
	assert.NoError(t, err)
	t.Logf("Transfer response: %s", transfer.String())

	waitForTx, err := cli.WaitForTx(ctx, transfer.TxHash)
	assert.NoError(t, err)
	t.Logf("Wair for tx: %s", waitForTx.String())

	balance, err = cli.GetAccountBalance(ctx, account1.GetAddress().String())
	assert.NoError(t, err)
	t.Logf("Balance: %s", balance.String())
	assert.True(t, balance.Amount.Equal(math.NewInt(1)))

	acc, err := cli.GetAccount(ctx, account1.GetAddress().String())
	assert.NoError(t, err)
	t.Logf("Acc: %s", acc.String())
	assert.Equal(t, acc.GetAddress(), account1.GetAddress())
	assert.Equal(t, acc.GetSequence(), uint64(0))

	txResp, err := cli.CreatePaymentAccount(ctx, account.GetAddress().String(), &types2.TxOption{})
	assert.NoError(t, err)
	t.Logf("Acc: %s", txResp.String())
	waitForTx, err = cli.WaitForTx(ctx, txResp.TxHash)
	assert.NoError(t, err)
	t.Logf("Wair for tx: %s", waitForTx.String())

	paymentAccountsByOwner, err := cli.GetPaymentAccountsByOwner(ctx, account.GetAddress().String())
	assert.NoError(t, err)
	assert.Equal(t, len(paymentAccountsByOwner), 1)
}
