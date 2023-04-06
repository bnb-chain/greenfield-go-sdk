# Greenfield Go SDK


## Disclaimer
**The software and related documentation are under active development, all subject to potential future change without
notification and not ready for production use. The code and security audit have not been fully completed and not ready
for any bug bounty. We advise you to be careful and experiment on the network at your own risk. Stay safe out there.**

## Instruction

The Greenfield-GO-SDK provides a thin wrapper for interacting with greenfield. 

### Requirement

Go version above 1.18

## Usage

### Importing

```go
import (
    "github.com/bnb-chain/greenfield-go-sdk" latest
)
```

### Account

A Account is needed to sign request and transaction messages or verify signatures. 

```go
type Account struct {
	name string
    km   keys.KeyManager
}
```

We provide three construction functions to generate the Key Manager:

```go
func NewAccountFromPrivateKey(name, privKey string) (*Account, error)

func NewAccountFromMnemonic(name, mnemonic string) (*Account, error)
```

- NewAccountFromPrivateKey: You should provide a Hex encoded string of your private key.
- NewAccountFromMnemonic: You should provide your mnemonic, usually a string of 24 words.

Examples:

From private key hex string:
```GO
privateKey := "9579fff0cab07a4379e845a890105004ba4c8276f8ad9d22082b2acbf02d884b"
account, err := types.NewAccountFromPrivateKey(privateKey)
```

From mnemonic:
```Go
mnemonic := "dragon shy author wave swamp avoid lens hen please series heavy squeeze alley castle crazy action peasant green vague camp mirror amount person legal"
account, _ := types.NewAccountFromMnemonic(mnemonic)
```

### Client

#### Init client
```go
mnemonic := ParseValidatorMnemonic(0)
account, err := types.NewAccountFromMnemonic("test", mnemonic)
assert.NoError(t, err)
cli, err := client.New(ChainID, GrpcAddress, account, &client.Option{GrpcDialOption: grpc.WithTransportCredentials(insecure.NewCredentials())})
```

#### Interface


```go
type Client interface {
	Basic
	Bucket
	Object
	Group
	Challenge
	Account
}
```

Bucket-related interfaces

```go
type Bucket interface {
	GetCreateBucketApproval(ctx context.Context, createBucketMsg *storageTypes.MsgCreateBucket) (*storageTypes.MsgCreateBucket, error)
	CreateBucket(ctx context.Context, bucketName string, primaryAddr sdk.AccAddress, opts types.CreateBucketOptions) (string, error)
	DeleteBucket(ctx context.Context, bucketName string, opt types.DeleteBucketOption) (string, error)
	UpdateBucketVisibility(ctx context.Context, bucketName string, visibility storageTypes.VisibilityType, opt types.UpdateVisibilityOption) (string, error)
	GetBucketReadQuota(ctx context.Context, bucketName string) (types.QuotaInfo, error)
	HeadBucket(ctx context.Context, bucketName string) (*storageTypes.BucketInfo, error)
	HeadBucketByID(ctx context.Context, bucketID string) (*storageTypes.BucketInfo, error)
	// PutBucketPolicy apply bucket policy to the principal, return the txn hash
	PutBucketPolicy(ctx context.Context, bucketName string, principalStr types.Principal, statements []*permTypes.Statement, opt types.PutPolicyOption) (string, error)
	// DeleteBucketPolicy delete the bucket policy of the principal
	DeleteBucketPolicy(ctx context.Context, bucketName string, principalAddr sdk.AccAddress, opt types.DeletePolicyOption) (string, error)
	// GetBucketPolicy get the bucket policy info of the user specified by principalAddr
	GetBucketPolicy(ctx context.Context, bucketName string, principalAddr sdk.AccAddress) (*permTypes.Policy, error)
	ListBuckets(ctx context.Context, userInfo types.UserInfo) (types.ListBucketsResult, error)
	ListBucketReadRecord(ctx context.Context, bucketName string, opts types.ListReadRecordOptions) (types.QuotaRecordInfo, error)
}
```

Object-related interface

```go
type Object interface {
	GetCreateObjectApproval(ctx context.Context, createObjectMsg *storageTypes.MsgCreateObject) (*storageTypes.MsgCreateObject, error)
	CreateObject(ctx context.Context, bucketName, objectName string,
		reader io.Reader, opts types.CreateObjectOptions) (string, error)
	PutObject(ctx context.Context, bucketName, objectName, txnHash string, objectSize int64,
		reader io.Reader, opt types.PutObjectOption) error
	CancelCreateObject(ctx context.Context, bucketName, objectName string, opt types.CancelCreateOption) (string, error)
	DeleteObject(ctx context.Context, bucketName, objectName string, opt types.DeleteObjectOption) (string, error)
	GetObject(ctx context.Context, bucketName, objectName string, opts types.GetObjectOption) (io.ReadCloser, types.ObjectStat, error)
	// HeadObject query the objectInfo on chain to check th object id, return the object info if exists
	// return err info if object not exist
	HeadObject(ctx context.Context, bucketName, objectName string) (*storageTypes.ObjectInfo, error)
	// HeadObjectByID query the objectInfo on chain by object id, return the object info if exists
	// return err info if object not exist
	HeadObjectByID(ctx context.Context, objID string) (*storageTypes.ObjectInfo, error)
	// PutObjectPolicy apply object policy to the principal, return the txn hash
	PutObjectPolicy(ctx context.Context, bucketName, objectName string, principalStr types.Principal,
		statements []*permTypes.Statement, opt types.PutPolicyOption) (string, error)
	DeleteObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr sdk.AccAddress, opt types.DeletePolicyOption) (string, error)
	// GetObjectPolicy get the object policy info of the user specified by principalAddr
	GetObjectPolicy(ctx context.Context, bucketName, objectName string, principalAddr sdk.AccAddress) (*permTypes.Policy, error)
	ListObjects(ctx context.Context, bucketName string) (types.ListObjectsResult, error)
}
```