package queryutil_test

import (
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/rosenhouse/awsfaker/queryutil"
)

type dataObject struct {
	SomeString  string
	Bytes       []byte `type:"blob"`
	StructArray []*Inner
	BasicMap    map[string]*string
	StructMap   map[string]*Inner
}

type Inner struct {
	Bool        *bool
	Int64       *int64
	Int         *int
	Float64     *float64
	Time        *time.Time `type:"timestamp"`
	ScalarArray []*string
}

var _ = Describe("Encode / decode cycle", func() {
	It("should encode and decode without data loss", func() {
		originalData := dataObject{
			SomeString: "some string",
			Bytes:      []byte("some bytes"),
			StructArray: []*Inner{
				&Inner{
					Bool:    aws.Bool(true),
					Int64:   aws.Int64(5123456789),
					Int:     aws.Int(-123),
					Float64: aws.Float64(1.234),
					Time:    aws.Time(time.Now().UTC().Round(time.Second)),
					ScalarArray: []*string{
						aws.String("one"),
						aws.String("two"),
					},
				},
			},
			BasicMap: map[string]*string{
				"A": aws.String("B"),
				"C": aws.String("D"),
				"E": aws.String("F"),
			},
			StructMap: map[string]*Inner{
				"One": &Inner{
					Bool:    aws.Bool(true),
					Int64:   aws.Int64(5123456789),
					Int:     aws.Int(-123),
					Float64: aws.Float64(1.234),
					Time:    aws.Time(time.Now().UTC().Round(time.Second)),
					ScalarArray: []*string{
						aws.String("one"),
						aws.String("two"),
					},
				},
			},
		}

		encoded := make(url.Values)

		err := queryutil.Encode(encoded, originalData, false)
		Expect(err).NotTo(HaveOccurred())

		decoded := dataObject{}
		err = queryutil.Decode(encoded, &decoded, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(decoded).To(Equal(originalData))
	})

	It("it should use nil values for missing data", func() {
		originalData := dataObject{
			SomeString: "some string",
			Bytes:      []byte("some bytes"),
			StructArray: []*Inner{
				&Inner{
					Bool: aws.Bool(true),
				},
			},
			BasicMap: map[string]*string{
				"A": aws.String("B"),
				"C": aws.String("D"),
				"E": aws.String("F"),
			},
			StructMap: map[string]*Inner{
				"One": &Inner{
					Int: aws.Int(-123),
					ScalarArray: []*string{
						aws.String("one"),
						aws.String("two"),
					},
				},
			},
		}

		encoded := make(url.Values)

		err := queryutil.Encode(encoded, originalData, false)
		Expect(err).NotTo(HaveOccurred())

		decoded := dataObject{}
		err = queryutil.Decode(encoded, &decoded, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(decoded).To(Equal(originalData))
	})

})
