package bsctypes

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type FailureHandleStrategy uint8

const (
	BlockOnFail FailureHandleStrategy = iota
	CacheOnFail
	SkipOnFail
)

type UpdateGroupOpType uint8

const (
	AddMembers UpdateGroupOpType = iota
	RemoveMembers
	RenewMembers
)

type BucketVisibilityType uint8

const (
	Unspecified BucketVisibilityType = iota
	PublicRead
	Private
	Inherit
)

type ExtraData struct {
	AppAddress            *common.Address       `json:"appAddress"`
	RefundAddress         *common.Address       `json:"refundAddress"`
	FailureHandleStrategy FailureHandleStrategy `json:"failureHandleStrategy"`
	CallbackData          []byte                `json:"callbackData"`
}

type Deployment struct {
	DeployCommitId           string `json:"DeployCommitId"`
	BlockNumber              int    `json:"BlockNumber"`
	EmergencyOperator        string `json:"EmergencyOperator"`
	EmergencyUpgradeOperator string `json:"EmergencyUpgradeOperator"`
	Deployer                 string `json:"Deployer"`
	ProxyAdmin               string `json:"ProxyAdmin"`
	GovHub                   string `json:"GovHub"`
	CrossChain               string `json:"CrossChain"`
	MultiMessage             string `json:"MultiMessage"`
	GreenfieldExecutor       string `json:"GreenfieldExecutor"`
	TokenHub                 string `json:"TokenHub"`
	LightClient              string `json:"LightClient"`
	RelayerHub               string `json:"RelayerHub"`
	BucketHub                string `json:"BucketHub"`
	ObjectHub                string `json:"ObjectHub"`
	GroupHub                 string `json:"GroupHub"`
	AdditionalBucketHub      string `json:"AdditionalBucketHub"`
	AdditionalObjectHub      string `json:"AdditionalObjectHub"`
	AdditionalGroupHub       string `json:"AdditionalGroupHub"`
	BucketERC721Token        string `json:"BucketERC721Token"`
	ObjectERC721Token        string `json:"ObjectERC721Token"`
	GroupERC721Token         string `json:"GroupERC721Token"`
	MemberERC1155Token       string `json:"MemberERC1155Token"`
	InitConsensusState       struct {
		ChainID              string `json:"chainID"`
		Height               int    `json:"height"`
		NextValidatorSetHash string `json:"nextValidatorSetHash"`
		Validators           []struct {
			PubKey         string `json:"pubKey"`
			VotingPower    int    `json:"votingPower"`
			RelayerAddress string `json:"relayerAddress"`
			RelayerBlsKey  string `json:"relayerBlsKey"`
		} `json:"validators"`
		ConsensusStateBytes string `json:"consensusStateBytes"`
	} `json:"initConsensusState"`
	GnfdChainId             int    `json:"gnfdChainId"`
	PermissionDeployer      string `json:"PermissionDeployer"`
	PermissionHub           string `json:"PermissionHub"`
	AdditionalPermissionHub string `json:"AdditionalPermissionHub"`
	PermissionToken         string `json:"PermissionToken"`
}

type CreateBucketSynPackage struct {
	Creator                        *common.Address      `json:"creator"`
	Name                           string               `json:"name"`
	Visibility                     BucketVisibilityType `json:"visibility"`
	PaymentAddress                 *common.Address      `json:"paymentAddress"`
	PrimarySpAddress               *common.Address      `json:"primarySpAddress"`
	PrimarySpApprovalExpiredHeight uint64               `json:"primarySpApprovalExpiredHeight"`
	GlobalVirtualGroupFamilyId     uint32               `json:"globalVirtualGroupFamilyId"`
	PrimarySpSignature             []byte               `json:"primarySpSignature"`
	ChargedReadQuota               uint64               `json:"chargedReadQuota"`
	ExtraData                      []byte               `json:"extraData"`
}

type UpdateGroupMemberSynPackage struct {
	Operator         *common.Address   `json:"operator"`
	Id               *big.Int          `json:"Id"`
	OpType           UpdateGroupOpType `json:"opType"`
	Members          []common.Address  `json:"members"`
	ExtraData        []byte            `json:"extraData"`
	MemberExpiration []uint64          `json:"memberExpiration"`
}
