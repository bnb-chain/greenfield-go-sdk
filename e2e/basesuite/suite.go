package basesuite

import (
	"bufio"
	"context"
	"fmt"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	"os"
	"path/filepath"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	"github.com/stretchr/testify/suite"
)

var (
	// Endpoint = "http://localhost:26750"
	// ChainID  = "greenfield_9000-121"
	//Endpoint = "https://gnfd-dev.qa.bnbchain.world:443"
	//ChainID  = "greenfield_8981-1"

	Endpoint = "https://gnfd-testnet-fullnode-tendermint-ap.bnbchain.org:443"
	ChainID  = "greenfield_5600-1"
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
	DefaultAccount  *types.Account
	Client          client.Client
	ClientContext   context.Context
	ChallengeClient client.Client
}

// ParseValidatorMnemonic read the validator mnemonic from file
func ParseValidatorMnemonic(i int) string {
	return ParseMnemonicFromFile(fmt.Sprintf("../../greenfield/deployment/localup/.local/validator%d/info", i))
}

func (s *BaseSuite) NewChallengeClient() {
	mnemonic := ParseMnemonicFromFile(fmt.Sprintf("../../greenfield/deployment/localup/.local/challenger%d/challenger_info", 0))
	challengeAcc, err := types.NewAccountFromMnemonic("challenge_account", mnemonic)
	s.Require().NoError(err)
	s.ChallengeClient, err = client.New(ChainID, Endpoint, client.Option{
		DefaultAccount: challengeAcc,
	})
	s.Require().NoError(err)
}

func (s *BaseSuite) SetupSuite() {
	//mnemonic := ParseValidatorMnemonic(0)
	//account, err := types.NewAccountFromMnemonic("test", mnemonic)
	account, err := types.NewAccountFromPrivateKey("test", "92b4cdd49090cab5f3b6bf334021486ac3d3de865e02d1bced72f95f933c15e8")
	s.Require().NoError(err)
	s.Client, err = client.New(ChainID, Endpoint, client.Option{
		DefaultAccount: account,
	})
	s.Require().NoError(err)
	s.ClientContext = context.Background()
	s.DefaultAccount = account
	s.NewChallengeClient()
}

func (s *BaseSuite) WaitSealObject(bucketName string, objectName string) {
	startCheckTime := time.Now()
	var (
		objectDetail *types.ObjectDetail
		err          error
	)

	// wait 300s
	for i := 0; i < 100; i++ {
		objectDetail, err = s.Client.HeadObject(s.ClientContext, bucketName, objectName)
		s.Require().NoError(err)
		if objectDetail.ObjectInfo.GetObjectStatus() == storageTypes.OBJECT_STATUS_SEALED {
			break
		}
		time.Sleep(3 * time.Second)
	}

	s.Require().Equal(objectDetail.ObjectInfo.GetObjectStatus().String(), "OBJECT_STATUS_SEALED")
	s.T().Logf("---> Wait Seal Object cost %d ms, <---", time.Since(startCheckTime).Milliseconds())
}
