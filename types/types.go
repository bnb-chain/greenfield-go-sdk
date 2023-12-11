package types

import (
	"io"
	"math/rand"
	"net/url"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Principal indicates the marshaled Principal content of greenfield permission types,
// user can generate it by NewPrincipalWithAccount or NewPrincipalWithGroupId method in utils.
type Principal string

// ObjectStat contains the metadata of the downloaded object.
type ObjectStat struct {
	ObjectName  string
	ContentType string
	Size        int64 // Object size
}

// ObjectDetail contains the detailed info of the object stored on Greenfield.
type ObjectDetail struct {
	ObjectInfo         *storagetypes.ObjectInfo  `protobuf:"bytes,1,opt,name=object_info" json:"object_info,omitempty"`
	GlobalVirtualGroup *types.GlobalVirtualGroup `protobuf:"bytes,2,opt,name=global_virtual_group" json:"global_virtual_group,omitempty"`
}

// QueryPieceInfo indicates the challenge or recovery object piece info.
// If it is primary sp, the RedundancyIndex value should be -1， else it indicates the index of secondary sp.
type QueryPieceInfo struct {
	ObjectId        string
	PieceIndex      int
	RedundancyIndex int
}

// ChallengeResult includes the integrity hash, data results and hashes for storage provide to respond to challenges.
type ChallengeResult struct {
	IntegrityHash string        // the integrity hash of the challenged object
	PieceData     io.ReadCloser // the data of the segment/piece being challenged
	PiecesHash    []string      // the hashes of the object's segments/pieces
}

// RandStr - Generate a random string for test usage.
func RandStr(n int) string {
	b := make([]rune, n)
	randMarker := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[randMarker.Intn(len(letters))]
	}
	return string(b)
}

// StorageProvider indicates the metadata of SP that stored on-chain.
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
