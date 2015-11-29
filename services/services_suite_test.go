package services_test

import (
	"math/rand"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"

	"testing"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Services Suite")
}

var _ = BeforeSuite(func() {
	rand.Seed(config.GinkgoConfig.RandomSeed)
})

func newSession(endpointBaseURL string) *session.Session {
	credentials := credentials.NewStaticCredentials("some-access-key", "some-secret-key", "")
	sdkConfig := &aws.Config{
		Credentials: credentials,
		Region:      aws.String("some-region"),
		Endpoint:    aws.String(endpointBaseURL),
	}
	return session.New(sdkConfig)
}
