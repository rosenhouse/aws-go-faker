package awsfaker_test

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/awsfaker"
	"github.com/rosenhouse/awsfaker/protocols/query"
	"github.com/rosenhouse/awsfaker/protocols/restxml"
)

type EC2 struct{}

func (e *EC2) CreateKeyPair(*ec2.CreateKeyPairInput) (*ec2.CreateKeyPairOutput, error) {
	return nil, nil
}

type CloudFormation struct{}

func (c *CloudFormation) DescribeStacks(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	return nil, nil
}

type S3 struct{}

func (s *S3) GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return nil, nil
}

var _ = Describe("New", func() {
	It("should detect the protocol correctly", func() {
		Expect(awsfaker.New(new(EC2))).To(BeAssignableToTypeOf(&query.Handler{}))

		Expect(awsfaker.New(new(CloudFormation))).To(BeAssignableToTypeOf(&query.Handler{}))

		Expect(awsfaker.New(new(S3))).To(BeAssignableToTypeOf(&restxml.Handler{}))
	})
})
