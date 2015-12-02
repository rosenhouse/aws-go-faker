package awsfaker_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAWSfaker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AWSFaker Suite")
}
