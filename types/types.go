package types

import (
	"io"
	"math/rand"
	"net/url"
	"time"

	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Principal indicates the marshaled Principal content of greenfield permission types,
// user can generate it by NewPrincipalWithAccount or NewPrincipalWithGroupId method in utils
type Principal string

// ObjectStat contains the metadata of downloaded objects
type ObjectStat struct {
	ObjectName  string
	ContentType string
	Size        int64 // Object size
}

// ObjectInfo

type ObjectDetail struct {
	ObjectInfo         *storagetypes.ObjectInfo
	GlobalVirtualGroup *types.GlobalVirtualGroup
}

// QueryPieceInfo indicates the challenge or recovery object piece info
// RedundancyIndex if it is primary sp, the value should be -1，
// else it indicates the index of secondary sp
type QueryPieceInfo struct {
	ObjectId        string
	PieceIndex      int
	RedundancyIndex int
}

// ChallengeResult indicates the challenge hash and data results
type ChallengeResult struct {
	PieceData     io.ReadCloser
	IntegrityHash string
	PiecesHash    []string
}

func RandStr(n int) string {
	b := make([]rune, n)
	randMarker := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[randMarker.Intn(len(letters))]
	}
	return string(b)
}

type StorageProvider struct {
	Id              uint32
	OperatorAddress sdk.AccAddress
	FundingAddress  sdk.AccAddress
	SealAddress     sdk.AccAddress
	ApprovalAddress sdk.AccAddress
	GcAddress       sdk.AccAddress
	Status          spTypes.Status
	EndPoint        *url.URL
	Description     spTypes.Description
	BlsKey          []byte
}
