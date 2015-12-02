package detect_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/awsfaker/internal/detect"
)

type SomeServiceBackend struct{}

func (n *SomeServiceBackend) SomeServiceCall(*strings.Reader) (*strings.Reader, error) {
	return nil, nil
}

type SomeInterface interface {
	SomeServiceCall(*strings.Reader) (*strings.Reader, error)
}

type TypeWithMethodWithTooManyArgs struct{}

func (i *TypeWithMethodWithTooManyArgs) SomeServiceCall(one, two *strings.Reader) (*strings.Reader, error) {
	return nil, nil
}

type TypeWithMethodWithArgWithoutPkgPath struct{}

func (i *TypeWithMethodWithArgWithoutPkgPath) SomeServiceCall(one **strings.Reader) (*strings.Reader, error) {
	return nil, nil
}

type TypeWithMethodWithArgWithoutPkgPath2 struct{}

func (i *TypeWithMethodWithArgWithoutPkgPath2) SomeServiceCall(one *string) (*strings.Reader, error) {
	return nil, nil
}

type TypeWithMethodWithArgNotAPointer struct{}

func (n *TypeWithMethodWithArgNotAPointer) SomeServiceCall(strings.Reader) (*strings.Reader, error) {
	return nil, nil
}

var _ = Describe("Detecting the service name", func() {
	It("should return the short name of the package containing the type of the first argument", func() {
		Expect(detect.GetServiceName(new(SomeServiceBackend))).To(Equal("strings"))
	})

	Context("when given bad inputs", func() {

		Context("when given a nil interface value", func() {
			It("should return an error", func() {
				var somethingNil SomeInterface
				_, err := detect.GetServiceName(somethingNil)
				Expect(err).To(MatchError("expected non-nil service backend"))
			})
		})

		Context("when given a non-pointer type", func() {
			It("should return an error", func() {
				_, err := detect.GetServiceName(SomeServiceBackend{})
				Expect(err).To(MatchError("expected pointer type"))
			})
		})

		Context("when the provided type has no methods", func() {
			It("should return an error", func() {
				var something struct{}
				_, err := detect.GetServiceName(&something)
				Expect(err).To(MatchError("no methods found"))
			})
		})

		Context("when the found method has an unexpected signature", func() {
			It("should return an error", func() {
				_, err := detect.GetServiceName(&TypeWithMethodWithTooManyArgs{})
				Expect(err).To(MatchError("expected method with receiver plus single argument, instead got: func(*detect_test.TypeWithMethodWithTooManyArgs, *strings.Reader, *strings.Reader) (*strings.Reader, error)"))
			})
		})

		Context("when the argument to the found method is not a pointer", func() {
			It("should return an error", func() {
				_, err := detect.GetServiceName(&TypeWithMethodWithArgNotAPointer{})
				Expect(err).To(MatchError("expected argument to be pointer type"))

			})
		})

		Context("when the argument to the found method doesn't yield a pkgpath", func() {
			It("should return an error", func() {
				_, err := detect.GetServiceName(&TypeWithMethodWithArgWithoutPkgPath{})
				Expect(err).To(MatchError("expected argument to be pointer to non-basic type"))

				_, err = detect.GetServiceName(&TypeWithMethodWithArgWithoutPkgPath2{})
				Expect(err).To(MatchError("expected argument to be pointer to non-basic type"))
			})
		})
	})
})
