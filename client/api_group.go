package client

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	gnfdTypes "github.com/bnb-chain/greenfield/types"
	permTypes "github.com/bnb-chain/greenfield/x/permission/types"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog/log"

	"github.com/bnb-chain/greenfield-go-sdk/pkg/utils"
)

// IGroupClient interface defines functions related to Group.
type IGroupClient interface {
	CreateGroup(ctx context.Context, groupName string, opt types.CreateGroupOptions) (string, error)
	DeleteGroup(ctx context.Context, groupName string, opt types.DeleteGroupOption) (string, error)
	UpdateGroupMember(ctx context.Context, groupName string, groupOwnerAddr string,
		addAddresses, removeAddresses []string, opts types.UpdateGroupMemberOption) (string, error)
	LeaveGroup(ctx context.Context, groupName string, groupOwnerAddr string, opt types.LeaveGroupOption) (string, error)
	HeadGroup(ctx context.Context, groupName string, groupOwnerAddr string) (*storageTypes.GroupInfo, error)
	HeadGroupMember(ctx context.Context, groupName string, groupOwner, headMember string) bool
	PutGroupPolicy(ctx context.Context, groupName string, principalAddr string, statements []*permTypes.Statement, opt types.PutPolicyOption) (string, error)
	DeleteGroupPolicy(ctx context.Context, groupName string, principalAddr string, opt types.DeletePolicyOption) (string, error)
	GetBucketPolicyOfGroup(ctx context.Context, bucketName string, groupId uint64) (*permTypes.Policy, error)
	GetObjectPolicyOfGroup(ctx context.Context, bucketName, objectName string, groupId uint64) (*permTypes.Policy, error)
	GetGroupPolicy(ctx context.Context, groupName string, principalAddr string) (*permTypes.Policy, error)
	ListGroup(ctx context.Context, name, prefix string, opts types.ListGroupsOptions) (types.ListGroupsResult, error)
	RenewGroupMember(ctx context.Context, groupOwnerAddr, groupName string, memberAddresses []string, opts types.RenewGroupMemberOption) (string, error)
	ListGroupMembers(ctx context.Context, groupID int64, opts types.GroupMembersPaginationOptions) (*types.GroupMembersResult, error)
	ListGroupsByAccount(ctx context.Context, opts types.GroupsPaginationOptions) (*types.GroupsResult, error)
	ListGroupsByOwner(ctx context.Context, opts types.GroupsOwnerPaginationOptions) (*types.GroupsResult, error)
	// ListGroupsByGroupID list groups by group ids
	ListGroupsByGroupID(ctx context.Context, groupIDs []uint64, opts types.EndPointOptions) (types.ListGroupsByGroupIDResponse, error)
}

// CreateGroup - Create a new group without group members on Greenfield blockchain, and group members can be added by UpdateGroupMember transaction.
//
// A `Group` is a collection of accounts that share the same permissions, allowing them to be handled as a single entity.
//
// Examples of permissions include:
//
// Put, List, Get, Delete, Copy, and Execute data objects;
// Create, Delete, and List buckets
// Create, Delete, ListMembers, Leave groups
// Create, Associate payment accounts
// Grant, Revoke the above permissions
//
// # For more details regarding `Group`, please refer to https://docs.bnbchain.org/greenfield-docs/docs/guide/greenfield-blockchain/modules/permission
//
// - ctx: Context variables for the current API call.
//
// - groupName: The group name identifies the group.
//
// - opt: The options for customizing a group and transaction.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret3: Return error when the request failed, otherwise return nil.
func (c *Client) CreateGroup(ctx context.Context, groupName string, opt types.CreateGroupOptions) (string, error) {
	createGroupMsg := storageTypes.NewMsgCreateGroup(c.MustGetDefaultAccount().GetAddress(), groupName, opt.Extra)
	return c.sendTxn(ctx, createGroupMsg, opt.TxOpts)
}

// DeleteGroup - Delete a group on Greenfield blockchain. The sender MUST only be the group owner, group members or others would fail to send this transaction.
//
// Note: Deleting a group will result in granted permission revoked. Members within the group will no longer have access to resources (bucket, object) which granted permission on.
//
// - ctx: Context variables for the current API call.
//
// - groupName: The group name identifies the group.
//
// - opt: The options for customizing the transaction
//
// - ret1: Transaction hash return from blockchain.
//
// - ret3: Return error when the request failed, otherwise return nil.
func (c *Client) DeleteGroup(ctx context.Context, groupName string, opt types.DeleteGroupOption) (string, error) {
	deleteGroupMsg := storageTypes.NewMsgDeleteGroup(c.MustGetDefaultAccount().GetAddress(), groupName)
	return c.sendTxn(ctx, deleteGroupMsg, opt.TxOpts)
}

