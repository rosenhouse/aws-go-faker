package awsfaker

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/private/protocol/xml/xmlutil"
	"github.com/rosenhouse/aws-go-faker/queryutil"
)

func writeResponse(w http.ResponseWriter, statusCode int, action string, data interface{}) {
	responseBuffer := &bytes.Buffer{}
	encoder := xml.NewEncoder(responseBuffer)
	resultWrapper := xml.StartElement{Name: xml.Name{Local: action + "Result"}}
	err := encoder.EncodeToken(resultWrapper)
	if err != nil {
		panic(err)
	}
	err = xmlutil.BuildXML(data, encoder)
	if err != nil {
		panic(err)
	}
	err = encoder.EncodeToken(resultWrapper.End())
	if err != nil {
		panic(err)
	}
	err = encoder.Flush()
	if err != nil {
		panic(err)
	}

	w.WriteHeader(statusCode)
	_, err = w.Write(responseBuffer.Bytes())
	if err != nil {
		panic(err)
	}
}

func writeError(w http.ResponseWriter, httpStatusCode int, errorResponse interface{}) {
	responseBodyBytes, err := xml.Marshal(errorResponse)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(httpStatusCode)
	_, err = w.Write(responseBodyBytes)
	if err != nil {
		panic(err)
	}
}

func parseQueryRequest(r *http.Request) (url.Values, error) {
	requestBodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read request body: %s", err)
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(requestBodyBytes))

	values, err := url.ParseQuery(string(requestBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("unable to parse request body as query syntax: %s", err)
	}

	return values, err
}

func methodIsEC2(method reflect.Value) bool {
	return strings.HasSuffix(method.Type().In(0).Elem().PkgPath(), "ec2")
}

func constructInput(method reflect.Value, queryValues url.Values) (interface{}, error) {
	inputValueType := method.Type().In(0).Elem()
	inputValue := reflect.New(inputValueType).Interface()
	isEC2 := methodIsEC2(method)
	queryutil.Decode(queryValues, inputValue, isEC2)
	return inputValue, nil
}
