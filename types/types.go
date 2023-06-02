package types

import (
	"io"
)

// Principal indicates the marshaled Principal content of greenfield permission types,
// user can generate it by NewPrincipalWithAccount or NewPrincipalWithGroupId method in utils
type Principal string

// ObjectStat contains the metadata of downloaded objects
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
