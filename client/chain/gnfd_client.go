package chain

import (
	_ "encoding/json"
	"github.com/bnb-chain/greenfield-go-sdk/keys"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	bridgetypes "github.com/bnb-chain/greenfield/x/bridge/types"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthQueryClient = authtypes.QueryClient
type AuthzQueryClient = authztypes.QueryClient
type AuthzMsgClient = authztypes.MsgClient
type BankQueryClient = banktypes.QueryClient
type BankMsgClient = banktypes.MsgClient
type CrosschainQueryClient = crosschaintypes.QueryClient
type DistrQueryClient = distrtypes.QueryClient
type DistrMsgClient = distrtypes.MsgClient
type FeegrantQueryClient = feegranttypes.QueryClient
type FeegrantMsgClient = feegranttypes.MsgClient
type GashubQueryClient = gashubtypes.QueryClient
type PaymentQueryClient = paymenttypes.QueryClient
type PaymentMsgClient = paymenttypes.MsgClient
type SpQueryClient = sptypes.QueryClient
type SpMsgClient = sptypes.MsgClient
type BridgeQueryClient = bridgetypes.QueryClient
type BridgeMsgClient = bridgetypes.MsgClient
type StorageQueryClient = storagetypes.QueryClient
type StorageMsgClient = storagetypes.MsgClient
type GovQueryClient = v1beta1.QueryClient
type GovMsgClient = v1beta1.MsgClient
type OracleQueryClient = oracletypes.QueryClient
type OracleMsgClient = oracletypes.MsgClient
type ParamsQueryClient = paramstypes.QueryClient
type SlashingQueryClient = slashingtypes.QueryClient
type SlashingMsgClient = slashingtypes.MsgClient
type StakingQueryClient = stakingtypes.QueryClient
type StakingMsgClient = stakingtypes.MsgClient
type TxClient = tx.ServiceClient
type UpgradeQueryClient = upgradetypes.QueryClient

type GreenfieldClient struct {
	AuthQueryClient
	AuthzQueryClient
	AuthzMsgClient
	BankQueryClient
	BankMsgClient
	CrosschainQueryClient
	DistrQueryClient
	DistrMsgClient
	FeegrantQueryClient
	FeegrantMsgClient
	GashubQueryClient
	PaymentQueryClient
	PaymentMsgClient
	SpQueryClient
	SpMsgClient
	BridgeQueryClient
	BridgeMsgClient
	StorageQueryClient
	StorageMsgClient
	GovQueryClient
	GovMsgClient
	OracleQueryClient
	OracleMsgClient
	ParamsQueryClient
	SlashingQueryClient
	SlashingMsgClient
	StakingQueryClient
	StakingMsgClient
	TxClient
	UpgradeQueryClient
	keyManager keys.KeyManager
	chainId    string
	codec      *codec.ProtoCodec
}

func grpcConn(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	return conn
}

func NewGreenfieldClient(grpcAddr, chainId string) GreenfieldClient {
	conn := grpcConn(grpcAddr)
	cdc := types.Cdc()
	return GreenfieldClient{
		authtypes.NewQueryClient(conn),
		authztypes.NewQueryClient(conn),
		authztypes.NewMsgClient(conn),
		banktypes.NewQueryClient(conn),
		banktypes.NewMsgClient(conn),
		crosschaintypes.NewQueryClient(conn),
		distrtypes.NewQueryClient(conn),
		distrtypes.NewMsgClient(conn),
		feegranttypes.NewQueryClient(conn),
		feegranttypes.NewMsgClient(conn),
		gashubtypes.NewQueryClient(conn),
		paymenttypes.NewQueryClient(conn),
		paymenttypes.NewMsgClient(conn),
		sptypes.NewQueryClient(conn),
		sptypes.NewMsgClient(conn),
		bridgetypes.NewQueryClient(conn),
		bridgetypes.NewMsgClient(conn),
		storagetypes.NewQueryClient(conn),
		storagetypes.NewMsgClient(conn),
		v1beta1.NewQueryClient(conn),
		v1beta1.NewMsgClient(conn),
		oracletypes.NewQueryClient(conn),
		oracletypes.NewMsgClient(conn),
		paramstypes.NewQueryClient(conn),
		slashingtypes.NewQueryClient(conn),
		slashingtypes.NewMsgClient(conn),
		stakingtypes.NewQueryClient(conn),
		stakingtypes.NewMsgClient(conn),
		tx.NewServiceClient(conn),
		upgradetypes.NewQueryClient(conn),
		nil,
		chainId,
		cdc,
	}
}

func NewGreenfieldClientWithKeyManager(grpcAddr, chainId string, keyManager keys.KeyManager) GreenfieldClient {
	gnfdClient := NewGreenfieldClient(grpcAddr, chainId)
	gnfdClient.keyManager = keyManager
	return gnfdClient
}

func (c *GreenfieldClient) GetKeyManager() (keys.KeyManager, error) {
	if c.keyManager == nil {
		return nil, types.KeyManagerNotInitError
	}
	return c.keyManager, nil
}

func (c *GreenfieldClient) SetChainId(id string) {
	c.chainId = id
}

func (c *GreenfieldClient) GetChainId() (string, error) {
	if c.chainId == "" {
		return "", types.ChainIdNotSetError
	}
	return c.chainId, nil
}
