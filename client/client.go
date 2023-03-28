package client

import (
	"context"
	"io"

	"github.com/bnb-chain/greenfield-go-sdk/client/sp"
	"github.com/bnb-chain/greenfield-go-sdk/pkg/api"
)

type IClient interface {
	Bucket
	Object
	Group
	Account
	SP
}

type Bucket interface {
	CreateBucket(ctx context.Context, bucketName string, opts *CreateBucketOptions) (string, error)
}

type Object interface {
	CreateObject(ctx context.Context, bucketName, objectName string,
		reader io.Reader, opts *CreateObjectOptions) (string, error)
	PutObject(ctx context.Context, bucketName, objectName, txnHash string, objectSize int64,
		reader io.Reader, opt sp.PutObjectOption) (res sp.UploadResult, err error)
	DeleteObject()
}

type Group interface {
	CreateGroup() (string, error)
}

type SP interface {
	CreateStorageProvider()
	EditStorageProvider()
	Deposit()
	GrantDeposit()
	QueryStorageProviders()
	QueryStorageProvider()
	QueryParams()
}

type Account interface {
	Send()
	MultiSend()
	QueryBalances()
}

func New() (IClient, error) {
	c := api.Client{}
	return &c, nil
}
