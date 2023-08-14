package e2e

import (
	"github.com/bnb-chain/greenfield-go-sdk/e2e/basesuite"
	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
)

type BucketMigrateTestSuite struct {
	basesuite.BaseSuite
	PrimarySP spTypes.StorageProvider
}

//func (s *BucketMigrateTestSuite) SetupSuite() {
//	s.BaseSuite.SetupSuite()
//
//	spList, err := s.Client.ListStorageProviders(s.ClientContext, false)
//	s.Require().NoError(err)
//	for _, sp := range spList {
//		if sp.Endpoint != "https://sp0.greenfield.io" {
//			s.PrimarySP = sp
//			break
//		}
//	}
//}

//func TestBucketMigrateTestSuiteTestSuite(t *testing.T) {
//	suite.Run(t, new(BucketMigrateTestSuite))
//}
