// Package awsfaker supports the creation of test doubles for AWS APIs.
//
// To use, build a "backend" for each AWS service that you use.
// The backend should implement the subset of that API that you need.
// Each API call should be implemented as a backend method like this
//	func (b *MyBackend) SomeAction(input *service.SomeActionInput) (*service.SomeActionOutput, error)
// Then initialize an HTTP server for each backend
//	myBackend := &MyBackend{ ... }
//	fakeServer := httptest.NewServer(awsfaker.New(myBackend))
// Finally, initialize your code under test, overriding the default AWS endpoint
// to instead use your fake:
//	app := myapp.App{ AWSEndpointOverride: fakeServer.URL }
//	app.Run()
//
// The method signatures of the backend match those of the service
// interfaces of the package github.com/aws/aws-sdk-go
// For example, a complete implementation of AWS CloudFormation would match the
// CloudFormationAPI interface here: https://godoc.org/github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface
// But your backend need only implement those methods used by your code under test.
package awsfaker

import (
	"fmt"
	"net/http"

	"github.com/rosenhouse/awsfaker/protocols/query"
)

// New returns a new FakeHandler that will dispatch incoming requests to
// the given service backend.
//
// A service backend should represent one AWS service, e.g. EC2
// A backend should implement some or all of the actions of the service.
// The methods on the backend should have signatures like
//	func (b *MyBackend) SomeAction(input *service.SomeActionInput) (*service.SomeActionOutput, error)
// where the input and output types are those in github.com/aws/aws-sdk-go
// When returning an error from a backend method, use the ErrorResponse type.
func New(serviceBackend interface{}) http.Handler {
	return query.New(serviceBackend)
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
