package awsfaker_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/rosenhouse/awsfaker"
)

type CloudFormationBackend struct {
	createStackCallCount int
}

func (b *CloudFormationBackend) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	stackName := aws.StringValue(input.StackName)
	fmt.Printf("[Server] CreateStack called on %q\n", stackName)

	if b.createStackCallCount == 0 {
		b.createStackCallCount++
		return &cloudformation.CreateStackOutput{
			StackId: aws.String("some-id"),
		}, nil
	} else {
		return nil, &awsfaker.ErrorResponse{
			HTTPStatusCode:  http.StatusBadRequest,
			AWSErrorCode:    "AlreadyExistsException",
			AWSErrorMessage: fmt.Sprintf("Stack [%s] already exists", stackName),
		}
	}
}

func Example() {
	// create a backend that implements the subset of the AWS API you need
	fakeBackend := &CloudFormationBackend{}

	// start a local HTTP server that dispatches requests to the backend
	fakeServer := httptest.NewServer(awsfaker.New(fakeBackend))

	// configure and use your client.  this might be a separate process,
	// with the endpoint override set via environment variable or other config
	client := cloudformation.New(session.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials("some-access-key", "some-secret-key", ""),
		Region:      aws.String("some-region"),
		Endpoint:    aws.String(fakeServer.URL), // override the default AWS endpoint
	}))

	out, err := client.CreateStack(&cloudformation.CreateStackInput{
		StackName: aws.String("some-stack"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("[Client] CreateStack returned ID: %q\n", *out.StackId)

	_, err = client.CreateStack(&cloudformation.CreateStackInput{
		StackName: aws.String("some-stack"),
	})
	fmt.Printf("[Client] CreateStack returned error:\n %s\n", err)
	// Output:
	// [Server] CreateStack called on "some-stack"
	// [Client] CreateStack returned ID: "some-id"
	// [Server] CreateStack called on "some-stack"
	// [Client] CreateStack returned error:
	//  AlreadyExistsException: Stack [some-stack] already exists
	// 	status code: 400, request id:
}
