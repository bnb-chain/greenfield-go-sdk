package sp

import (
	"context"
	"encoding/json"
	storageType "github.com/bnb-chain/greenfield/x/storage/types"
	"net/http"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	spClient "github.com/bnb-chain/greenfield-go-sdk/client/sp"
)

func TestListObjectsByBucketName(t *testing.T) {
	setup()
	defer shutdown()

	bucketName := "test-bucket"

	var expectedRes spClient.ListObjectsByBucketNameResult
	var objects []storageType.ObjectInfo
	object1 := storageType.ObjectInfo{
		Owner: "test-owner-object",
	}
	objects = append(objects, object1)
	expectedRes = spClient.ListObjectsByBucketNameResult{Objects: objects}

	out, err := json.Marshal(expectedRes)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startHandle(t, r)
		testMethod(t, r, "GET")

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)

		_, err := w.Write(out)
		require.NoError(t, err)
	})

	body, err := client.ListObjectsByBucketName(context.Background(), bucketName, spClient.NewAuthInfo(false, ""))
	require.NoError(t, err)

	// check ListObjectsByBucketName content
	if body.Objects[0].Owner != expectedRes.Objects[0].Owner {
		t.Errorf("TestGetUserBuckets content not same")
	}

}

func TestGetUserBuckets(t *testing.T) {
	setup()
	defer shutdown()

	var buckets []storageType.BucketInfo
	bucket1 := storageType.BucketInfo{
		Owner: "test-owner-bucket",
		Id:    sdkmath.NewUint(1),
	}
	buckets = append(buckets, bucket1)
	expectedRes := spClient.GetUserBucketsResult{Buckets: buckets}

	out, err := json.Marshal(expectedRes)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startHandle(t, r)
		testMethod(t, r, "GET")

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)

		_, err := w.Write(out)
		require.NoError(t, err)
	})

	body, err := client.GetUserBuckets(context.Background(), spClient.MetadataInfo{Address: "test-address"}, spClient.NewAuthInfo(false, ""))
	require.NoError(t, err)

	if body.Buckets[0].Owner != expectedRes.Buckets[0].Owner {
		t.Errorf("TestGetUserBuckets content not same")
	}

}
