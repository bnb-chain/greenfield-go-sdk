package bsc

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/bnb-chain/greenfield-go-sdk/bsctypes"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"google.golang.org/grpc"
)

// IClient - Declare all Greenfield SDK Client APIs, including APIs for interacting with Greenfield Blockchain and SPs.
type IClient interface {
	IMultiMessageClient
	IGreenfieldExecutorClient
	IBasicClient
}

// Client - The implementation for IClient, implement all Client APIs for Greenfield SDK.
type Client struct {
	// The chain Client is used to interact with the blockchain
	chainClient *ethclient.Client
	// The HTTP Client is used to send HTTP requests to the greenfield blockchain and sp
	httpClient *http.Client
	// The default account to use when sending transactions.
	defaultAccount *bsctypes.BscAccount
	// Whether the connection to the blockchain node is secure (HTTPS) or not (HTTP).
	secure bool
	// Host is the target sp server hostname，it is the host info in the request which sent to SP
	host string
	// The user agent info
	userAgent string
	// define if trace the error request to BSC
	isTraceEnabled   bool
	traceOutput      io.Writer
	onlyTraceError   bool
	useWebsocketConn bool
	expireSeconds    uint64
	rpcURL           string
	deployment       *bsctypes.Deployment
}

// Option - Configurations for providing optional parameters for the Binance Smart Chain SDK Client.
type Option struct {
	// GrpcDialOption is the list of gRPC dial options used to configure the connection to the blockchain node.
	GrpcDialOption grpc.DialOption
	// DefaultAccount is the default account of Client.
	DefaultAccount *bsctypes.BscAccount
	// Secure is a flag that specifies whether the Client should use HTTPS or not.
	Secure bool
	// Transport is the HTTP transport used to send requests to the storage provider endpoint.
	Transport http.RoundTripper
	// Host is the target sp server hostname.
	Host string
}

func New(rpcURL string, env string, option Option) (IClient, error) {
	if rpcURL == "" {
		return nil, errors.New("fail to get grpcAddress and chainID to construct Client")
	}
	var (
		cc         *ethclient.Client
		deployment *bsctypes.Deployment
		path       string
		err        error
	)
	cc, err = ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	switch env {
	case "dev-net":
		path = "./common/contract/dev-net.json"
	case "qa-net":
		path = "./common/contract/qa-net.json"
	case "test-net":
		path = "./common/contract/test-net.json"
	case "main-net":
		path = "./common/contract/main-net.json"
	}

	jsonFile, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open JSON file: %v", err)
	}
	defer jsonFile.Close()

	// Read the JSON file into a byte slice
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
		return nil, err
	}

	err = json.Unmarshal(byteValue, &deployment)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON data: %v", err)
		return nil, err
	}

	c := Client{
		chainClient:    cc,
		httpClient:     &http.Client{Transport: option.Transport},
		userAgent:      types.UserAgent,
		defaultAccount: option.DefaultAccount, // it allows to be nil
		secure:         option.Secure,
		host:           option.Host,
		rpcURL:         rpcURL,
		deployment:     deployment,
	}

	return &c, nil
}
