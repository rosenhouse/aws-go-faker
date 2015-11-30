package query

import (
	"encoding/xml"
	"reflect"
)

type queryErrorResponse struct {
	XMLName   xml.Name `xml:"ErrorResponse"`
	Code      string   `xml:"Error>Code"`
	Message   string   `xml:"Error>Message"`
	RequestID string   `xml:"RequestId"`
}

type ec2ErrorResponse struct {
	XMLName   xml.Name `xml:"Response"`
	Code      string   `xml:"Errors>Error>Code"`
	Message   string   `xml:"Errors>Error>Message"`
	RequestID string   `xml:"RequestId"`
}

type errorResponse interface {
	HTTPStatus() int
	AWSCode() string
	AWSMessage() string
}

func specializeErrorResponse(method reflect.Value, err errorResponse) interface{} {
	if methodIsEC2(method) {
		return ec2ErrorResponse{
			Code:    err.AWSCode(),
			Message: err.AWSMessage(),
		}
	} else {
		return queryErrorResponse{
			Code:    err.AWSCode(),
			Message: err.AWSMessage(),
		}
	}
}