// UpdateGroupMember - Update a group by adding or removing members. The sender can be the group owner or any individual account(Principle) that
// has been granted permission by the group owner.
//
// Note: The group owner can only grant ACTION_UPDATE_GROUP_MEMBER permission to an individual account, there is no way to grant permission to a group to allow members
// within such group to manipulate another group or the group itself.
//
// - ctx: Context variables for the current API call.
//
// - groupName: The group name identifies the group.
//
// - groupOwnerAddr: The HEX-encoded string of the group owner address.
//
// - addAddresses: The HEX-encoded string list of the member addresses to be added.
//
// - removeAddresses: The HEX-encoded string list of the member addresses to be removed.
//
// - opt: The options for customizing the group members expiration time and transaction.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) UpdateGroupMember(ctx context.Context, groupName string, groupOwnerAddr string,
	addAddresses, removeAddresses []string, opts types.UpdateGroupMemberOption,
) (string, error) {
	groupOwner, err := sdk.AccAddressFromHexUnsafe(groupOwnerAddr)
	if err != nil {
		return "", err
	}
	if groupName == "" {
		return "", errors.New("group name is empty")
	}
	if len(addAddresses) == 0 && len(removeAddresses) == 0 {
		return "", errors.New("no update member")
	}
	if opts.ExpirationTime != nil && len(addAddresses) != len(opts.ExpirationTime) {
		return "", errors.New("please provide expirationTime for every new add member")
	}
	addMembers := make([]*storageTypes.MsgGroupMember, 0)
	removeMembers := make([]sdk.AccAddress, 0)
	expirationTime := make([]*time.Time, len(addAddresses))
	for idx, addr := range addAddresses {
		_, err := sdk.AccAddressFromHexUnsafe(addr)
		if err != nil {
			return "", err
		}
		if opts.ExpirationTime != nil && opts.ExpirationTime[idx] != nil {
			expirationTime[idx] = opts.ExpirationTime[idx]
		} else {
			expirationTime[idx] = &storageTypes.MaxTimeStamp
		}
		m := &storageTypes.MsgGroupMember{
			Member:         addr,
			ExpirationTime: expirationTime[idx],
		}
		addMembers = append(addMembers, m)
	}

	for _, addr := range removeAddresses {
		member, err := sdk.AccAddressFromHexUnsafe(addr)
		if err != nil {
			return "", err
		}
		removeMembers = append(removeMembers, member)
	}

	updateGroupMsg := storageTypes.NewMsgUpdateGroupMember(c.MustGetDefaultAccount().GetAddress(), groupOwner, groupName, addMembers, removeMembers)

	return c.sendTxn(ctx, updateGroupMsg, opts.TxOpts)
}

// LeaveGroup - Leave a group. A group member initially leaves a group.
//
// - ctx: Context variables for the current API call.
//
// - groupName: The group name identifies the group.
//
// - groupOwnerAddr: The HEX-encoded string of the group owner address.
//
// - opt: The options for customizing the transaction.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) LeaveGroup(ctx context.Context, groupName string, groupOwnerAddr string, opt types.LeaveGroupOption) (string, error) {
	groupOwner, err := sdk.AccAddressFromHexUnsafe(groupOwnerAddr)
	if err != nil {
		return "", err
	}
	leaveGroupMsg := storageTypes.NewMsgLeaveGroup(c.MustGetDefaultAccount().GetAddress(), groupOwner, groupName)
	return c.sendTxn(ctx, leaveGroupMsg, opt.TxOpts)
}

// HeadGroup - Query the groupInfo on chain, return the group info if exists otherwise error.
//
// - ctx: Context variables for the current API call.
//
// - groupName: The group name identifies the group.
//
// - groupOwnerAddr: The HEX-encoded string of the group owner address.
//
// - ret1: The group info details
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) HeadGroup(ctx context.Context, groupName string, groupOwnerAddr string) (*storageTypes.GroupInfo, error) {
	headGroupRequest := storageTypes.QueryHeadGroupRequest{
		GroupOwner: groupOwnerAddr,
		GroupName:  groupName,
	}

	headGroupResponse, err := c.chainClient.HeadGroup(ctx, &headGroupRequest)
	if err != nil {
		return nil, err
	}

	return headGroupResponse.GroupInfo, nil
}

