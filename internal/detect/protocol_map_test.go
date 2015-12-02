package detect_test

import (
	"go/build"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/awsfaker/internal/detect"
	"github.com/rosenhouse/awsfaker/internal/usage"

	"golang.org/x/tools/refactor/importgraph"
)

var _ = Describe("Mapping all the services to their protocols", func() {
	It("returns the complete usage map", func() {
		_, reverseImports, _ := importgraph.Build(&build.Default)
		expectedUsage := usage.GetShortUsage(reverseImports)

		numServicesCounted := 0
		for protocol, expectedServices := range expectedUsage {
			for _, service := range expectedServices {
				Expect(detect.ProtocolForService[service]).To(Equal(protocol))
				numServicesCounted++
			}
		}

		Expect(detect.ProtocolForService).To(HaveLen(numServicesCounted))
	})
})
