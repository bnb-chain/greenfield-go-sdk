package e2e

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	RpcAddress  = "http://localhost:26750"
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
	cli, err := client.New(ChainID, GrpcAddress, account, &client.Option{GrpcDialOption: grpc.WithTransportCredentials(insecure.NewCredentials())})
	assert.NoError(t, err)
	ctx := context.Background()
	_, _, err = cli.GetNodeInfo(ctx)
	assert.NoError(t, err)

	latestBlock, err := cli.LatestBlock(ctx)
	fmt.Println(latestBlock.String())

	heightBefore := latestBlock.Header.Height
	err = cli.WaitForBlockHeight(ctx, heightBefore+10)
	assert.NoError(t, err)
	height, err := cli.LatestBlockHeight(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, height, heightBefore+10)
}
