package sp

import (
	"context"
	"net/http"
	"testing"

	spClient "github.com/bnb-chain/greenfield-go-sdk/client/sp"
	storage_type "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/stretchr/testify/require"
)

// TestCreateBucket test creating a new bucket
func TestCreateBucket(t *testing.T) {
	setup()
	defer shutdown()

	bucketName := "testbucket"
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startHandle(t, r)
		testMethod(t, r, "GET")
		testHeader(t, r, spClient.HTTPHeaderContentSHA256, spClient.EmptyStringSHA256)

		msg := r.Header.Get(spClient.HTTPHeaderUnsignedMsg)
		w.Header().Set(spClient.HTTPHeaderSignedMsg, msg)
		w.WriteHeader(200)
	})

	_, _, testAddr := testdata.KeyEthSecp256k1TestPubAddr()
	createBucketMsg := storage_type.NewMsgCreateBucket(client.GetAccount(), bucketName, true, testAddr, nil, 0, nil)

	err := createBucketMsg.ValidateBasic()
	require.NoError(t, err)

	// test preCreateBucket
	_, err = client.GetCreateBucketApproval(context.Background(), createBucketMsg, spClient.NewAuthInfo(false, ""))
	require.NoError(t, err)
}
