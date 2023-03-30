package api

import (
	"context"
	"errors"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield/sdk/types"
	gnfdTypes "github.com/bnb-chain/greenfield/types"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CreateGroup create a new group on greenfield chain
// the group members can be initialized  or not
func (c *Client) CreateGroup(groupName string, opt client.CreateGroupOptions) (string, error) {
	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	createGroupMsg := storageTypes.NewMsgCreateGroup(km.GetAddr(), groupName, opt.InitGroupMember)

	if err = createGroupMsg.ValidateBasic(); err != nil {
		return "", err
	}

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{createGroupMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// DeleteGroup send DeleteGroup txn to greenfield chain and return txn hash
func (c *Client) DeleteGroup(groupName string, txOpts types.TxOption) (string, error) {
	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	deleteGroupMsg := storageTypes.NewMsgDeleteGroup(km.GetAddr(), groupName)
	if err = deleteGroupMsg.ValidateBasic(); err != nil {
		return "", err
	}

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{deleteGroupMsg}, &txOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, err
}

// UpdateGroupMember support adding or removing members from the group and return the txn hash
func (c *Client) UpdateGroupMember(groupName string, groupOwner sdk.AccAddress,
	addMembers, removeMembers []sdk.AccAddress, opts client.UpdateGroupMemberOption) (string, error) {
	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	if groupName == "" {
		return "", errors.New("group name is empty")
	}

	if len(addMembers) == 0 && len(removeMembers) == 0 {
		return "", errors.New("no update member")
	}
	updateGroupMsg := storageTypes.NewMsgUpdateGroupMember(km.GetAddr(), groupOwner, groupName, addMembers, removeMembers)
	if err = updateGroupMsg.ValidateBasic(); err != nil {
		return "", err
	}

	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{updateGroupMsg}, opts.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, nil
}

func (c *Client) LeaveGroup(groupName string, groupOwner sdk.AccAddress, opt client.LeaveGroupOption) (string, error) {
	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}

	leaveGroupMsg := storageTypes.NewMsgLeaveGroup(km.GetAddr(), groupOwner, groupName)
	resp, err := c.chainClient.BroadcastTx([]sdk.Msg{leaveGroupMsg}, opt.TxOpts)
	if err != nil {
		return "", err
	}

	return resp.TxResponse.TxHash, nil
}

// HeadGroup query the groupInfo on chain, return the group info if exists
// return err info if group not exist
func (c *Client) HeadGroup(ctx context.Context, groupName string, groupOwner sdk.AccAddress) (*storageTypes.GroupInfo, error) {
	headGroupRequest := storageTypes.QueryHeadGroupRequest{
		GroupOwner: groupOwner.String(),
		GroupName:  groupName,
	}

	headGroupResponse, err := c.chainClient.HeadGroup(ctx, &headGroupRequest)
	if err != nil {
		return nil, err
	}

	return headGroupResponse.GroupInfo, nil
}

// HeadGroupMember query the group member info on chain, return true if the member exists in group
func (c *Client) HeadGroupMember(ctx context.Context, groupName string, groupOwner, headMember sdk.AccAddress) bool {
	headGroupRequest := storageTypes.QueryHeadGroupMemberRequest{
		GroupName:  groupName,
		GroupOwner: groupOwner.String(),
		Member:     headMember.String(),
	}

	_, err := c.chainClient.HeadGroupMember(ctx, &headGroupRequest)
	return err == nil
}

// PutGroupPolicy apply group policy to user specified by principalAddr, the sender need to be the owner of the group
func (c *Client) PutGroupPolicy(groupName string, principalAddr sdk.AccAddress,
	statements []*permTypes.Statement, opt client.PutPolicyOption) (string, error) {
	km, err := c.chainClient.GetKeyManager()
	if err != nil {
		return "", errors.New("key manager is nil")
	}
	sender := km.GetAddr()

	resource := gnfdTypes.NewGroupGRN(sender, groupName)
	putPolicyMsg := storageTypes.NewMsgPutPolicy(km.GetAddr(), resource.String(),
		permTypes.NewPrincipalWithAccount(principalAddr), statements, opt.PolicyExpireTime)

	return c.sendPutPolicyTxn(putPolicyMsg, opt.TxOpts)
}

// GetBucketPolicyOfGroup get the bucket policy info of the group specified by group id
// it queries a bucket policy that grants permission to a group
func (c *Client) GetBucketPolicyOfGroup(ctx context.Context, bucketName string, groupId uint64) (*permTypes.Policy, error) {
	resource := gnfdTypes.NewBucketGRN(bucketName).String()

	queryPolicy := storageTypes.QueryPolicyForGroupRequest{
		Resource:         resource,
		PrincipalGroupId: sdkmath.NewUint(groupId).String(),
	}

	queryPolicyResp, err := c.chainClient.QueryPolicyForGroup(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}

// GetObjectPolicyOfGroup get the object policy info of the group specified by group id
// it queries an object policy that grants permission to a group
func (c *Client) GetObjectPolicyOfGroup(ctx context.Context, bucketName, objectName string, groupId uint64) (*permTypes.Policy, error) {
	resource := gnfdTypes.NewObjectGRN(bucketName, objectName)
	queryPolicy := storageTypes.QueryPolicyForGroupRequest{
		Resource:         resource.String(),
		PrincipalGroupId: sdkmath.NewUint(groupId).String(),
	}

	queryPolicyResp, err := c.chainClient.QueryPolicyForGroup(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}
