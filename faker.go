package awsfaker

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol/xml/xmlutil"
)

type Faker struct{}

func writeResponse(w http.ResponseWriter, statusCode int, action string, data interface{}) {
	responseBuffer := &bytes.Buffer{}
	encoder := xml.NewEncoder(responseBuffer)
	resultWrapper := xml.StartElement{Name: xml.Name{Local: action + "Result"}}
	err := encoder.EncodeToken(resultWrapper)
	if err != nil {
		panic(err)
	}
	err = xmlutil.BuildXML(data, encoder)
	if err != nil {
		panic(err)
	}
	err = encoder.EncodeToken(resultWrapper.End())
	if err != nil {
		panic(err)
	}
	err = encoder.Flush()
	if err != nil {
		panic(err)
	}

	w.WriteHeader(statusCode)
	_, err = w.Write(responseBuffer.Bytes())
	if err != nil {
		panic(err)
	}
}

func writeError(w http.ResponseWriter, errorResponse *ErrorResponse) {
	responseBodyBytes, err := xml.Marshal(errorResponse)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(errorResponse.StatusCode)
	_, err = w.Write(responseBodyBytes)
	if err != nil {
		panic(err)
	}
}

func parseQueryRequest(r *http.Request) (url.Values, error) {
	requestBodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read request body: %s", err)
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(requestBodyBytes))

	values, err := url.ParseQuery(string(requestBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("unable to parse request body as query syntax: %s", err)
	}

	return values, err
}

func ConstructInput(method reflect.Value, queryValues url.Values) (interface{}, error) {
	inputValueType := method.Type().In(0).Elem()
	inputValue := reflect.New(inputValueType).Elem()
	for k, vs := range queryValues {
		field := inputValue.FieldByName(k)
		if field.IsValid() {
			field.Set(reflect.ValueOf(aws.String(vs[0])))
		}
	}
	return inputValue.Addr().Interface(), nil
}

func (f *Faker) Handler(fakeBackend interface{}) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			queryValues, err := parseQueryRequest(r)
			if err != nil {
				panic(err)
			}
			methodName := queryValues.Get("Action")
			method := reflect.ValueOf(fakeBackend).MethodByName(methodName)
			if method.Kind() != reflect.Func {
				panic("missing method: " + methodName)
			}

			input, err := ConstructInput(method, queryValues)
			if err != nil {
				panic(err)
			}

			results := method.Call([]reflect.Value{reflect.ValueOf(input)})
			errVal := results[1].Interface()
			if errVal != nil {
				errorResponse := errVal.(*ErrorResponse)
				writeError(w, errorResponse)
				return
			}

			outVal := results[0].Interface()
			if err != nil {
				panic(err)
			}
			writeResponse(w, 200, methodName, outVal)
		})
}

type ErrorResponse struct {
	XMLName    xml.Name `xml:"ErrorResponse"`
	Code       string   `xml:"Error>Code"`
	Message    string   `xml:"Error>Message"`
	RequestID  string   `xml:"RequestId"`
	StatusCode int      `xml:"-"`
}

func (e *ErrorResponse) Error() string { return fmt.Sprintf("you shouldn't see this msg") }
