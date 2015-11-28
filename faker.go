// Package awsfaker supports the creation of test doubles for AWS APIs.
//
// Integration testing of applications that use AWS can be difficult.
// A test suite that interacts with a live AWS account will provide good
// test coverage, but may be slow and expensive.
//
// An alternative is to create a test double or "fake" of the AWS APIs that
// your application uses.  The fake boots an HTTP server that stands in for
// the real AWS endpoints, recording requests and providing arbitrary responses.
//
// This package provides a generic HTTP handler that can form the front-end
// of a test double (mock, fake or stub) for an AWS API.  To use it in tests,
// build a "backend" that implements the subset of the AWS API used by your code.
// Each API call should be implemented as a backend method like this
//			func (b *MyBackend) SomeAction(input *service.SomeActionInput) (*service.SomeActionOutput, error)
// Then initialize an HTTP server
//			myBackend := &MyBackend{ ... }
//			fakeServer := httptest.NewServer(awsfaker.New(myBackend))
// Finally, initialize your code under test, overriding the default AWS endpoint
// to instead use your fake:
//      app := myapp.App{ AWSOverride: fakeServer.URL }
//      app.Run()
//
// The method signatures of the backend exactly match those of the service
// interfaces of the package github.com/aws/aws-sdk-go
//
// For example, a complete implementation of AWS CloudFormation would match the
// CloudFormationAPI interface here: https://godoc.org/github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface
//
// But your backend need only implement those methods used by your code under test.
package awsfaker

import (
	"fmt"
	"net/http"
	"reflect"
)

// A FakeHandler is an http.Handler that can mimic an AWS service API
type FakeHandler struct {
	actions map[string]reflect.Value
}

func (h *FakeHandler) registerService(awsService interface{}) {
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

// New returns a new FakeHandler that will dispatch incoming requests to
// one or more fake service backends given as arguments.
//
// Each service backend should represent an AWS service, e.g. EC2 or CloudFormation.
// A backend should implement some or all of the actions of the service.
// The methods on the backend should have signatures like
//			func (b *MyBackend) SomeAction(input *service.SomeActionInput) (*service.SomeActionOutput, error)
// where the input and output types are those in github.com/aws/aws-sdk-go
func New(serviceBackends ...interface{}) *FakeHandler {
	handler := &FakeHandler{make(map[string]reflect.Value)}
	for _, backend := range serviceBackends {
		handler.registerService(backend)
	}
	return handler
}

// An ErrorResponse represents an error from a backend method
//
// If a backend method returns an instance of ErrorResponse, then ServeHTTP
// will respond with the given HTTPStatusCode and marshal the AWSErrorCode and
// AWSErrorMessage fields appropriately in the HTTP response body.
type ErrorResponse struct {
	AWSErrorCode    string
	AWSErrorMessage string
	HTTPStatusCode  int
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%T: %+v", e, *e)
}

func (f *FakeHandler) findMethod(actionName string) (reflect.Value, error) {
	method, ok := f.actions[actionName]
	if !ok {
		return reflect.Value{}, fmt.Errorf("action %s not found, check that you've fully implemented your fake backend", actionName)
	}
	return method, nil
}

// ServeHTTP dispatches a request to a backend method and writes the response
func (f *FakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		errorResponse := errVal.(*ErrorResponse)
		err := specializeErrorResponse(method, errorResponse)
		writeError(w, errorResponse.HTTPStatusCode, err)
		return
	}

	outVal := results[0].Interface()
	if err != nil {
		panic(err)
	}
	writeResponse(w, 200, methodName, outVal)
}
