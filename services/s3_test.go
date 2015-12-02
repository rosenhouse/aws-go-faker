package services_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/rosenhouse/awsfaker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type FakeS3Backend struct {
	PutObjectCall struct {
		Receives      *s3.PutObjectInput
		ReturnsResult *s3.PutObjectOutput
		ReturnsError  error
	}
}

func (f *FakeS3Backend) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	f.PutObjectCall.Receives = input
	return f.PutObjectCall.ReturnsResult, f.PutObjectCall.ReturnsError
}

var _ = Describe("Mocking out the S3 service", func() {
	var (
		fakeBackend *FakeS3Backend
		fakeServer  *httptest.Server
		client      *s3.S3
	)

	BeforeEach(func() {
		fakeBackend = &FakeS3Backend{}
		fakeServer = httptest.NewServer(awsfaker.New(fakeBackend))
		client = s3.New(
			newSession(fakeServer.URL),
			&aws.Config{S3ForcePathStyle: aws.Bool(true)},
		)
	})

	AfterEach(func() {
		if fakeServer != nil {
			fakeServer.Close()
		}
	})

	It("should call the backend method", func() {
		client.PutObject(
			&s3.PutObjectInput{
				Bucket: aws.String("some-bucket"),
				Key:    aws.String("some/object/path"),
				Body:   strings.NewReader("some object data to upload"),
				Metadata: map[string]*string{
					"key1": aws.String("a-value"),
					"key2": aws.String("b-value"),
				},
			})

		Expect(fakeBackend.PutObjectCall.Receives).NotTo(BeNil())
		Expect(fakeBackend.PutObjectCall.Receives.Bucket).To(Equal(aws.String("some-bucket")))
		Expect(fakeBackend.PutObjectCall.Receives.Key).To(Equal(aws.String("some/object/path")))
		Expect(fakeBackend.PutObjectCall.Receives.Body).To(Equal(strings.NewReader("some object data to upload")))
		Expect(fakeBackend.PutObjectCall.Receives.Metadata).To(Equal(map[string]*string{
			"key1": aws.String("a-value"),
			"key2": aws.String("b-value"),
		}))
	})

	Context("when the backend succeeds", func() {
		It("should return the data in a format parsable by the client library", func() {
			fakeBackend.PutObjectCall.ReturnsResult = &s3.PutObjectOutput{
				ETag: aws.String("some-etag"),
			}

			output, err := client.PutObject(
				&s3.PutObjectInput{
					Bucket: aws.String("some-bucket"),
					Key:    aws.String("some/object/path"),
					Body:   strings.NewReader("some object data to upload"),
				})

			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(Equal(&s3.PutObjectOutput{
				ETag: aws.String("some-etag"),
			}))
		})
	})

	Context("when the backend returns an error", func() {
		It("should return the error in a format that is parsable by the client library", func() {
			fakeBackend.PutObjectCall.ReturnsError = &awsfaker.ErrorResponse{
				AWSErrorCode:    "ValidationError",
				AWSErrorMessage: "some error message",
				HTTPStatusCode:  http.StatusBadRequest,
			}

			_, err := client.PutObject(
				&s3.PutObjectInput{
					Bucket: aws.String("some-bucket"),
					Key:    aws.String("some/object/path"),
					Body:   strings.NewReader("some object data to upload"),
				})

			Expect(err).To(HaveOccurred())
			awsErr := err.(awserr.RequestFailure)
			Expect(awsErr.StatusCode()).To(Equal(400))
			Expect(awsErr.Code()).To(Equal("ValidationError"))
			Expect(awsErr.Message()).To(Equal("some error message"))
		})
	})
})
