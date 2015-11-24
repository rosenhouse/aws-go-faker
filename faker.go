package awsfaker

import (
	"fmt"
	"net/http"
	"reflect"
)

type Backend struct {
	CloudFormation interface{}
	EC2            interface{}
}

type FakeHandler struct {
	backend *Backend
}

func New(backend *Backend) *FakeHandler {
	return &FakeHandler{backend}
}

type ErrorResponse struct {
	AWSErrorCode    string
	AWSErrorMessage string
	HTTPStatusCode  int
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%T: %+v", e, *e)
}

func (f *FakeHandler) findMethod(actionName string) (reflect.Value, error) {
	for _, iface := range []interface{}{f.backend.CloudFormation, f.backend.EC2} {
		ifaceValue := reflect.ValueOf(iface)
		if !ifaceValue.IsValid() {
			continue
		}
		method := ifaceValue.MethodByName(actionName)
		if method.Kind() == reflect.Func {
			return method, nil
		}
	}
	return reflect.Value{}, fmt.Errorf("action %s not found, check that you've fully implemented your fake backend", actionName)
}

func (f *FakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	queryValues, err := parseQueryRequest(r)
	if err != nil {
		panic(err)
	}
	methodName := queryValues.Get("Action")
	method, err := f.findMethod(methodName)
	if err != nil {
		panic(err)
	}

	input, err := constructInput(method, queryValues)
	if err != nil {
		panic(err)
	}

	results := method.Call([]reflect.Value{reflect.ValueOf(input)})
	errVal := results[1].Interface()
	if errVal != nil {
		errorResponse := errVal.(*ErrorResponse)
		err := specializeErrorResponse(method, errorResponse)
		writeError(w, errorResponse.HTTPStatusCode, err)
		return
	}

	outVal := results[0].Interface()
	if err != nil {
		panic(err)
	}
	writeResponse(w, 200, methodName, outVal)
}