// HeadGroupMember - Query the group member info on chain.
//
// - ctx: Context variables for the current API call.
//
// - groupName: The group name identifies the group.
//
// - groupOwnerAddr: The HEX-encoded string of the group owner address.
//
// - eadMemberAddr: The HEX-encoded string of the group member address
//
// - ret: The boolean value indicates whether the group member exists
func (c *Client) HeadGroupMember(ctx context.Context, groupName string, groupOwnerAddr, headMemberAddr string) bool {
	headGroupRequest := storageTypes.QueryHeadGroupMemberRequest{
		GroupName:  groupName,
		GroupOwner: groupOwnerAddr,
		Member:     headMemberAddr,
	}

	_, err := c.chainClient.HeadGroupMember(ctx, &headGroupRequest)
	return err == nil
}

// PutGroupPolicy - Apply group policy to user specified by principalAddr, the sender needs to be the owner of the group.
//
// - ctx: Context variables for the current API call.
//
// - groupName: The group name identifies the group.
//
// - principalAddr: The HEX-encoded string of the principal address.
//
// - statements: Policies outline the specific details of permissions, including the Effect, ActionList, and Resources.
//
// - opt: The options for customizing the policy expiration time and transaction.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) PutGroupPolicy(ctx context.Context, groupName string, principalAddr string,
	statements []*permTypes.Statement, opt types.PutPolicyOption,
) (string, error) {
	sender := c.MustGetDefaultAccount().GetAddress()

	resource := gnfdTypes.NewGroupGRN(sender, groupName)

	principal, err := sdk.AccAddressFromHexUnsafe(principalAddr)
	if err != nil {
		return "", err
	}

	putPolicyMsg := storageTypes.NewMsgPutPolicy(sender, resource.String(),
		permTypes.NewPrincipalWithAccount(principal), statements, opt.PolicyExpireTime)

	return c.sendPutPolicyTxn(ctx, putPolicyMsg, opt.TxOpts)
}

// GetBucketPolicyOfGroup - Get the bucket policy info of the group.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The bucket name identifies the bucket.
//
// - groupId: The group id identity the group.
//
// - ret1: The bucket policy.
//
// - ret2: Return error when the request failed, otherwise return nil.
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

// GetObjectPolicyOfGroup - Get the object policy info of the group.
//
// - ctx: Context variables for the current API call.
//
// - bucketName: The bucket name identifies the bucket.
//
// - objectName: The object name identifies the object.
//
// - groupId: The group id identity the group.
//
// - ret1: The object policy
//
// - ret2: Return error when the request failed, otherwise return nil.
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

// DeleteGroupPolicy - Delete the group policy of the principal, the sender needs to be the owner of the group
//
// - ctx: Context variables for the current API call.
//
// - groupName: The group name identifies the group.
//
// - principalAddr: The HEX-encoded string of the principal address
//
// - opt: The options for customizing the transaction.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) DeleteGroupPolicy(ctx context.Context, groupName string, principalAddr string, opt types.DeletePolicyOption) (string, error) {
	sender := c.MustGetDefaultAccount().GetAddress()
	resource := gnfdTypes.NewGroupGRN(sender, groupName).String()

	addr, err := sdk.AccAddressFromHexUnsafe(principalAddr)
	if err != nil {
		return "", err
	}

	principal := permTypes.NewPrincipalWithAccount(addr)

	return c.sendDelPolicyTxn(ctx, sender, resource, principal, opt.TxOpts)
}

// GetGroupPolicy - Get the group policy info of the user.
//
// - ctx: Context variables for the current API call.
//
// - groupName: The group name identifies the group.
//
// - principalAddr: The HEX-encoded string of the principal address.
//
// - ret1: The group policy.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) GetGroupPolicy(ctx context.Context, groupName string, principalAddr string) (*permTypes.Policy, error) {
	_, err := sdk.AccAddressFromHexUnsafe(principalAddr)
	if err != nil {
		return nil, err
	}
	sender := c.MustGetDefaultAccount().GetAddress()
	resource := gnfdTypes.NewGroupGRN(sender, groupName).String()

	queryPolicy := storageTypes.QueryPolicyForAccountRequest{
		Resource:         resource,
		PrincipalAddress: principalAddr,
	}

	queryPolicyResp, err := c.chainClient.QueryPolicyForAccount(ctx, &queryPolicy)
	if err != nil {
		return nil, err
	}

	return queryPolicyResp.Policy, nil
}

