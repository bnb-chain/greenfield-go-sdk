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

type Group interface {
	// CreateGroup create a new group on greenfield chain the group members can be initialized  or not
	CreateGroup(ctx context.Context, groupName string, opt types.CreateGroupOptions) (string, error)
	// DeleteGroup send DeleteGroup txn to greenfield chain and return txn hash
	DeleteGroup(ctx context.Context, groupName string, opt types.DeleteGroupOption) (string, error)
	// UpdateGroupMember support adding or removing members from the group and return the txn hash
	UpdateGroupMember(ctx context.Context, groupName string, groupOwnerAddr string, addAddresses, removeAddresses []string, opts types.UpdateGroupMemberOption) (string, error)
	// LeaveGroup make the member leave the specific group
	LeaveGroup(ctx context.Context, groupName string, groupOwnerAddr string, opt types.LeaveGroupOption) (string, error)
	// HeadGroup query the groupInfo on chain, return the group info if exists return err info if group not exist
	HeadGroup(ctx context.Context, groupName string, groupOwnerAddr string) (*storageTypes.GroupInfo, error)
	// HeadGroupMember query the group member info on chain, return true if the member exists in group
	HeadGroupMember(ctx context.Context, groupName string, groupOwner, headMember string) bool
	// PutGroupPolicy apply group policy to user specified by principalAddr, the sender need to be the owner of the group
	PutGroupPolicy(ctx context.Context, groupName string, principalAddr string, statements []*permTypes.Statement, opt types.PutPolicyOption) (string, error)
	// DeleteGroupPolicy  delete group policy of the principal, the sender need to be the owner of the group
	DeleteGroupPolicy(ctx context.Context, groupName string, principalAddr string, opt types.DeletePolicyOption) (string, error)
	// GetBucketPolicyOfGroup get the bucket policy info of the group specified by group id
	// it queries a bucket policy that grants permission to a group
	GetBucketPolicyOfGroup(ctx context.Context, bucketName string, groupId uint64) (*permTypes.Policy, error)
	// GetObjectPolicyOfGroup get the object policy info of the group specified by group id
	// it queries an object policy that grants permission to a group
	GetObjectPolicyOfGroup(ctx context.Context, bucketName, objectName string, groupId uint64) (*permTypes.Policy, error)
	// GetGroupPolicy get the group policy info of the user specified by principalAddr
	GetGroupPolicy(ctx context.Context, groupName string, principalAddr string) (*permTypes.Policy, error)
	// ListGroup get the group list by name and prefix.
	ListGroup(ctx context.Context, name, prefix string, opts types.ListGroupsOptions) (types.ListGroupsResult, error)
	// RenewGroupMember renew a list of group members and their expiration time
	RenewGroupMember(ctx context.Context, groupOwnerAddr, groupName string, memberAddresses []string, opts types.RenewGroupMemberOption) (string, error)
	// ListGroupMembers returns a list of members contained within the group specified by the group id, including those for which the user's expiration time has already elapsed
	ListGroupMembers(ctx context.Context, groupID int64, opts types.GroupMembersPaginationOptions) (*types.GroupMembersResult, error)
	// ListGroupsByAccount returns a list of all groups that the user has joined, including those for which the user's expiration time has already elapsed
	// By default, the user is the sender. Other users can be set using the option
	ListGroupsByAccount(ctx context.Context, opts types.GroupsPaginationOptions) (*types.GroupsResult, error)
	// ListGroupsByOwner returns a list of groups owned by the specified user, including those for which the user's expiration time has already elapsed
	// By default, the user is the sender. Other users can be set using the option
	ListGroupsByOwner(ctx context.Context, opts types.GroupsOwnerPaginationOptions) (*types.GroupsResult, error)
}

// CreateGroup create a new group on greenfield chain, the group members can be initialized or not
func (c *client) CreateGroup(ctx context.Context, groupName string, opt types.CreateGroupOptions) (string, error) {
	createGroupMsg := storageTypes.NewMsgCreateGroup(c.MustGetDefaultAccount().GetAddress(), groupName, opt.Extra)
	return c.sendTxn(ctx, createGroupMsg, opt.TxOpts)
}

// DeleteGroup send DeleteGroup txn to greenfield chain and return txn hash
func (c *client) DeleteGroup(ctx context.Context, groupName string, opt types.DeleteGroupOption) (string, error) {
	deleteGroupMsg := storageTypes.NewMsgDeleteGroup(c.MustGetDefaultAccount().GetAddress(), groupName)
	return c.sendTxn(ctx, deleteGroupMsg, opt.TxOpts)
}

