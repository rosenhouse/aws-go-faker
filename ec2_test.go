package awsfaker_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/rosenhouse/awsfaker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type FakeEC2Backend struct {
	CreateKeyPairCall struct {
		Receives      *ec2.CreateKeyPairInput
		ReturnsResult *ec2.CreateKeyPairOutput
		ReturnsError  error
	}
}

func (f *FakeEC2Backend) CreateKeyPair(input *ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error) {
	f.CreateKeyPairCall.Receives = input
	return f.CreateKeyPairCall.ReturnsResult, f.CreateKeyPairCall.ReturnsError
}

var _ = Describe("Mocking out the EC2 service", func() {
	var (
		fakeBackend *FakeEC2Backend
		fakeServer  *httptest.Server
		client      *ec2.EC2
	)

	BeforeEach(func() {
		fakeBackend = &FakeEC2Backend{}
		fakeServer = httptest.NewServer(awsfaker.New(fakeBackend))
		client = ec2.New(newSession(fakeServer.URL))
	})

	AfterEach(func() {
		if fakeServer != nil {
			fakeServer.Close()
		}
	})

	It("should call the backend method", func() {
		client.CreateKeyPair(
			&ec2.CreateKeyPairInput{
				KeyName: aws.String("some-key-name"),
			})

		Expect(fakeBackend.CreateKeyPairCall.Receives).NotTo(BeNil())
		Expect(fakeBackend.CreateKeyPairCall.Receives.KeyName).To(Equal(aws.String("some-key-name")))
	})

	Context("when the backend succeeds", func() {
		It("should return the data in a format parsable by the client library", func() {
			fakeBackend.CreateKeyPairCall.ReturnsResult = &ec2.CreateKeyPairOutput{
				KeyFingerprint: aws.String("some-fingerprint"),
				KeyMaterial:    aws.String("some-pem-data"),
				KeyName:        aws.String("some-key-name"),
			}

			output, err := client.CreateKeyPair(
				&ec2.CreateKeyPairInput{
					KeyName: aws.String("some-key-name"),
				})

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(&ec2.CreateKeyPairOutput{
				KeyFingerprint: aws.String("some-fingerprint"),
				KeyMaterial:    aws.String("some-pem-data"),
				KeyName:        aws.String("some-key-name"),
			}))
		})
	})

	Context("when the backend returns an error", func() {
		It("should return the error in a format that is parsable by the client library", func() {
			fakeBackend.CreateKeyPairCall.ReturnsError = &awsfaker.ErrorResponse{
				AWSErrorCode:    "ValidationError",
				AWSErrorMessage: "some error message",
				HTTPStatusCode:  http.StatusBadRequest,
			}

			_, err := client.CreateKeyPair(
				&ec2.CreateKeyPairInput{
					KeyName: aws.String("some-key-name"),
				})

			Expect(err).To(HaveOccurred())
			awsErr := err.(awserr.RequestFailure)
			Expect(awsErr.StatusCode()).To(Equal(400))
			Expect(awsErr.Code()).To(Equal("ValidationError"))
			Expect(awsErr.Message()).To(Equal("some error message"))
		})
	})
})
