package query

import (
	"encoding/json"
	"encoding/xml"
	"reflect"
)

type withHTTPCode interface {
	HTTPStatusCode() int
}

type queryErrorResponse struct {
	XMLName         xml.Name `xml:"ErrorResponse"`
	AWSErrorCode    string   `xml:"Error>Code"`
	AWSErrorMessage string   `xml:"Error>Message"`
	RequestID       string   `xml:"RequestId"`
	HttpStatusCode  int      `xml:"-"`
}

func (q queryErrorResponse) HTTPStatusCode() int { return q.HttpStatusCode }

type ec2ErrorResponse struct {
	XMLName         xml.Name `xml:"Response"`
	AWSErrorCode    string   `xml:"Errors>Error>Code"`
	AWSErrorMessage string   `xml:"Errors>Error>Message"`
	RequestID       string   `xml:"RequestId"`
	HttpStatusCode  int      `xml:"-"`
}

func (q ec2ErrorResponse) HTTPStatusCode() int { return q.HttpStatusCode }

func errCopy(src interface{}, dst interface{}) {
	srcBytes, err := json.Marshal(src)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(srcBytes, dst)
	if err != nil {
		panic(err)
	}
}

func specializeErrorResponse(method reflect.Value, genericError error) withHTTPCode {
	if methodIsEC2(method) {
		var specialized ec2ErrorResponse
		specialized.AWSErrorCode = "[awsfaker missing error code]"
		specialized.AWSErrorMessage = "[awsfaker missing error message]"
		errCopy(genericError, &specialized)
		return specialized
	} else {
		var specialized queryErrorResponse
		specialized.AWSErrorCode = "[awsfaker missing error code]"
		specialized.AWSErrorMessage = "[awsfaker missing error message]"
		errCopy(genericError, &specialized)
		return specialized
	}
}
