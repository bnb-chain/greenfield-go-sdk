package types

import (
	"io"
)

// ObjectInfo contains the metadata of downloaded objects
type ObjectStat struct {
	ObjectName  string
	ContentType string
	Size        int64
}

// ChallengeInfo indicates the challenge object info
// RedundancyIndex if it is primary sp, the value should be -1，
// else it indicates the index of secondary sp
type ChallengeInfo struct {
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

// AuthInfo is the authorization info of requests
type AuthInfo struct {
	SignType      string // if using wallet sign, set authV2
	WalletSignStr string
}

// NewAuthInfo returns the AuthInfo which need to pass to api
// useWalletSign indicates whether you need use wallet to sign
// signStr indicates the wallet signature or jwt token
func NewAuthInfo(useWalletSign bool, signStr string) AuthInfo {
	if !useWalletSign {
		return AuthInfo{
			SignType:      AuthV1,
			WalletSignStr: "",
		}
	} else {
		return AuthInfo{
			SignType:      AuthV2,
			WalletSignStr: signStr,
		}
	}
}