// ListGroup - Get the group list by name and prefix. It provides fuzzy searches by inputting a specific name and prefix.
//
// - prefix: The start of the search pattern. The system will only return groups that start with this prefix.
//
// - name: The ending of the search pattern.
//
// - ret1: The groups response.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) ListGroup(ctx context.Context, name, prefix string, opts types.ListGroupsOptions) (types.ListGroupsResult, error) {
	const (
		MaximumGetGroupListLimit  = 1000
		MaximumGetGroupListOffset = 100000
		DefaultGetGroupListLimit  = 50
	)

	if name == "" {
		return types.ListGroupsResult{}, nil
	}

	if prefix == "" {
		return types.ListGroupsResult{}, nil
	}

	if opts.Limit < 0 {
		return types.ListGroupsResult{}, nil
	} else if opts.Limit > 1000 {
		opts.Limit = MaximumGetGroupListLimit
	} else if opts.Limit == 0 {
		opts.Limit = DefaultGetGroupListLimit
	}

	if opts.Offset < 0 || opts.Offset > MaximumGetGroupListOffset {
		return types.ListGroupsResult{}, nil
	}

	if opts.SourceType != "" {
		if _, ok := storageTypes.SourceType_value[opts.SourceType]; !ok {
			return types.ListGroupsResult{}, nil
		}
	}

	params := url.Values{}
	params.Set("group-query", "")
	params.Set("name", name)
	params.Set("prefix", prefix)
	params.Set("source-type", opts.SourceType)
	params.Set("limit", strconv.FormatInt(opts.Limit, 10))
	params.Set("offset", strconv.FormatInt(opts.Offset, 10))
	reqMeta := requestMeta{
		urlValues:     params,
		contentSHA256: types.EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getEndpointByOpt(&types.EndPointOptions{
		Endpoint:  opts.Endpoint,
		SPAddress: opts.SPAddress,
	})
	if err != nil {
		log.Error().Msg(fmt.Sprintf("get endpoint by option failed %s", err.Error()))
		return types.ListGroupsResult{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		log.Error().Msg("the list of groups failed: " + err.Error())
		return types.ListGroupsResult{}, err
	}
	defer utils.CloseResponse(resp)

	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msg("the list of groups failed: " + err.Error())
		return types.ListGroupsResult{}, err
	}

	listGroupsResult := types.ListGroupsResult{}
	bufStr := buf.String()
	err = xml.Unmarshal([]byte(bufStr), &listGroupsResult)
	if err != nil {
		log.Error().Msg("the list of groups failed: " + err.Error())
		return types.ListGroupsResult{}, err
	}

	return listGroupsResult, nil
}

// RenewGroupMember - Renew a list group members and their expiration time.
//
// - ctx: Context variables for the current API call.
//
// - groupName: The group name identifies the group.
//
// - groupOwnerAddr: The HEX-encoded string of the group owner address.
//
// - memberAddresses: The HEX-encoded string list of the member addresses to be renewed.
//
// - opts: The options for customizing the transaction and each group member's expiration time.
//
// - ret1: Transaction hash return from blockchain.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) RenewGroupMember(ctx context.Context, groupOwnerAddr, groupName string,
	memberAddresses []string, opts types.RenewGroupMemberOption,
) (string, error) {
	groupOwner, err := sdk.AccAddressFromHexUnsafe(groupOwnerAddr)
	if err != nil {
		return "", err
	}
	if groupName == "" {
		return "", errors.New("group name is empty")
	}
	renewMembers := make([]*storageTypes.MsgGroupMember, 0)
	if opts.ExpirationTime != nil && len(memberAddresses) != len(opts.ExpirationTime) {
		return "", errors.New("please provide expirationTime for every new add member")
	}
	expirationTime := make([]*time.Time, len(memberAddresses))
	for idx, addr := range memberAddresses {
		_, err := sdk.AccAddressFromHexUnsafe(addr)
		if err != nil {
			return "", err
		}
		if opts.ExpirationTime != nil && opts.ExpirationTime[idx] != nil {
			expirationTime[idx] = opts.ExpirationTime[idx]
		} else {
			expirationTime[idx] = &storageTypes.MaxTimeStamp
		}
		m := &storageTypes.MsgGroupMember{
			Member:         addr,
			ExpirationTime: expirationTime[idx],
		}
		renewMembers = append(renewMembers, m)
	}
	msg := storageTypes.NewMsgRenewGroupMember(c.MustGetDefaultAccount().GetAddress(), groupOwner, groupName, renewMembers)
	return c.sendTxn(ctx, msg, opts.TxOpts)
}

