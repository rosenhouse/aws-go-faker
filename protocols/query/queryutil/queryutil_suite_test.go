package queryutil_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestQueryutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "QueryUtil Suite")
}
