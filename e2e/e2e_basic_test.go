package e2e

import (
	"bufio"
	"context"
	"cosmossdk.io/math"
	"fmt"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdsdktypes "github.com/bnb-chain/greenfield/sdk/types"
	types2 "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

var (
	// Endpoint = "gnfd-testnet-fullnode-cosmos-us.nodereal.io:443"
	Endpoint = "http://localhost:26750"
	ChainID  = "greenfield_9000-121"
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
	cli, err := client.New(ChainID, Endpoint, client.Option{
		DefaultAccount: account,
	})
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
	cli, err := client.New(ChainID, Endpoint, client.Option{
		DefaultAccount: account,
	},
	)
	assert.NoError(t, err)
	ctx := context.Background()

	balance, err := cli.GetAccountBalance(ctx, account.GetAddress().String())
	assert.NoError(t, err)
	t.Logf("Balance: %s", balance.String())

	account1, _, err := types.NewAccount("test2")
	assert.NoError(t, err)
	transferTxHash, err := cli.Transfer(ctx, account1.GetAddress().String(), math.NewIntFromUint64(1), types2.TxOption{})
	assert.NoError(t, err)
	t.Logf("Transfer response: %s", transferTxHash)

	waitForTx, err := cli.WaitForTx(ctx, transferTxHash)
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

	txHash, err := cli.CreatePaymentAccount(ctx, account.GetAddress().String(), &types2.TxOption{})
	assert.NoError(t, err)
	t.Logf("Acc: %s", txHash)
	waitForTx, err = cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)
	t.Logf("Wair for tx: %s", waitForTx.String())

	paymentAccountsByOwner, err := cli.GetPaymentAccountsByOwner(ctx, account.GetAddress().String())
	assert.NoError(t, err)
	assert.Equal(t, len(paymentAccountsByOwner), 1)
}

func Test_MultiTransfer(t *testing.T) {
	mnemonic := ParseValidatorMnemonic(0)
	sender, err := types.NewAccountFromMnemonic("sender", mnemonic)
	assert.NoError(t, err)

	ctx := context.Background()
	cli, err := client.New(ChainID, Endpoint, client.Option{
		DefaultAccount: sender,
	})
	cli.SetDefaultAccount(sender)

	transferDetails := make([]types.TransferDetail, 0)
	totalSendAmount := math.NewInt(0)

	receiver1, _, err := types.NewAccount("receiver1")
	assert.NoError(t, err)
	receiver1Amount := math.NewInt(1000)
	totalSendAmount = totalSendAmount.Add(receiver1Amount)
	transferDetails = append(transferDetails, types.TransferDetail{
		ToAddress: receiver1.GetAddress().String(),
		Amount:    receiver1Amount,
	})

	receiver2, _, err := types.NewAccount("receiver2")
	assert.NoError(t, err)
	receiver2Amount := math.NewInt(1000)
	totalSendAmount = totalSendAmount.Add(receiver2Amount)
	transferDetails = append(transferDetails, types.TransferDetail{
		ToAddress: receiver2.GetAddress().String(),
		Amount:    receiver2Amount,
	})

	receiver3, _, err := types.NewAccount("receiver3")
	assert.NoError(t, err)
	receiver3Amount := math.NewInt(1000)
	totalSendAmount = totalSendAmount.Add(receiver3Amount)
	transferDetails = append(transferDetails, types.TransferDetail{
		ToAddress: receiver3.GetAddress().String(),
		Amount:    receiver3Amount,
	})

	t.Logf("totally sending %s", totalSendAmount.String())

	txHash, err := cli.MultiTransfer(ctx, transferDetails, gnfdsdktypes.TxOption{})
	assert.NoError(t, err)
	t.Log(txHash)

	_, err = cli.WaitForTx(ctx, txHash)
	assert.NoError(t, err)

	balance1, err := cli.GetAccountBalance(ctx, receiver1.GetAddress().String())
	assert.NoError(t, err)
	assert.Equal(t, receiver1Amount, balance1.Amount)

	balance2, err := cli.GetAccountBalance(ctx, receiver2.GetAddress().String())
	assert.NoError(t, err)
	assert.Equal(t, receiver2Amount, balance2.Amount)

	balance3, err := cli.GetAccountBalance(ctx, receiver3.GetAddress().String())
	assert.NoError(t, err)
	assert.Equal(t, receiver3Amount, balance3.Amount)

}
