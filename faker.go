package awsfaker

import (
	"fmt"
	"net/http"
	"reflect"
)

type FakeHandler struct {
	actions map[string]reflect.Value
}

func (h *FakeHandler) registerService(awsService interface{}) {
	service := reflect.ValueOf(awsService)
	if !service.IsValid() {
		panic("invalid service interface")
	}
	if service.Kind() != reflect.Ptr {
		panic("expecting struct pointer as service interface")
	}
	if !service.Elem().IsValid() {
		panic("expectingn non-nil pointer as service interface")
	}
	serviceType := service.Type()
	n := service.NumMethod()
	if n == 0 {
		panic("no methods on service interface")
	}
	for i := 0; i < n; i++ {
		h.actions[serviceType.Method(i).Name] = service.Method(i)
	}
}

func New(serviceBackends ...interface{}) *FakeHandler {
	handler := &FakeHandler{make(map[string]reflect.Value)}
	for _, backend := range serviceBackends {
		handler.registerService(backend)
	}
	return handler
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
	method, ok := f.actions[actionName]
	if !ok {
		return reflect.Value{}, fmt.Errorf("action %s not found, check that you've fully implemented your fake backend", actionName)
	}
	return method, nil
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
