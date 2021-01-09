package njson

import (
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/valyala/fastjson"
)

func benchmark(src string) func(b *testing.B) {
	p := Document{}
	return func(b *testing.B) {
		n, tail, err := p.Parse(src)
		if err != nil {
			b.Errorf("Parse error: %s", err)
			return
		}
		if strings.TrimSpace(tail) != "" {
			b.Errorf("Non empty tail: %q", tail)
			return
		}
		if n.value() == nil {
			b.Errorf("Nil root")
			return
		}
		b.ReportAllocs()
		b.SetBytes(int64(len(src)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			p.Reset()
			p.Parse(src)
		}
	}
}
func BenchmarkParse(b *testing.B) {
	b.Run("small.json", benchmark(smallJSON))
	b.Run("medium.min.json", benchmark(mediumJSON))
	b.Run("medium.json", benchmark(mediumJSONFormatted))
	b.Run("large.json", benchmark(largeJSON))
	b.Run("twitter.json", benchmark(twitterJSON))
	b.Run("canada.json", benchmark(canadaJSON))
	b.Run("aws.json", benchmark(awsJSON))
	b.Run("fastjson-aws.json", func(b *testing.B) {
		p := fastjson.Parser{}
		b.ReportAllocs()
		b.SetBytes(int64(len(awsJSON)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v, err := p.Parse(awsJSON)
			if err != nil {
				b.Error(err)
			}
			_ = v
		}
	})
	b.Run("encoding_json-aws.json", func(b *testing.B) {
		data := []byte(awsJSON)
		var records struct {
			Records []awsCloudTrail
		}
		b.ReportAllocs()
		b.SetBytes(int64(len(awsJSON)))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := json.Unmarshal(data, &records); err != nil {
				b.Error(err)
			}
		}

	})
	b.Run("jsoniter-aws.json", func(b *testing.B) {
		data := []byte(awsJSON)
		var records struct {
			Records []awsCloudTrail
		}
		iter := jsoniter.ConfigDefault.BorrowIterator(data)
		b.ReportAllocs()
		b.SetBytes(int64(len(awsJSON)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			iter.ResetBytes(data)
			iter.ReadVal(&records)
			if err := iter.Error; err != nil {
				b.Error(err)
			}
		}
	})
}

var (
	largeJSON           string
	mediumJSON          string
	mediumJSONFormatted string
	smallJSON           string
	twitterJSON         string
	canadaJSON          string
	awsJSON             = `{"Records": [{"eventVersion":"1.05","userIdentity":{"type":"AWSService","invokedBy":"cloudtrail.amazonaws.com"},"eventTime":"2018-08-26T14:17:23Z","eventSource":"kms.amazonaws.com","eventName":"GenerateDataKey","awsRegion":"us-west-2","sourceIPAddress":"cloudtrail.amazonaws.com","userAgent":"cloudtrail.amazonaws.com","requestParameters":{"keySpec":"AES_256","encryptionContext":{"aws:cloudtrail:arn":"arn:aws:cloudtrail:us-west-2:888888888888:trail/panther-lab-cloudtrail","aws:s3:arn":"arn:aws:s3:::panther-lab-cloudtrail/AWSLogs/888888888888/CloudTrail/us-west-2/2018/08/26/888888888888_CloudTrail_us-west-2_20180826T1410Z_inUwlhwpSGtlqmIN.json.gz"},"keyId":"arn:aws:kms:us-west-2:888888888888:key/72c37aae-1000-4058-93d4-86374c0fe9a0"},"responseElements":null,"requestID":"3cff2472-5a91-4bd9-b6d2-8a7a1aaa9086","eventID":"7a215e16-e0ad-4f6c-82b9-33ff6bbdedd2","readOnly":true,"resources":[{"ARN":"arn:aws:kms:us-west-2:888888888888:key/72c37aae-1000-4058-93d4-86374c0fe9a0","accountId":"888888888888","type":"AWS::KMS::Key"}],"eventType":"AwsApiCall","recipientAccountId":"777777777777","sharedEventID":"238c190c-1a30-4756-8e08-19fc36ad1b9f"}]}`
)

func init() {
	if data, err := ioutil.ReadFile("./testdata/large.min.json"); err != nil {
		panic(err)
	} else {
		largeJSON = string(data)
	}
	if data, err := ioutil.ReadFile("./testdata/medium.min.json"); err != nil {
		panic(err)
	} else {
		mediumJSON = string(data)
	}
	if data, err := ioutil.ReadFile("./testdata/small.json"); err != nil {
		panic(err)
	} else {
		smallJSON = string(data)
	}
	if data, err := ioutil.ReadFile("./testdata/medium.json"); err != nil {
		panic(err)
	} else {
		mediumJSONFormatted = string(data)
	}
	if data, err := ioutil.ReadFile("./testdata/twitter.json"); err != nil {
		panic(err)
	} else {
		twitterJSON = string(data)
	}
	if data, err := ioutil.ReadFile("./testdata/canada.json"); err != nil {
		panic(err)
	} else {
		canadaJSON = string(data)
	}
}

type awsCloudTrail struct {
	EventVersion string `json:"eventVersion"`
	UserIdentity *struct {
		Type      string `json:"type"`
		InvokedBy string `json:"invokedBy"`
	}
	EventTime         string `json:"eventTime"`
	EventSource       string `json:"eventSource"`
	EventName         string `json:"eventName"`
	AWSRegion         string `json:"awsRegion"`
	SourceIPAddress   string `json:"sourceIPAddress"`
	UserAgent         string `json:"userAgent"`
	RequestID         string `json:"requestID"`
	RequestParameters *struct {
		KeySpec           string `json:"keySpec"`
		EncryptionContext map[string]string
		KeyID             string `json:"keyId"`
	} `json:"requestParameters"`
	Resources []struct {
		ARN       string
		AccountID string `json:"accountId"`
		Type      string `json:"type"`
	} `json:"resources"`
	EventID            string `json:"eventID"`
	ReadOnly           bool   `json:"readOnly"`
	EventType          string `json:"eventType"`
	RecipientAccountID string `json:"recipientAccountId"`
	SharedEventID      string `json:"sharedEventID"`
}
