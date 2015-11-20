package awsfaker_test

import (
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"

	"testing"
)

func TestAwsFaker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AwsFaker Suite")
}

var _ = BeforeSuite(func() {
	rand.Seed(config.GinkgoConfig.RandomSeed)
})
