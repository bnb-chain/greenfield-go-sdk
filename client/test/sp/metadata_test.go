package sp

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/rs/zerolog/log"

	"github.com/stretchr/testify/require"

	spClient "github.com/bnb-chain/greenfield-go-sdk/client/sp"
	storageType "github.com/bnb-chain/greenfield/x/storage/types"
)

func TestListObjectsByBucketName(t *testing.T) {
	setup()
	defer shutdown()

	bucketName := "test-bucket"

	var expectedRes spClient.ListObjectsByBucketNameResponse
	var objects []*spClient.Object
	object1 := spClient.Object{
		ObjectInfo: &storageType.ObjectInfo{
			Owner: "test-owner-object",
		},
	}
	objects = append(objects, &object1)
	expectedRes = spClient.ListObjectsByBucketNameResponse{Objects: objects}

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
	log.Print(body)

	// check ListObjectsByBucketName content
	if body.Objects[0].ObjectInfo.Owner != expectedRes.Objects[0].ObjectInfo.Owner {
		t.Errorf("TestListObjectsByBucketName content not same")
	}

}

func TestListBucketsByUser(t *testing.T) {
	setup()
	defer shutdown()

	var buckets []*spClient.Bucket
	bucket1 := spClient.Bucket{
		BucketInfo: &storageType.BucketInfo{
			Owner: "test-owner-bucket",
			Id:    sdkmath.NewUint(1),
		},
	}
	buckets = append(buckets, &bucket1)
	expectedRes := spClient.ListBucketsByUserResponse{Buckets: buckets}

	out, err := json.Marshal(expectedRes)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startHandle(t, r)
		testMethod(t, r, "GET")

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)

		_, err := w.Write(out)
		require.NoError(t, err)
	})

	body, err := client.ListBucketsByUser(context.Background(), spClient.UserInfo{Address: "test-address"}, spClient.NewAuthInfo(false, ""))
	require.NoError(t, err)

	if body.Buckets[0].BucketInfo.Owner != expectedRes.Buckets[0].BucketInfo.Owner {
		t.Errorf("TestGetUserBuckets content not same")
	}

}
