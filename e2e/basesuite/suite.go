package basesuite

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	"github.com/stretchr/testify/suite"
)

var (
	//Endpoint = "http://localhost:26750"
	//ChainID  = "greenfield_9000-121"

	//Endpoint = "https://gnfd-dev.qa.bnbchain.world:443"
	//ChainID  = "greenfield_8981-1"

	Endpoint = "https://gnfd.qa.bnbchain.world:443"
	ChainID  = "greenfield_9000-1741"
)

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

type BaseSuite struct {
	suite.Suite
	DefaultAccount *types.Account
	Client         client.Client
	ClientContext  context.Context
}

// ParseValidatorMnemonic read the validator mnemonic from file
func ParseValidatorMnemonic(i int) string {
	return ParseMnemonicFromFile(fmt.Sprintf("/Users/fynn/Workspace/greenfield-local-deploy-script/greenfield/deployment/localup/.local/validator%d/info", i))
}

func (s *BaseSuite) SetupSuite() {
	//mnemonic := ParseValidatorMnemonic(0)
	//account, err := types.NewAccountFromMnemonic("test", mnemonic)
	account, err := types.NewAccountFromPrivateKey("test", "d128099d8b107cd1119287cc881dcaceca73469e116a33098102d50286de3353")
	s.Require().NoError(err)
	s.Client, err = client.New(ChainID, Endpoint, client.Option{
		DefaultAccount: account,
	})
	s.Require().NoError(err)
	s.ClientContext = context.Background()
	s.DefaultAccount = account
}
