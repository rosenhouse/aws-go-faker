// Package query implements the AWS query protocol
package query

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
	"github.com/rosenhouse/awsfaker/protocols/query/queryutil"
)

// A Handler is an http.Handler that can mimic an AWS service API
type Handler struct {
	actions map[string]reflect.Value
}

func (h *Handler) registerService(awsService interface{}) {
	service := reflect.ValueOf(awsService)
	if !service.IsValid() {
		panic("invalid service interface")
	}
	if service.Kind() != reflect.Ptr {
		panic("expecting struct pointer as service interface")
	}
	if !service.Elem().IsValid() {
		panic("expectingn non-nil pointer as service interface")
	}
	serviceType := service.Type()
	n := service.NumMethod()
	if n == 0 {
		panic("no methods on service interface")
	}
	for i := 0; i < n; i++ {
		h.actions[serviceType.Method(i).Name] = service.Method(i)
	}
}

// New returns a new Handler that will dispatch incoming requests to
// one or more fake service backends given as arguments.
func New(serviceBackend interface{}) *Handler {
	handler := &Handler{make(map[string]reflect.Value)}
	handler.registerService(serviceBackend)
	return handler
}

func (f *Handler) findMethod(actionName string) (reflect.Value, error) {
	method, ok := f.actions[actionName]
	if !ok {
		return reflect.Value{}, fmt.Errorf("action %s not found, check that you've fully implemented your fake backend", actionName)
	}
	return method, nil
}

// ServeHTTP dispatches a request to a backend method and writes the response
func (f *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	queryValues, err := parseQueryRequest(r)
	if err != nil {
		panic(err)
	}
	methodName := queryValues.Get("Action")
	method, err := f.findMethod(methodName)
	if err != nil {
		panic(err)
	}

	input, err := constructInput(method, queryValues)
	if err != nil {
		panic(err)
	}

	results := method.Call([]reflect.Value{reflect.ValueOf(input)})
	errVal := results[1].Interface()
	if errVal != nil {
		errorResponse := errVal.(errorResponse)
		err := specializeErrorResponse(method, errorResponse)
		writeError(w, errorResponse.HTTPStatus(), err)
		return
	}

	outVal := results[0].Interface()
	if err != nil {
		panic(err)
	}
	writeResponse(w, http.StatusOK, methodName, outVal)
}

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