// UpdateGroupMember support adding or removing members from the group and return the txn hash
// groupOwnerAddr is the HEX-encoded string of the group owner address
// addAddresses indicates the HEX-encoded string list of the member addresses to be added
// removeAddresses indicates the HEX-encoded string list of the member addresses to be removed
// At least one of add Addresses or remove Addresses must be set, or both can be set.
func (c *client) UpdateGroupMember(ctx context.Context, groupName string, groupOwnerAddr string,
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

// LeaveGroup make the member leave the specific group
// groupOwnerAddr indicates the HEX-encoded string of the group owner address
func (c *client) LeaveGroup(ctx context.Context, groupName string, groupOwnerAddr string, opt types.LeaveGroupOption) (string, error) {
	groupOwner, err := sdk.AccAddressFromHexUnsafe(groupOwnerAddr)
	if err != nil {
		return "", err
	}
	leaveGroupMsg := storageTypes.NewMsgLeaveGroup(c.MustGetDefaultAccount().GetAddress(), groupOwner, groupName)
	return c.sendTxn(ctx, leaveGroupMsg, opt.TxOpts)
}

// HeadGroup query the groupInfo on chain, return the group info if exists
// return err info if group not exist
func (c *client) HeadGroup(ctx context.Context, groupName string, groupOwnerAddr string) (*storageTypes.GroupInfo, error) {
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

// HeadGroupMember query the group member info on chain, return true if the member exists in group
// groupOwnerAddr indicates the HEX-encoded string of the group owner address
// headMemberAddr indicates the HEX-encoded string of the member address to query for
func (c *client) HeadGroupMember(ctx context.Context, groupName string, groupOwnerAddr, headMemberAddr string) bool {
	headGroupRequest := storageTypes.QueryHeadGroupMemberRequest{
		GroupName:  groupName,
		GroupOwner: groupOwnerAddr,
		Member:     headMemberAddr,
	}

	_, err := c.chainClient.HeadGroupMember(ctx, &headGroupRequest)
	return err == nil
}

// PutGroupPolicy apply group policy to user specified by principalAddr, the sender need to be the owner of the group
// principalAddr indicates the HEX-encoded string of the principal address
// Statement defines the permission info of a resource, users can use NewStatement function to init it
func (c *client) PutGroupPolicy(ctx context.Context, groupName string, principalAddr string,
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

// GetBucketPolicyOfGroup get the bucket policy info of the group specified by group id
// it queries a bucket policy that grants permission to a group
func (c *client) GetBucketPolicyOfGroup(ctx context.Context, bucketName string, groupId uint64) (*permTypes.Policy, error) {
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
func (c *client) GetObjectPolicyOfGroup(ctx context.Context, bucketName, objectName string, groupId uint64) (*permTypes.Policy, error) {
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

// DeleteGroupPolicy delete group policy of the principal, the sender need to be the owner of the group
// principalAddr indicates the HEX-encoded string of the principal address
func (c *client) DeleteGroupPolicy(ctx context.Context, groupName string, principalAddr string, opt types.DeletePolicyOption) (string, error) {
	sender := c.MustGetDefaultAccount().GetAddress()
	resource := gnfdTypes.NewGroupGRN(sender, groupName).String()

	addr, err := sdk.AccAddressFromHexUnsafe(principalAddr)
	if err != nil {
		return "", err
	}

	principal := permTypes.NewPrincipalWithAccount(addr)

	return c.sendDelPolicyTxn(ctx, sender, resource, principal, opt.TxOpts)
}

// GetGroupPolicy get the group policy info of the user specified by principalAddr
// principalAddr indicates the HEX-encoded string of the principal address
func (c *client) GetGroupPolicy(ctx context.Context, groupName string, principalAddr string) (*permTypes.Policy, error) {
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

// ListGroup get the group list by name and prefix.
// prefix is the start of the search pattern. The system will only return groups that start with this prefix.
// name is the ending of the search pattern. it providers fuzzy searches by inputting a specific name and prefix
func (c *client) ListGroup(ctx context.Context, name, prefix string, opts types.ListGroupsOptions) (types.ListGroupsResult, error) {
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

// RenewGroupMember renew a list of group members and their expiration time
func (c *client) RenewGroupMember(ctx context.Context, groupOwnerAddr, groupName string,
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

// ListGroupMembers returns a list of members contained within the group specified by the group id, including those for which the user's expiration time has already elapsed
func (c *client) ListGroupMembers(ctx context.Context, groupID int64, opts types.GroupMembersPaginationOptions) (*types.GroupMembersResult, error) {
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

// ListGroupsByAccount returns a list of all groups that the user has joined, including those for which the user's expiration time has already elapsed
// By default, the user is the sender. Other users can be set using the option
func (c *client) ListGroupsByAccount(ctx context.Context, opts types.GroupsPaginationOptions) (*types.GroupsResult, error) {
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

// ListGroupsByOwner returns a list of groups owned by the specified user, including those for which the user's expiration time has already elapsed
// By default, the user is the sender. Other users can be set using the option
func (c *client) ListGroupsByOwner(ctx context.Context, opts types.GroupsOwnerPaginationOptions) (*types.GroupsResult, error) {
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
