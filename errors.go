package awsfaker

import (
	"encoding/xml"
	"reflect"
)

type cloudFormationErrorResponse struct {
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

func specializeErrorResponse(method reflect.Value, err *ErrorResponse) interface{} {
	if methodIsEC2(method) {
		return ec2ErrorResponse{
			Code:    err.AWSErrorCode,
			Message: err.AWSErrorMessage,
		}
	} else {
		return cloudFormationErrorResponse{
			Code:    err.AWSErrorCode,
			Message: err.AWSErrorMessage,
		}
	}
}
