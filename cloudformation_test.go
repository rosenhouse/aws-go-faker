package awsfaker_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/rosenhouse/awsfaker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type FakeCloudFormationBackend struct {
	DescribeStacksCall struct {
		Receives      *cloudformation.DescribeStacksInput
		ReturnsResult *cloudformation.DescribeStacksOutput
		ReturnsError  error
	}
	UpdateStackCall struct {
		Receives      *cloudformation.UpdateStackInput
		ReturnsResult *cloudformation.UpdateStackOutput
		ReturnsError  error
	}
}

func (f *FakeCloudFormationBackend) UpdateStack(input *cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	f.UpdateStackCall.Receives = input
	return f.UpdateStackCall.ReturnsResult, f.UpdateStackCall.ReturnsError
}

func (f *FakeCloudFormationBackend) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	f.DescribeStacksCall.Receives = input
	return f.DescribeStacksCall.ReturnsResult, f.DescribeStacksCall.ReturnsError
}

var _ = Describe("Mocking out the CloudFormation service", func() {
	newClient := func(endpointBaseURL string) *cloudformation.CloudFormation {
		credentials := credentials.NewStaticCredentials("some-access-key", "some-secret-key", "")
		sdkConfig := &aws.Config{
			Credentials: credentials,
			Region:      aws.String("some-region"),
			Endpoint:    aws.String(endpointBaseURL),
		}
		return cloudformation.New(session.New(sdkConfig))
	}

	var (
		fakeBackend *FakeCloudFormationBackend
		fakeHandler *awsfaker.FakeHandler
		fakeServer  *httptest.Server
	)

	BeforeEach(func() {
		fakeBackend = &FakeCloudFormationBackend{}
		fakeHandler = awsfaker.New(awsfaker.Backend{CloudFormation: fakeBackend})
		fakeServer = httptest.NewServer(fakeHandler)
	})
	AfterEach(func() {
		if fakeServer != nil {
			fakeServer.Close()
		}
	})

	It("should properly parse nested structs in the input", func() {
		newClient(fakeServer.URL).UpdateStack(
			&cloudformation.UpdateStackInput{
				StackName: aws.String("some-stack-name"),
				Parameters: []*cloudformation.Parameter{
					&cloudformation.Parameter{
						ParameterKey:   aws.String("some-key-0"),
						ParameterValue: aws.String("some-value-0"),
					},
					&cloudformation.Parameter{
						ParameterKey:   aws.String("some-key-1"),
						ParameterValue: aws.String("some-value-1"),
					},
				},
			})

		Expect(fakeBackend.UpdateStackCall.Receives).NotTo(BeNil())
		Expect(fakeBackend.UpdateStackCall.Receives.StackName).To(Equal(aws.String("some-stack-name")))
		Expect(fakeBackend.UpdateStackCall.Receives.Parameters).To(Equal(
			[]*cloudformation.Parameter{
				&cloudformation.Parameter{
					ParameterKey:   aws.String("some-key-0"),
					ParameterValue: aws.String("some-value-0"),
				},
				&cloudformation.Parameter{
					ParameterKey:   aws.String("some-key-1"),
					ParameterValue: aws.String("some-value-1"),
				},
			},
		))
	})

	It("should call the backend method", func() {
		newClient(fakeServer.URL).DescribeStacks(
			&cloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			})

		Expect(fakeBackend.DescribeStacksCall.Receives).NotTo(BeNil())
		Expect(fakeBackend.DescribeStacksCall.Receives.StackName).To(Equal(aws.String("some-stack-name")))
	})

	Context("when the backend succeeds", func() {
		It("should return the data in a format parsable by the client library", func() {
			fakeBackend.DescribeStacksCall.ReturnsResult = &cloudformation.DescribeStacksOutput{
				Stacks: []*cloudformation.Stack{
					&cloudformation.Stack{
						StackName: aws.String("first stack"),
						Outputs: []*cloudformation.Output{
							&cloudformation.Output{
								OutputKey:   aws.String("some-key"),
								OutputValue: aws.String("some-value"),
							},
						},
					},
					&cloudformation.Stack{
						StackName: aws.String("second stack"),
					},
				},
			}

			output, err := newClient(fakeServer.URL).DescribeStacks(
				&cloudformation.DescribeStacksInput{
					StackName: aws.String("some-stack-name"),
				})

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(&cloudformation.DescribeStacksOutput{
				Stacks: []*cloudformation.Stack{
					&cloudformation.Stack{
						StackName: aws.String("first stack"),
						Outputs: []*cloudformation.Output{
							&cloudformation.Output{
								OutputKey:   aws.String("some-key"),
								OutputValue: aws.String("some-value"),
							},
						},
					},
					&cloudformation.Stack{
						StackName: aws.String("second stack"),
					},
				},
			}))
		})
	})

	Context("when the backend returns an error", func() {
		It("should return the error in a format that is parsable by the client library", func() {
			fakeBackend.DescribeStacksCall.ReturnsError = &awsfaker.ErrorResponse{
				AWSErrorCode:    "ValidationError",
				AWSErrorMessage: "some error message",
				HTTPStatusCode:  http.StatusBadRequest,
			}

			_, err := newClient(fakeServer.URL).DescribeStacks(
				&cloudformation.DescribeStacksInput{
					StackName: aws.String("some-stack-name"),
				})

			Expect(err).To(HaveOccurred())
			awsErr := err.(awserr.RequestFailure)
			Expect(awsErr.StatusCode()).To(Equal(400))
			Expect(awsErr.Code()).To(Equal("ValidationError"))
			Expect(awsErr.Message()).To(Equal("some error message"))
		})
	})
})
