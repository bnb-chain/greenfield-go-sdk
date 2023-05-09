package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/rs/zerolog/log"
)

var EmptyURL = url.URL{}

func IsIPValid(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsDomainNameValid validates if input string is a valid domain name.
func IsDomainNameValid(hostName string) bool {
	// See RFC 1035, RFC 3696.
	hostName = strings.TrimSpace(hostName)
	if len(hostName) == 0 || len(hostName) > 255 {
		return false
	}
	if hostName[len(hostName)-1:] == "-" || hostName[:1] == "-" {
		return false
	}
	if hostName[len(hostName)-1:] == "_" || hostName[:1] == "_" {
		return false
	}
	if hostName[:1] == "." {
		return false
	}

	if strings.ContainsAny(hostName, "`~!@#$%^&*()+={}[]|\\\"';:><?/") {
		return false
	}
	return true
}

// GetEndpointURL - constructs a new endpoint.
func GetEndpointURL(endpoint string, secure bool) (*url.URL, error) {
	// If secure is false, use 'http' scheme.
	scheme := "https"
	if !secure {
		scheme = "http"
	}

	if strings.Contains(endpoint, "http") {
		s := strings.Split(endpoint, "//")
		endpoint = s[1]
	}
	// Construct a secured endpoint URL.
	endpointURLStr := scheme + "://" + endpoint
	endpointURL, err := url.Parse(endpointURLStr)
	if err != nil {
		return nil, err
	}
	// check endpoint if it is valid
	if err := checkEndpointUrl(*endpointURL); err != nil {
		return nil, err
	}
	return endpointURL, nil
}

// checkEndpointUrl verifies if endpoint url is valid, and return error
func checkEndpointUrl(endpointURL url.URL) error {
	if endpointURL == EmptyURL {
		return errors.New("Endpoint url is empty.")
	}

	if endpointURL.Path != "/" && endpointURL.Path != "" {
		return errors.New("Endpoint paths invalid")
	}

	host := endpointURL.Hostname()
	if !IsIPValid(host) && !IsDomainNameValid(host) {
		msg := endpointURL.Host + " does not meet ip address or domain name standards"
		return errors.New(msg)
	}

	return nil
}

// CalcSHA256Hex computes checksum of sha256 hash and encode it to hex
func CalcSHA256Hex(buf []byte) (hexStr string) {
	sum := CalcSHA256(buf)
	hexStr = hex.EncodeToString(sum)
	return
}

// CalcSHA256 computes checksum of sha256 from byte array
func CalcSHA256(buf []byte) []byte {
	h := sha256.New()
	h.Write(buf)
	sum := h.Sum(nil)
	return sum[:]
}

func DecodeURIComponent(s string) (string, error) {
	decodeStr, err := url.QueryUnescape(s)
	if err != nil {
		return s, err
	}
	return decodeStr, err
}

// AddQueryValues adds queryValue to url
func AddQueryValues(s string, qs url.Values) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	q := u.RawQuery
	rq := qs.Encode()
	if q != "" {
		if rq != "" {
			u.RawQuery = fmt.Sprintf("%s&%s", q, qs.Encode())
		}
	} else {
		u.RawQuery = rq
	}
	return u.String(), nil
}

// CloseResponse closes the response body
func CloseResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_, err := io.Copy(io.Discard, resp.Body)
		if err != nil {
			log.Info().Msg("close resp copy error" + err.Error())
		}
		resp.Body.Close()
	}
}

// GetContentLength returns the size of reader
func GetContentLength(reader io.Reader) (int64, error) {
	var contentLength int64
	var err error
	switch v := reader.(type) {
	case *bytes.Buffer:
		contentLength = int64(v.Len())
	case *bytes.Reader:
		contentLength = int64(v.Len())
	case *strings.Reader:
		contentLength = int64(v.Len())
	case *os.File:
		fInfo, fError := v.Stat()
		if fError != nil {
			err = fmt.Errorf("can't get reader content length,%s", fError.Error())
		} else {
			contentLength = fInfo.Size()
		}
	default:
		err = fmt.Errorf("can't get reader content length,unkown reader type")
	}
	return contentLength, err
}

// StringToUint64 converts string to uint64
func StringToUint64(str string) (uint64, error) {
	ui64, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return ui64, nil
}

// hasInvalidPath checks if the given path contains "." or ".." as path segments.
func hasInvalidPath(path string) bool {
	path = filepath.ToSlash(strings.TrimSpace(path))
	for _, p := range strings.Split(path, "/") {
		// Check for special characters "." and ".." which have special meaning in file systems.
		// In object storage systems, these characters should not be used as they can cause confusion.
		switch strings.TrimSpace(p) {
		case ".":
			return true
		case "..":
			return true
		}
	}
	return false
}

// IsValidObjectPrefix checks if the given object prefix is valid:
// - does not have invalid path segments
// - is a valid UTF-8 string
// - does not contain double slashes "//"
func IsValidObjectPrefix(prefix string) bool {
	if hasInvalidPath(prefix) {
		return false
	}
	if !utf8.ValidString(prefix) {
		return false
	}
	if strings.Contains(prefix, `//`) {
		return false
	}
	return true
}
