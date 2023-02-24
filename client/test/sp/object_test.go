package sp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	spClient "github.com/bnb-chain/greenfield-go-sdk/client/sp"
	"github.com/bnb-chain/greenfield-go-sdk/utils"
	"github.com/stretchr/testify/require"
)

func TestPutObject(t *testing.T) {
	setup()
	defer shutdown()

	bucketName := "testbucket"
	ObjectName := "testobject"

	reader := bytes.NewReader([]byte("test content of object"))
	length, err := utils.GetContentLength(reader)
	assert.NoError(t, err)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startHandle(t, r)
		testMethod(t, r, "PUT")
		testHeader(t, r, "Content-Type", spClient.ContentDefault)
		testHeader(t, r, "Content-Length", strconv.FormatInt(length, 10))
		testBody(t, r, "test content of object")
	})

	txnHash := "test hash"
	newReader := bytes.NewReader([]byte("test content of object"))

	_, err = client.PutObject(context.Background(), bucketName,
		ObjectName, txnHash, length, newReader, spClient.NewAuthInfo(false, ""), spClient.UploadOption{})
	require.NoError(t, err)
}

func TestFPutObject(t *testing.T) {
	setup()
	defer shutdown()

	bucketName := "testbucket"
	ObjectName := "testobject"
	filePath := "./object_test.go"
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startHandle(t, r)
		testMethod(t, r, "PUT")
		testHeader(t, r, "Content-Type", spClient.ContentDefault)

		fileReader, err := os.Open(filePath)
		require.NoError(t, err)
		defer fileReader.Close()

		length, err := utils.GetContentLength(fileReader)
		require.NoError(t, err)
		testHeader(t, r, "Content-Length", strconv.FormatInt(length, 10))
	})

	txnHash := "test hash"

	_, err := client.FPutObject(context.Background(), bucketName,
		ObjectName, filePath, txnHash, spClient.ContentDefault, spClient.NewAuthInfo(false, ""))
	require.NoError(t, err)
}

func TestGetObject(t *testing.T) {
	setup()
	defer shutdown()

	bucketName := "test-bucket"
	ObjectName := "test-object"

	bodyContent := "test content of object"
	etag := "test etag"
	size := int64(len(bodyContent))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startHandle(t, r)
		testMethod(t, r, "GET")

		w.Header().Set("Etag", etag)
		w.Header().Set("Content-Type", "text/plain")
		s := strconv.FormatInt(size, 10) // s == "97" (decimal)
		w.Header().Set(spClient.HTTPHeaderContentLength, s)
		w.WriteHeader(200)

		if r.Header.Get("Range") != "" {
			w.Write([]byte(bodyContent)[1:10])
		} else {
			w.Write([]byte(bodyContent))
		}
	})

	body, info, err := client.GetObject(context.Background(), bucketName, ObjectName, spClient.DownloadOption{}, spClient.NewAuthInfo(false, ""))
	require.NoError(t, err)

	buf := new(strings.Builder)
	io.Copy(buf, body)
	// check download content
	if buf.String() != bodyContent {
		t.Errorf("download content not same")
	}
	// check etag
	if info.Etag != etag {
		t.Errorf("etag error")
		fmt.Println("etag", info.Etag)
	}

	if info.Size != size {
		t.Errorf("size error")
	}

	option := spClient.DownloadOption{}
	option.SetRange(1, 10)
	part_data, _, err := client.GetObject(context.Background(), bucketName, ObjectName, option, spClient.NewAuthInfo(false, ""))
	require.NoError(t, err)

	buf = new(strings.Builder)
	io.Copy(buf, part_data)
	// check download content
	if buf.String() != bodyContent[1:10] {
		t.Errorf("download range fail")
	}

}
