package client

import (
	"context"

	spTypes "github.com/bnb-chain/greenfield/x/sp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SP interface {
	// ListSP return the storage provider info on chain
	// isInService indicates if only display the sp with STATUS_IN_SERVICE status
	ListSP(ctx context.Context, isInService bool) ([]spTypes.StorageProvider, error)
	// GetSPInfo return the sp info  the sp chain address
	GetSPInfo(ctx context.Context, SPAddr sdk.AccAddress) (*spTypes.StorageProvider, error)
	// GetSpAddrFromEndpoint return the chain addr according to the SP endpoint
	GetSpAddrFromEndpoint(ctx context.Context) (sdk.AccAddress, error)
	CreateStorageProvider()
	EditStorageProvider()
	VoteCreateStorageProvider()
}