// ListGroupMembers - List members within a group, including those for which the user's expiration time has already elapsed.
//
// - groupId: The group id identifies a group.
//
// - ret1: Group members detail.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) ListGroupMembers(ctx context.Context, groupID int64, opts types.GroupMembersPaginationOptions) (*types.GroupMembersResult, error) {
	params := url.Values{}
	params.Set("group-members", "")
	params.Set("group-id", strconv.FormatInt(groupID, 10))
	params.Set("start-after", opts.StartAfter)
	params.Set("limit", strconv.FormatInt(opts.Limit, 10))

	reqMeta := requestMeta{
		urlValues:     params,
		contentSHA256: types.EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getEndpointByOpt(&types.EndPointOptions{
		Endpoint:  opts.Endpoint,
		SPAddress: opts.SPAddress,
	})
	if err != nil {
		log.Error().Msg(fmt.Sprintf("get endpoint by option failed %s", err.Error()))
		return &types.GroupMembersResult{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return &types.GroupMembersResult{}, err
	}
	defer utils.CloseResponse(resp)

	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msgf("get groups info by a user address in group id:%v failed: %s", groupID, err.Error())
		return &types.GroupMembersResult{}, err
	}

	var groups *types.GroupMembersResult
	bufStr := buf.String()
	// TODO change the format to XML later
	err = xml.Unmarshal([]byte(bufStr), &groups)
	if err != nil {
		log.Error().Msgf("get groups info by a user address in group id:%v failed: %s", groupID, err.Error())
		return &types.GroupMembersResult{}, err
	}

	return groups, nil
}

// ListGroupsByAccount - List groups that a user has joined, including those which the user's expiration time has already elapsed
//
// - ctx: Context variables for the current API call.
//
// - opts: The query option, By default, the user is the sender. Other users can be set using the option
//
// - ret1: Groups details.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) ListGroupsByAccount(ctx context.Context, opts types.GroupsPaginationOptions) (*types.GroupsResult, error) {
	params := url.Values{}
	params.Set("user-groups", "")
	params.Set("start-after", opts.StartAfter)
	params.Set("limit", strconv.FormatInt(opts.Limit, 10))

	account := opts.Account
	if account == "" {
		acc, err := c.GetDefaultAccount()
		if err != nil {
			log.Error().Msg(fmt.Sprintf("failed to get default account:  %s", err.Error()))
			return &types.GroupsResult{}, err
		}
		account = acc.GetAddress().String()
	}

	reqMeta := requestMeta{
		urlValues:     params,
		contentSHA256: types.EmptyStringSHA256,
		userAddress:   account,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getEndpointByOpt(&types.EndPointOptions{
		Endpoint:  opts.Endpoint,
		SPAddress: opts.SPAddress,
	})
	if err != nil {
		log.Error().Msg(fmt.Sprintf("get endpoint by option failed %s", err.Error()))
		return &types.GroupsResult{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return &types.GroupsResult{}, err
	}
	defer utils.CloseResponse(resp)

	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msgf("get group members by group id in account id:%v failed: %s", account, err.Error())
		return &types.GroupsResult{}, err
	}

	var groups *types.GroupsResult
	bufStr := buf.String()
	// TODO change the format to XML later
	err = xml.Unmarshal([]byte(bufStr), &groups)
	if err != nil {
		log.Error().Msgf("get group members by group id in account id:%v failed: %s", account, err.Error())
		return &types.GroupsResult{}, err
	}

	return groups, nil
}

