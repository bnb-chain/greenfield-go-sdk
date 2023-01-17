package inscription_go_sdk

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
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	libclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"google.golang.org/grpc"
)

type GreenlandClient struct {
	upgradetypes.QueryClient
	distrtypes.MsgClient
	slashingtypes.QueryClient
	slashingtypes.MsgClient
	stakingtypes.QueryClient
	stakingtypes.MsgClient
	authtypes.QueryClient
	banktypes.QueryClient
	banktypes.MsgClient
	v1beta1.QueryClient
	v1beta1.MsgClient
	authztypes.QueryClient
	authztypes.MsgClient
	feegranttypes.QueryClient
	feegranttypes.MsgClient
	paramstypes.QueryClient
	bfstypes.QueryClient
	bfstypes.MsgClient
}

func grpcConn(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return conn
}

func NewRpcClient(addr string) *rpchttp.HTTP {
	httpClient, err := libclient.DefaultHTTPClient(addr)
	if err != nil {
		panic(err)
	}
	rpcClient, err := rpchttp.NewWithClient(addr, "/websocket", httpClient)
	if err != nil {
		panic(err)
	}
	return rpcClient
}

func NewGreenlandClient(rpcAddrs, grpcAddrs []string) (GreenlandClient, error) {
	conn := grpcConn()

	return GreenlandClient{
		upgradetypes.NewQueryClient(conn),
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

//
//func NewInscriptionClient(baseUrl string, network types.ChainNetwork, keyManager keys.KeyManager) (InscriptionClient, error) {
//	types.SetNetwork(network)
//	baseClient := basic.NewClient(baseUrl, "")
//	queryClient := query.NewClient(base)
//	node, err := q.GetNodeInfo()
//	if err != nil {
//		return nil, err
//	}
//	txClient := transaction.NewClient(node.NodeInfo.Network, keyManager, queryClient, baseClient)
//	return &inscriptionClient{
//		BasicClient:       baseClient,
//		QueryClient:       queryClient,
//		TransactionClient: txClient,
//	}, nil
//}
