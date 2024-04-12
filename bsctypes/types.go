package bsctypes

type FailureHandleStrategy int

const (
	BlockOnFail FailureHandleStrategy = iota
	CacheOnFail
	SkipOnFail
)

type ExtraData struct {
	AppAddress            string                `json:"appAddress"`
	RefundAddress         string                `json:"refundAddress"`
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
