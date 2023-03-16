package sp

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/utils"
)

// QuotaInfo indicates the quota info of bucket
type QuotaInfo struct {
	XMLName             xml.Name `xml:"GetReadQuotaResult"`
	Version             string   `xml:"version,attr"`
	BucketName          string   `xml:"BucketName"`
	BucketID            string   `xml:"BucketID"`
	ReadQuotaSize       uint64   `xml:"ReadQuotaSize"`       // the bucket read quota value on chain
	SPFreeReadQuotaSize uint64   `xml:"SPFreeReadQuotaSize"` // the free quota of this month
	ReadConsumedSize    uint64   `xml:"ReadConsumedSize"`    // the consumed read quota of this month
}

type ReadRecord struct {
	XMLName            xml.Name `xml:"ReadRecord"`
	ObjectName         string   `xml:"ObjectName"`
	ObjectID           string   `xml:"ObjectID"`
	ReadAccountAddress string   `xml:"ReadAccountAddress"`
	ReadTimestampUs    int64    `xml:"ReadTimestampUs"`
	ReadSize           uint64   `xml:"ReadSize"`
}

// QuotaRecordInfo indicates the quota read record
type QuotaRecordInfo struct {
	XMLName              xml.Name     `xml:"GetBucketReadQuotaResult"`
	Version              string       `xml:"version,attr"`
	NextStartTimestampUs int64        `xml:"NextStartTimestampUs"`
	ReadRecords          []ReadRecord `xml:"ReadRecord"`
}

// ListReadRecordOption indicates the record timestamp which start search from
type ListReadRecordOption struct {
	StartTimeStamp int64
}

// GetBucketReadQuota return the bucket quota info of this month
func (c *SPClient) GetBucketReadQuota(ctx context.Context, bucketName string, authInfo AuthInfo) (QuotaInfo, error) {
	year, month, _ := time.Now().Date()

	var date string
	if int(month) < 10 {
		date = strconv.Itoa(year) + "-" + "0" + strconv.Itoa(int(month))
	} else {
		date = strconv.Itoa(year) + "-" + strconv.Itoa(int(month))
	}

	params := url.Values{}
	params.Add("read-quota", "")
	params.Add("year-month", date)

	reqMeta := requestMeta{
		urlValues:     params,
		bucketName:    bucketName,
		contentSHA256: EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		return QuotaInfo{}, err
	}
	defer utils.CloseResponse(resp)

	QuotaResult := QuotaInfo{}
	// decode the xml content from response body
	err = xml.NewDecoder(resp.Body).Decode(&QuotaResult)
	if err != nil {
		return QuotaInfo{}, err
	}

	return QuotaResult, nil
}

// ListBucketReadRecord return the read record of this month, the return items should be no more than maxRecords
// ListReadRecordOption indicates the start timestamp of return read records
func (c *SPClient) ListBucketReadRecord(ctx context.Context, bucketName string, maxRecords int, opt ListReadRecordOption, authInfo AuthInfo) (QuotaRecordInfo, error) {
	timeNow := time.Now()
	timeToday := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location()) // 获取当天0点时间 time类型
	if opt.StartTimeStamp < 0 {
		return QuotaRecordInfo{}, errors.New("start timestamp  less than 0")
	}
	var startTimeStamp int64
	if opt.StartTimeStamp == 0 {
		startTimeStamp = timeToday.AddDate(0, 0, -timeToday.Day()+1).UnixMicro()
	} else {
		startTimeStamp = opt.StartTimeStamp
	}
	timeMonthEnd := timeToday.AddDate(0, 1, -timeToday.Day()+1).UnixMicro()

	if timeMonthEnd < startTimeStamp {
		return QuotaRecordInfo{}, errors.New("start timestamp larger than the end timestamp of this month")
	}

	params := url.Values{}
	params.Set("list-read-record", "")
	params.Set("max-records", strconv.Itoa(maxRecords))
	params.Set("start-timestamp", strconv.FormatInt(startTimeStamp, 10))
	params.Set("end-timestamp", strconv.FormatInt(timeMonthEnd, 10))

	reqMeta := requestMeta{
		urlValues:     params,
		bucketName:    bucketName,
		contentSHA256: EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		return QuotaRecordInfo{}, err
	}
	defer utils.CloseResponse(resp)

	QuotaRecords := QuotaRecordInfo{}
	// decode the xml content from response body
	err = xml.NewDecoder(resp.Body).Decode(&QuotaRecords)
	if err != nil {
		return QuotaRecordInfo{}, err
	}

	return QuotaRecords, nil
}
