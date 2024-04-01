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
	"regexp"
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

func GetSegmentSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64) int64 {
	segmentCount := GetSegmentCount(payloadSize, maxSegmentSize)
	if segmentCount == 1 {
		return int64(payloadSize)
	} else if segmentIdx == segmentCount-1 {
		return int64(payloadSize) - (int64(segmentCount)-1)*int64(maxSegmentSize)
	} else {
		return int64(maxSegmentSize)
	}
}

func GetECPieceSize(payloadSize uint64, segmentIdx uint32, maxSegmentSize uint64, dataChunkNum uint32) int64 {
	segmentSize := GetSegmentSize(payloadSize, segmentIdx, maxSegmentSize)

	pieceSize := segmentSize / int64(dataChunkNum)
	// EC padding will cause the EC pieces to have one extra byte if it cannot be evenly divided.
	// for example, the segment size is 15, the ec piece size should be 15/4 + 1 = 4
	if segmentSize > 0 && segmentSize%int64(dataChunkNum) != 0 {
		pieceSize++
	}

	return pieceSize
}

func GetSegmentCount(payloadSize uint64, maxSegmentSize uint64) uint32 {
	count := payloadSize / maxSegmentSize
	if payloadSize%maxSegmentSize > 0 {
		count++
	}
	return uint32(count)
}

func ParseRange(rangeStr string) (bool, int64, int64) {
	if rangeStr == "" {
		return false, -1, -1
	}
	rangeStr = strings.ToLower(rangeStr)
	rangeStr = strings.ReplaceAll(rangeStr, " ", "")
	if !strings.HasPrefix(rangeStr, "bytes=") {
		return false, -1, -1
	}
	rangeStr = rangeStr[len("bytes="):]
	if strings.HasSuffix(rangeStr, "-") {
		rangeStr = rangeStr[:len(rangeStr)-1]
		rangeStart, err := stringToInt64(rangeStr)
		if err != nil {
			return false, -1, -1
		}
		return true, rangeStart, -1
	}
	pair := strings.Split(rangeStr, "-")
	if len(pair) == 2 {
		rangeStart, err := stringToInt64(pair[0])
		if err != nil {
			return false, -1, -1
		}
		rangeEnd, err := stringToInt64(pair[1])
		if err != nil {
			return false, -1, -1
		}
		return true, rangeStart, rangeEnd
	}
	return false, -1, -1
}

// StringToUint64 converts string to uint64
func stringToUint64(str string) (uint64, error) {
	ui64, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return ui64, nil
}

// stringToInt64 converts string to int64
func stringToInt64(str string) (int64, error) {
	i64, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return i64, nil
}

func ReadFull(r io.Reader, buf []byte) (n int, err error) {
	// ReadFull reads exactly len(buf) bytes from r into buf.
	// It returns the number of bytes copied and an error if
	// fewer bytes were read. The error is EOF only if no bytes
	// were read. If an EOF happens after reading some but not
	// all the bytes, ReadFull returns ErrUnexpectedEOF.
	// On return, n == len(buf) if and only if err == nil.
	// If r returns an error having read at least len(buf) bytes,
	// the error is dropped.
	for n < len(buf) && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])
		// Some spurious io.Reader's return
		// io.ErrUnexpectedEOF when nn == 0
		// this behavior is undocumented
		// so we are on purpose not using io.ReadFull
		// implementation because this can lead
		// to custom handling, to avoid that
		// we simply modify the original io.ReadFull
		// implementation to avoid this issue.
		// io.ErrUnexpectedEOF with nn == 0 really
		// means that io.EOF
		if err == io.ErrUnexpectedEOF && nn == 0 {
			err = io.EOF
		}
		n += nn
	}
	if n >= len(buf) {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}

// CheckObjectName  This code block checks for unsupported or potentially risky formats in object names.
// The checks are essential for ensuring the security and compatibility of the object names within the system.
// 1. ".." in object names: Checked to prevent path traversal attacks which might access directories outside the intended scope.
// 2. Object name being "/": The root directory should not be used as an object name due to potential security risks and ambiguity.
// 3. "\\" in object names: Backslashes are checked because they are often not supported in UNIX-like file systems and can cause issues in path parsing.
// 4. SQL Injection patterns in object names: Ensures that the object name does not contain patterns that could be used for SQL injection attacks, maintaining the integrity of the database.
func CheckObjectName(objectName string) bool {
	if strings.Contains(objectName, "..") ||
		objectName == "/" ||
		strings.Contains(objectName, "\\") ||
		IsSQLInjection(objectName) {
		return false
	}
	return true
}

func IsSQLInjection(input string) bool {
	// define patterns that may indicate SQL injection, especially those with a semicolon followed by common SQL keywords
	patterns := []string{
		"(?i).*;.*select", // Matches any string with a semicolon followed by "select"
		"(?i).*;.*insert", // Matches any string with a semicolon followed by "insert"
		"(?i).*;.*update", // Matches any string with a semicolon followed by "update"
		"(?i).*;.*delete", // Matches any string with a semicolon followed by "delete"
		"(?i).*;.*drop",   // Matches any string with a semicolon followed by "drop"
		"(?i).*;.*alter",  // Matches any string with a semicolon followed by "alter"
		"/\\*.*\\*/",      // Matches SQL block comment
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, input)
		if err != nil {
			return false
		}
		if matched {
			return true
		}
	}

	return false
}
