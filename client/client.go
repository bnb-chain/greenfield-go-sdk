package client

import (
	"context"
	_ "encoding/json"
	bfstypes "github.com/bnb-chain/bfs/x/bfs/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"google.golang.org/grpc"
)

//
//type GreenlandClient struct {
//	upgradetypes.QueryClient
//	distrtypes.MsgClient
//	slashingtypes.QueryClient
//	slashingtypes.MsgClient
//	stakingtypes.QueryClient
//	stakingtypes.MsgClient
//	authtypes.QueryClient
//	banktypes.QueryClient
//	banktypes.MsgClient
//	v1beta1.QueryClient
//	v1beta1.MsgClient
//	authztypes.QueryClient
//	authztypes.MsgClient
//	feegranttypes.QueryClient
//	feegranttypes.MsgClient
//	paramstypes.QueryClient
//	bfstypes.QueryClient
//	bfstypes.MsgClient
//}

type UpgradeQueryClient = upgradetypes.QueryClient
type DistrQueryClient = distrtypes.QueryClient
type DistrMsgClient = distrtypes.MsgClient
type SlashingQueryClient = slashingtypes.QueryClient
type SlashingMsgClient = slashingtypes.MsgClient
type StakingQueryClient = stakingtypes.QueryClient
type StakingMsgClient = stakingtypes.MsgClient
type AuthQueryClient = authtypes.QueryClient
type BankQueryClient = banktypes.QueryClient
type BankMsgClient = banktypes.MsgClient
type GovQueryClient = v1beta1.QueryClient
type GovMsgClient = v1beta1.MsgClient
type AuthzQueryClient = authztypes.QueryClient
type AuthzMsgClient = authztypes.MsgClient
type FeegrantQueryClient = feegranttypes.QueryClient
type FeegrantMsgClient = feegranttypes.MsgClient
type ParamsQueryClient = paramstypes.QueryClient
type BfsQueryClient = bfstypes.QueryClient
type BfsMsgClient = bfstypes.MsgClient

type GreenlandClient struct {
	UpgradeQueryClient
	DistrQueryClient
	DistrMsgClient
	SlashingQueryClient
	SlashingMsgClient
	StakingQueryClient
	StakingMsgClient
	AuthQueryClient
	BankQueryClient
	BankMsgClient
	GovQueryClient
	GovMsgClient
	AuthzQueryClient
	AuthzMsgClient
	FeegrantQueryClient
	FeegrantMsgClient
	ParamsQueryClient
	BfsQueryClient
	BfsMsgClient
}

func grpcConn(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return conn
}

func NewGreenlandClient(grpcAddr string) (GreenlandClient, error) {
	conn := grpcConn(grpcAddr)

	return GreenlandClient{
		upgradetypes.NewQueryClient(conn),
		distrtypes.NewQueryClient(conn),
		distrtypes.NewMsgClient(conn),
		slashingtypes.NewQueryClient(conn),
		slashingtypes.NewMsgClient(conn),
		stakingtypes.NewQueryClient(conn),
		stakingtypes.NewMsgClient(conn),
		authtypes.NewQueryClient(conn),
		banktypes.NewQueryClient(conn),
		banktypes.NewMsgClient(conn),
		v1beta1.NewQueryClient(conn),
		v1beta1.NewMsgClient(conn),
		authztypes.NewQueryClient(conn),
		authztypes.NewMsgClient(conn),
		feegranttypes.NewQueryClient(conn),
		feegranttypes.NewMsgClient(conn),
		paramstypes.NewQueryClient(conn),
		bfstypes.NewQueryClient(conn),
		bfstypes.NewMsgClient(conn),
	}, nil
}

func NewGreenlandClients(rpcAddrs []string, grpcAddrs []string) ([]*GreenlandClient, error) {
	greenlandClients := make([]*GreenlandClient, 0)

	for i := 0; i < len(rpcAddrs); i++ {
		conn := grpcConn(grpcAddrs[i])
		greenlandClients = append(greenlandClients, &GreenlandClient{
			upgradetypes.NewQueryClient(conn),
			distrtypes.NewQueryClient(conn),
			distrtypes.NewMsgClient(conn),
			slashingtypes.NewQueryClient(conn),
			slashingtypes.NewMsgClient(conn),
			stakingtypes.NewQueryClient(conn),
			stakingtypes.NewMsgClient(conn),
			authtypes.NewQueryClient(conn),
			banktypes.NewQueryClient(conn),
			banktypes.NewMsgClient(conn),
			v1beta1.NewQueryClient(conn),
			v1beta1.NewMsgClient(conn),
			authztypes.NewQueryClient(conn),
			authztypes.NewMsgClient(conn),
			feegranttypes.NewQueryClient(conn),
			feegranttypes.NewMsgClient(conn),
			paramstypes.NewQueryClient(conn),
			bfstypes.NewQueryClient(conn),
			bfstypes.NewMsgClient(conn),
		})
	}
	return greenlandClients, nil
}

