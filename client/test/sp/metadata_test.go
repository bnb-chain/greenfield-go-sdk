package sp

import (
	"context"
	"encoding/json"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	"net/http"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	spClient "github.com/bnb-chain/greenfield-go-sdk/client/sp"
)

func TestListObjects(t *testing.T) {
	setup()
	defer shutdown()

	bucketName := "test-bucket"

	var expectedRes spClient.ListObjectsResponse
	var objects []*types.Object
	object1 := types.Object{
		ObjectInfo: &types.ObjectInfo{
			Owner: "test-owner-object",
		},
	}
	objects = append(objects, &object1)
	expectedRes = spClient.ListObjectsResponse{Objects: objects}

	out, err := json.Marshal(expectedRes)

	if err != nil {
		log.Error().Msg("the marshal of expectedRes failed: " + err.Error())
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startHandle(t, r)
		testMethod(t, r, "GET")

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)

		_, err := w.Write(out)
		require.NoError(t, err)
	})

	body, err := client.ListObjects(context.Background(), bucketName, spClient.NewAuthInfo(false, ""))
	require.NoError(t, err)

	// check ListObjects content
	if body.Objects[0].ObjectInfo.Owner != expectedRes.Objects[0].ObjectInfo.Owner {
		t.Errorf("TestListObjects content not same")
	}

}

func TestListBuckets(t *testing.T) {
	setup()
	defer shutdown()

	var buckets []*types.Bucket
	bucket1 := types.Bucket{
		BucketInfo: &types.BucketInfo{
			Owner: "test-owner-bucket",
		},
	}
	buckets = append(buckets, &bucket1)
	expectedRes := spClient.ListBucketsResponse{Buckets: buckets}

	out, err := json.Marshal(expectedRes)
	if err != nil {
		log.Error().Msg("the marshal of expectedRes failed: " + err.Error())
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startHandle(t, r)
		testMethod(t, r, "GET")

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(200)

		_, err := w.Write(out)
		require.NoError(t, err)
	})

	body, err := client.ListBuckets(context.Background(), spClient.UserInfo{Address: "test-address"}, spClient.NewAuthInfo(false, ""))
	require.NoError(t, err)

	if body.Buckets[0].BucketInfo.Owner != expectedRes.Buckets[0].BucketInfo.Owner {
		t.Errorf("TestListBuckets content not same")
	}

}
