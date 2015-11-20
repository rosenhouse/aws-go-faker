package awsfaker_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"

	"github.com/rosenhouse/aws-go-faker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type FakeCloudFormationBackend struct {
	DescribeStacksCall struct {
		Receives      *cloudformation.DescribeStacksInput
		ReturnsResult *cloudformation.DescribeStacksOutput
		ReturnsError  error
	}
}

func (f *FakeCloudFormationBackend) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	f.DescribeStacksCall.Receives = input
	return f.DescribeStacksCall.ReturnsResult, f.DescribeStacksCall.ReturnsError
}

var _ = Describe("Mocking out an AWS service over the network", func() {
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
		faker       *awsfaker.Faker
	)

	BeforeEach(func() {
		fakeBackend = &FakeCloudFormationBackend{}
		faker = &awsfaker.Faker{}
	})

	It("should call the backend method", func() {
		fakeServer := httptest.NewServer(faker.Handler(fakeBackend))
		defer fakeServer.Close()

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

			fakeServer := httptest.NewServer(faker.Handler(fakeBackend))
			defer fakeServer.Close()

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
				Code:       "ValidationError",
				Message:    "some error message",
				StatusCode: http.StatusBadRequest,
			}

			fakeServer := httptest.NewServer(faker.Handler(fakeBackend))
			defer fakeServer.Close()

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

var _ = Describe("inner functions", func() {
	Describe("ConstructInput", func() {
		It("should return a value of the correct type", func() {
			fakeBackend := &FakeCloudFormationBackend{}

			method := reflect.ValueOf(fakeBackend).MethodByName("DescribeStacks")

			queryValues, _ := url.ParseQuery("Action=DescribeStacks&StackName=some-stack-name&Version=2010-05-15")
			expectedInput, err := awsfaker.ConstructInput(method, queryValues)
			Expect(err).NotTo(HaveOccurred())

			Expect(expectedInput).To(Equal(&cloudformation.DescribeStacksInput{
				StackName: aws.String("some-stack-name"),
			}))

		})
	})

})