// ListGroupsByOwner - List groups owned by the specified user, including those for which the user's expiration time has already elapsed
//
// - ctx: Context variables for the current API call.
//
// - opts: The query option, By default, the user is the sender. Other users can be set using the option
//
// - ret1: Groups details.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *Client) ListGroupsByOwner(ctx context.Context, opts types.GroupsOwnerPaginationOptions) (*types.GroupsResult, error) {
	params := url.Values{}
	params.Set("owned-groups", "")
	params.Set("start-after", opts.StartAfter)
	params.Set("limit", strconv.FormatInt(opts.Limit, 10))

	owner := opts.Owner
	if owner == "" {
		acc, err := c.GetDefaultAccount()
		if err != nil {
			log.Error().Msg(fmt.Sprintf("get default owner failed %s", err.Error()))
			return &types.GroupsResult{}, err
		}
		owner = acc.GetAddress().String()
	}

	reqMeta := requestMeta{
		urlValues:     params,
		contentSHA256: types.EmptyStringSHA256,
		userAddress:   owner,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getEndpointByOpt(&types.EndPointOptions{
		Endpoint:  opts.Endpoint,
		SPAddress: opts.SPAddress,
	})
	if err != nil {
		log.Error().Msg(fmt.Sprintf("get endpoint by option failed %s", err.Error()))
		return &types.GroupsResult{}, err
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return &types.GroupsResult{}, err
	}
	defer utils.CloseResponse(resp)

	// unmarshal the json content from response body
	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msgf("retrieve groups where the user is the owner in account id:%v failed: %s", owner, err.Error())
		return &types.GroupsResult{}, err
	}

	var groups *types.GroupsResult
	// TODO change the format to XML later
	bufStr := buf.String()
	err = xml.Unmarshal([]byte(bufStr), &groups)
	if err != nil {
		log.Error().Msgf("retrieve groups where the user is the owner in account id:%v failed: %s", owner, err.Error())
		return &types.GroupsResult{}, err
	}

	return groups, nil
}

type GfSpListGroupsByGroupIDsResponse map[uint64]*types.GroupMeta

type GroupEntry struct {
	Id    uint64
	Value *types.GroupMeta
}

func (m *GfSpListGroupsByGroupIDsResponse) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*m = GfSpListGroupsByGroupIDsResponse{}
	for {
		var e GroupEntry

		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		} else {
			(*m)[e.Id] = e.Value
		}
	}
	return nil
}

// ListGroupsByGroupID - List groups by group ids.
//
// By inputting a collection of group IDs, we can retrieve the corresponding object data. If the group is nonexistent or has been deleted, a null value will be returned
//
// - ctx: Context variables for the current API call.
//
// - groupIDs: The list of group ids.
//
// - opts: The options to set the meta to list groups by group id.
//
// - ret1: The result of group info map by given group ids.
//
// - ret2: Return error when the request failed, otherwise return nil.
func (c *client) ListGroupsByGroupID(ctx context.Context, groupIDs []uint64, opts types.EndPointOptions) (types.ListGroupsByGroupIDResponse, error) {
	const MaximumListBucketsSize = 1000
	if len(groupIDs) == 0 || len(groupIDs) > MaximumListBucketsSize {
		return types.ListGroupsByGroupIDResponse{}, nil
	}

	groupIDMap := make(map[uint64]bool)
	for _, id := range groupIDs {
		if _, ok := groupIDMap[id]; ok {
			// repeat id keys in request
			return types.ListGroupsByGroupIDResponse{}, nil
		}
		groupIDMap[id] = true
	}

	idStr := make([]string, len(groupIDs))
	for i, id := range groupIDs {
		idStr[i] = strconv.FormatUint(id, 10)
	}
	IDs := strings.Join(idStr, ",")

	params := url.Values{}
	params.Set("groups-query", "")
	params.Set("ids", IDs)

	reqMeta := requestMeta{
		urlValues:     params,
		contentSHA256: types.EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	endpoint, err := c.getEndpointByOpt(&opts)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("get endpoint by option failed %s", err.Error()))
		return types.ListGroupsByGroupIDResponse{}, err

	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, endpoint)
	if err != nil {
		return types.ListGroupsByGroupIDResponse{}, err
	}
	defer utils.CloseResponse(resp)

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Error().Msgf("the list of groups in group ids:%v failed: %s", groupIDs, err.Error())
		return types.ListGroupsByGroupIDResponse{}, err
	}

	groups := types.ListGroupsByGroupIDResponse{}
	bufStr := buf.String()
	err = xml.Unmarshal([]byte(bufStr), (*GfSpListGroupsByGroupIDsResponse)(&groups.Groups))
	if err != nil && groups.Groups == nil {
		log.Error().Msgf("the list of groups in group ids:%v failed: %s", groups, err.Error())
		return types.ListGroupsByGroupIDResponse{}, err
	}

	return groups, nil
}
