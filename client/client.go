package client

import (
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

type GreenfieldClient struct {
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

func NewGreenfieldClient(grpcAddr string) (GreenfieldClient, error) {
	conn := grpcConn(grpcAddr)

	return GreenfieldClient{
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