// TODO: Bank Query
func (c GreenlandClient) Balance(ctx context.Context, in *banktypes.QueryBalanceRequest, opts ...grpc.CallOption) (*banktypes.QueryBalanceResponse, error) {
	res, err := c.BankQueryClient.Balance(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) AllBalances(ctx context.Context, in *banktypes.QueryAllBalancesRequest, opts ...grpc.CallOption) (*banktypes.QueryAllBalancesResponse, error) {
	res, err := c.BankQueryClient.AllBalances(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) SpendableBalances(ctx context.Context, in *banktypes.QuerySpendableBalancesRequest, opts ...grpc.CallOption) (*banktypes.QuerySpendableBalancesResponse, error) {
	res, err := c.BankQueryClient.SpendableBalances(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) TotalSupply(ctx context.Context, in *banktypes.QueryTotalSupplyRequest, opts ...grpc.CallOption) (*banktypes.QueryTotalSupplyResponse, error) {
	res, err := c.BankQueryClient.TotalSupply(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) SupplyOf(ctx context.Context, in *banktypes.QuerySupplyOfRequest, opts ...grpc.CallOption) (*banktypes.QuerySupplyOfResponse, error) {
	res, err := c.BankQueryClient.SupplyOf(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) BankParams(ctx context.Context, in *banktypes.QueryParamsRequest, opts ...grpc.CallOption) (*banktypes.QueryParamsResponse, error) {
	res, err := c.BankQueryClient.Params(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) DenomMetadata(ctx context.Context, in *banktypes.QueryDenomMetadataRequest, opts ...grpc.CallOption) (*banktypes.QueryDenomMetadataResponse, error) {
	res, err := c.BankQueryClient.DenomMetadata(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) DenomsMetadata(ctx context.Context, in *banktypes.QueryDenomsMetadataRequest, opts ...grpc.CallOption) (*banktypes.QueryDenomsMetadataResponse, error) {
	res, err := c.BankQueryClient.DenomsMetadata(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) DenomOwners(ctx context.Context, in *banktypes.QueryDenomOwnersRequest, opts ...grpc.CallOption) (*banktypes.QueryDenomOwnersResponse, error) {
	res, err := c.BankQueryClient.DenomOwners(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO: Upgrade Query
func (c GreenlandClient) CurrentPlan(ctx context.Context, in *upgradetypes.QueryCurrentPlanRequest, opts ...grpc.CallOption) (*upgradetypes.QueryCurrentPlanResponse, error) {
	res, err := c.UpgradeQueryClient.CurrentPlan(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) AppliedPlan(ctx context.Context, in *upgradetypes.QueryAppliedPlanRequest, opts ...grpc.CallOption) (*upgradetypes.QueryAppliedPlanResponse, error) {
	res, err := c.UpgradeQueryClient.AppliedPlan(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) UpgradedConsensusState(ctx context.Context, in *upgradetypes.QueryUpgradedConsensusStateRequest, opts ...grpc.CallOption) (*upgradetypes.QueryUpgradedConsensusStateResponse, error) {
	res, err := c.UpgradeQueryClient.UpgradedConsensusState(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil

}

func (c GreenlandClient) ModuleVersions(ctx context.Context, in *upgradetypes.QueryModuleVersionsRequest, opts ...grpc.CallOption) (*upgradetypes.QueryModuleVersionsResponse, error) {
	res, err := c.UpgradeQueryClient.ModuleVersions(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO: Distr Query
func (c GreenlandClient) DistrParams(ctx context.Context, in *distrtypes.QueryParamsRequest, opts ...grpc.CallOption) (*distrtypes.QueryParamsResponse, error) {
	res, err := c.DistrQueryClient.Params(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) ValidatorOutstandingRewards(ctx context.Context, in *distrtypes.QueryValidatorOutstandingRewardsRequest, opts ...grpc.CallOption) (*distrtypes.QueryValidatorOutstandingRewardsResponse, error) {
	res, err := c.DistrQueryClient.ValidatorOutstandingRewards(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) ValidatorCommission(ctx context.Context, in *distrtypes.QueryValidatorCommissionRequest, opts ...grpc.CallOption) (*distrtypes.QueryValidatorCommissionResponse, error) {
	res, err := c.DistrQueryClient.ValidatorCommission(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) ValidatorSlashes(ctx context.Context, in *distrtypes.QueryValidatorSlashesRequest, opts ...grpc.CallOption) (*distrtypes.QueryValidatorSlashesResponse, error) {
	res, err := c.DistrQueryClient.ValidatorSlashes(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) DelegationRewards(ctx context.Context, in *distrtypes.QueryDelegationRewardsRequest, opts ...grpc.CallOption) (*distrtypes.QueryDelegationRewardsResponse, error) {
	res, err := c.DistrQueryClient.DelegationRewards(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) DelegationTotalRewards(ctx context.Context, in *distrtypes.QueryDelegationTotalRewardsRequest, opts ...grpc.CallOption) (*distrtypes.QueryDelegationTotalRewardsResponse, error) {
	res, err := c.DistrQueryClient.DelegationTotalRewards(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) DelegatorValidators(ctx context.Context, in *distrtypes.QueryDelegatorValidatorsRequest, opts ...grpc.CallOption) (*distrtypes.QueryDelegatorValidatorsResponse, error) {
	res, err := c.DistrQueryClient.DelegatorValidators(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) DelegatorWithdrawAddress(ctx context.Context, in *distrtypes.QueryDelegatorWithdrawAddressRequest, opts ...grpc.CallOption) (*distrtypes.QueryDelegatorWithdrawAddressResponse, error) {
	res, err := c.DistrQueryClient.DelegatorWithdrawAddress(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) CommunityPool(ctx context.Context, in *distrtypes.QueryCommunityPoolRequest, opts ...grpc.CallOption) (*distrtypes.QueryCommunityPoolResponse, error) {
	res, err := c.DistrQueryClient.CommunityPool(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO: Slashing Query
func (c GreenlandClient) SlashingParams(ctx context.Context, in *slashingtypes.QueryParamsRequest, opts ...grpc.CallOption) (*slashingtypes.QueryParamsResponse, error) {
	res, err := c.SlashingQueryClient.Params(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) SigningInfo(ctx context.Context, in *slashingtypes.QuerySigningInfoRequest, opts ...grpc.CallOption) (*slashingtypes.QuerySigningInfoResponse, error) {
	res, err := c.SlashingQueryClient.SigningInfo(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) SigningInfos(ctx context.Context, in *slashingtypes.QuerySigningInfosRequest, opts ...grpc.CallOption) (*slashingtypes.QuerySigningInfosResponse, error) {
	res, err := c.SlashingQueryClient.SigningInfos(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO: Staking Query
func (c GreenlandClient) Validators(ctx context.Context, in *stakingtypes.QueryValidatorsRequest, opts ...grpc.CallOption) (*stakingtypes.QueryValidatorsResponse, error) {
	res, err := c.StakingQueryClient.Validators(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Validator(ctx context.Context, in *stakingtypes.QueryValidatorRequest, opts ...grpc.CallOption) (*stakingtypes.QueryValidatorResponse, error) {
	res, err := c.StakingQueryClient.Validator(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) ValidatorDelegations(ctx context.Context, in *stakingtypes.QueryValidatorDelegationsRequest, opts ...grpc.CallOption) (*stakingtypes.QueryValidatorDelegationsResponse, error) {
	res, err := c.StakingQueryClient.ValidatorDelegations(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) ValidatorUnbondingDelegations(ctx context.Context, in *stakingtypes.QueryValidatorUnbondingDelegationsRequest, opts ...grpc.CallOption) (*stakingtypes.QueryValidatorUnbondingDelegationsResponse, error) {
	res, err := c.StakingQueryClient.ValidatorUnbondingDelegations(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Delegation(ctx context.Context, in *stakingtypes.QueryDelegationRequest, opts ...grpc.CallOption) (*stakingtypes.QueryDelegationResponse, error) {
	res, err := c.StakingQueryClient.Delegation(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) UnbondingDelegation(ctx context.Context, in *stakingtypes.QueryUnbondingDelegationRequest, opts ...grpc.CallOption) (*stakingtypes.QueryUnbondingDelegationResponse, error) {
	res, err := c.StakingQueryClient.UnbondingDelegation(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) DelegatorDelegations(ctx context.Context, in *stakingtypes.QueryDelegatorDelegationsRequest, opts ...grpc.CallOption) (*stakingtypes.QueryDelegatorDelegationsResponse, error) {
	res, err := c.StakingQueryClient.DelegatorDelegations(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) DelegatorUnbondingDelegations(ctx context.Context, in *stakingtypes.QueryDelegatorUnbondingDelegationsRequest, opts ...grpc.CallOption) (*stakingtypes.QueryDelegatorUnbondingDelegationsResponse, error) {
	res, err := c.StakingQueryClient.DelegatorUnbondingDelegations(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Redelegations(ctx context.Context, in *stakingtypes.QueryRedelegationsRequest, opts ...grpc.CallOption) (*stakingtypes.QueryRedelegationsResponse, error) {
	res, err := c.StakingQueryClient.Redelegations(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) DelegatorValidators(ctx context.Context, in *stakingtypes.QueryDelegatorValidatorsRequest, opts ...grpc.CallOption) (*stakingtypes.QueryDelegatorValidatorsResponse, error) {
	res, err := c.StakingQueryClient.DelegatorValidators(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) DelegatorValidator(ctx context.Context, in *stakingtypes.QueryDelegatorValidatorRequest, opts ...grpc.CallOption) (*stakingtypes.QueryDelegatorValidatorResponse, error) {
	res, err := c.StakingQueryClient.DelegatorValidator(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) HistoricalInfo(ctx context.Context, in *stakingtypes.QueryHistoricalInfoRequest, opts ...grpc.CallOption) (*stakingtypes.QueryHistoricalInfoResponse, error) {
	res, err := c.StakingQueryClient.HistoricalInfo(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Pool(ctx context.Context, in *stakingtypes.QueryPoolRequest, opts ...grpc.CallOption) (*stakingtypes.QueryPoolResponse, error) {
	res, err := c.StakingQueryClient.Pool(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) StakingParams(ctx context.Context, in *stakingtypes.QueryParamsRequest, opts ...grpc.CallOption) (*stakingtypes.QueryParamsResponse, error) {
	res, err := c.StakingQueryClient.Params(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO: Auth Query
func (c GreenlandClient) Accounts(ctx context.Context, in *authtypes.QueryAccountsRequest, opts ...grpc.CallOption) (*authtypes.QueryAccountsResponse, error) {
	res, err := c.AuthQueryClient.Accounts(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Account(ctx context.Context, in *authtypes.QueryAccountRequest, opts ...grpc.CallOption) (*authtypes.QueryAccountResponse, error) {
	res, err := c.AuthQueryClient.Account(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) AccountAddressByID(ctx context.Context, in *authtypes.QueryAccountAddressByIDRequest, opts ...grpc.CallOption) (*authtypes.QueryAccountAddressByIDResponse, error) {
	res, err := c.AuthQueryClient.AccountAddressByID(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Params(ctx context.Context, in *authtypes.QueryParamsRequest, opts ...grpc.CallOption) (*authtypes.QueryParamsResponse, error) {
	res, err := c.AuthQueryClient.Params(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) ModuleAccounts(ctx context.Context, in *authtypes.QueryModuleAccountsRequest, opts ...grpc.CallOption) (*authtypes.QueryModuleAccountsResponse, error) {
	res, err := c.AuthQueryClient.ModuleAccounts(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) ModuleAccountByName(ctx context.Context, in *authtypes.QueryModuleAccountByNameRequest, opts ...grpc.CallOption) (*authtypes.QueryModuleAccountByNameResponse, error) {
	res, err := c.AuthQueryClient.ModuleAccountByName(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Bech32Prefix(ctx context.Context, in *authtypes.Bech32PrefixRequest, opts ...grpc.CallOption) (*authtypes.Bech32PrefixResponse, error) {
	res, err := c.AuthQueryClient.Bech32Prefix(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) AddressBytesToString(ctx context.Context, in *authtypes.AddressBytesToStringRequest, opts ...grpc.CallOption) (*authtypes.AddressBytesToStringResponse, error) {
	res, err := c.AuthQueryClient.AddressBytesToString(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) AddressStringToBytes(ctx context.Context, in *authtypes.AddressStringToBytesRequest, opts ...grpc.CallOption) (*authtypes.AddressStringToBytesResponse, error) {
	res, err := c.AuthQueryClient.AddressStringToBytes(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO: Gov Query
func (c GreenlandClient) Proposal(ctx context.Context, in *v1beta1.QueryProposalRequest, opts ...grpc.CallOption) (*v1beta1.QueryProposalResponse, error) {
	res, err := c.GovQueryClient.Proposal(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Proposals(ctx context.Context, in *v1beta1.QueryProposalsRequest, opts ...grpc.CallOption) (*v1beta1.QueryProposalsResponse, error) {
	res, err := c.GovQueryClient.Proposals(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (c GreenlandClient) Vote(ctx context.Context, in *v1beta1.QueryVoteRequest, opts ...grpc.CallOption) (*v1beta1.QueryVoteResponse, error) {
	res, err := c.GovQueryClient.Vote(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Votes(ctx context.Context, in *v1beta1.QueryVotesRequest, opts ...grpc.CallOption) (*v1beta1.QueryVotesResponse, error) {
	res, err := c.GovQueryClient.Votes(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Params(ctx context.Context, in *v1beta1.QueryParamsRequest, opts ...grpc.CallOption) (*v1beta1.QueryParamsResponse, error) {
	res, err := c.GovQueryClient.Params(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Deposit(ctx context.Context, in *v1beta1.QueryDepositRequest, opts ...grpc.CallOption) (*v1beta1.QueryDepositResponse, error) {
	res, err := c.GovQueryClient.Deposit(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Deposits(ctx context.Context, in *v1beta1.QueryDepositsRequest, opts ...grpc.CallOption) (*v1beta1.QueryDepositsResponse, error) {
	res, err := c.GovQueryClient.Deposits(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) TallyResult(ctx context.Context, in *v1beta1.QueryTallyResultRequest, opts ...grpc.CallOption) (*v1beta1.QueryTallyResultResponse, error) {
	res, err := c.GovQueryClient.TallyResult(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//TODO: Authz Query
func (c GreenlandClient) Grants(ctx context.Context, in *authztypes.QueryGrantsRequest, opts ...grpc.CallOption) (*authztypes.QueryGrantsResponse, error) {
	res, err := c.AuthzQueryClient.Grants(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) GranterGrants(ctx context.Context, in *authztypes.QueryGranterGrantsRequest, opts ...grpc.CallOption) (*authztypes.QueryGranterGrantsResponse, error) {
	res, err := c.AuthzQueryClient.GranterGrants(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) GranteeGrants(ctx context.Context, in *authztypes.QueryGranteeGrantsRequest, opts ...grpc.CallOption) (*authztypes.QueryGranteeGrantsResponse, error) {
	res, err := c.AuthzQueryClient.GranteeGrants(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO: Feegrant query
func (c GreenlandClient) Allowance(ctx context.Context, in *feegranttypes.QueryAllowanceRequest, opts ...grpc.CallOption) (*feegranttypes.QueryAllowanceResponse, error) {
	res, err := c.FeegrantQueryClient.Allowance()l(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Allowances(ctx context.Context, in *feegranttypes.QueryAllowancesRequest, opts ...grpc.CallOption) (*feegranttypes.QueryAllowancesResponse, error) {
	res, err := c.FeegrantQueryClient.Allowances(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) AllowancesByGranter(ctx context.Context, in *feegranttypes.QueryAllowancesByGranterRequest, opts ...grpc.CallOption) (*feegranttypes.QueryAllowancesByGranterResponse, error) {
	res, err := c.FeegrantQueryClient.AllowancesByGranter(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO: Params Query
func (c GreenlandClient) Params(ctx context.Context, in *paramstypes.QueryParamsRequest) (*paramstypes.QueryParamsResponse, error) {
	res, err := c.ParamsQueryClient.Params(ctx, in)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c GreenlandClient) Subspaces(ctx context.Context, in *paramstypes.QuerySubspacesRequest) (*paramstypes.QuerySubspacesResponse, error) {
	res, err := c.ParamsQueryClient.Subspaces(ctx, in)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO: BFS Query
func (c GreenlandClient) Params(ctx context.Context, in *bfstypes.QueryParamsRequest, opts ...grpc.CallOption) (*bfstypes.QueryParamsResponse, error) {
	res, err := c.BfsQueryClient.Params(ctx, in, opts...)  
	if err != nil {
		return nil, err
	}
	return res, nil
}

