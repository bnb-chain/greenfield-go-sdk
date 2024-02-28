package client

import (
	"context"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

// IVirtualGroupClient interface defines basic functions related to Virtual Group.
type IVirtualGroupClient interface {
	QueryVirtualGroupFamily(ctx context.Context, globalVirtualGroupFamilyID uint32) (*types.GlobalVirtualGroupFamily, error)
	QuerySpAvailableGlobalVirtualGroupFamilies(ctx context.Context, spID uint32) ([]uint32, error)
}

// QueryVirtualGroupFamily - Query the virtual group family by ID.
//
// Virtual group family(VGF) serve as a means of grouping global virtual groups. Each bucket must be associated with a unique global virtual group family and cannot cross families.
//
// - ctx: Context variables for the current API call.
//
// - globalVirtualGroupFamilyID: Identify the virtual group family.
//
// - ret1: The virtual group family detail.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) QueryVirtualGroupFamily(ctx context.Context, globalVirtualGroupFamilyID uint32) (*types.GlobalVirtualGroupFamily, error) {
	queryResponse, err := c.chainClient.GlobalVirtualGroupFamily(ctx, &types.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: globalVirtualGroupFamilyID,
	})
	if err != nil {
		return nil, err
	}
	return queryResponse.GlobalVirtualGroupFamily, nil
}

// QuerySpAvailableGlobalVirtualGroupFamilies - Query the virtual group family IDs by SP ID.
//
// Virtual group family(VGF) serve as a means of grouping global virtual groups. Each bucket must be associated with a unique global virtual group family and cannot cross families.
//
// - ctx: Context variables for the current API call.
//
// - spID: Identify the storage provider.
//
// - ret1: The virtual group family detail.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) QuerySpAvailableGlobalVirtualGroupFamilies(ctx context.Context, spID uint32) ([]uint32, error) {
	queryResponse, err := c.chainClient.QuerySpAvailableGlobalVirtualGroupFamilies(ctx, &types.QuerySPAvailableGlobalVirtualGroupFamiliesRequest{
		SpId: spID,
	})
	if err != nil {
		return nil, err
	}
	return queryResponse.GlobalVirtualGroupFamilyIds, nil
}
