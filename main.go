package inscription_go_sdk

import (
	_ "encoding/json"
	bfstypes "github.com/bnb-chain/bfs/x/bfs/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	//genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	libclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"google.golang.org/grpc"
)

type InscriptionClient struct {
	upgradeQueryClient  upgradetypes.QueryClient
	distrMsgClient      distrtypes.MsgClient
	slashingQueryClient slashingtypes.QueryClient
	slashingMsgClient   slashingtypes.MsgClient
	stakingQueryClient  stakingtypes.QueryClient
	stakingMsgClient    stakingtypes.MsgClient
	authQueryClient     authtypes.QueryClient
	bankQueryClient     banktypes.QueryClient
	bankMsgClient       banktypes.MsgClient
	govQueryClient      v1beta1.QueryClient
	govMsgClient        v1beta1.MsgClient
	authzQueryClient    authztypes.QueryClient
	authzMsgClient      authztypes.MsgClient
	feegrantQueryClient feegranttypes.QueryClient
	feegrantMsgClient   feegranttypes.MsgClient
	paramsClient        paramstypes.QueryClient
	bfsQueryClient      bfstypes.QueryClient
	bfsMsgClient        bfstypes.MsgClient
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

func NewInscriptionClient(rpcAddrs, grpcAddrs []string) (InscriptionClient, error) {
	conn := grpcConn()

	return InscriptionClient{
		upgradeQueryClient:  upgradetypes.NewQueryClient(conn),
		distrMsgClient:      distrtypes.NewMsgClient(conn),
		slashingQueryClient: slashingtypes.NewQueryClient(conn),
		slashingMsgClient:   slashingtypes.NewMsgClient(conn),
		stakingQueryClient:  stakingtypes.NewQueryClient(conn),
		stakingMsgClient:    stakingtypes.NewMsgClient(conn),
		authQueryClient:     authtypes.NewQueryClient(conn),
		bankQueryClient:     banktypes.NewQueryClient(conn),
		bankMsgClient:       banktypes.NewMsgClient(conn),
		govQueryClient:      v1beta1.NewQueryClient(conn),
		govMsgClient:        v1beta1.NewMsgClient(conn),
		authzQueryClient:    authztypes.NewQueryClient(conn),
		authzMsgClient:      authztypes.NewMsgClient(conn),
		feegrantQueryClient: feegranttypes.NewQueryClient(conn),
		feegrantMsgClient:   feegranttypes.NewMsgClient(conn),
		paramsClient:        paramstypes.NewQueryClient(conn),
		bfsQueryClient:      bfstypes.NewQueryClient(conn),
		bfsMsgClient:        bfstypes.NewMsgClient(conn),
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
