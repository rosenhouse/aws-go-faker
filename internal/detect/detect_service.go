package detect

import (
	"fmt"
	"reflect"
	"strings"
)

func getShortPkgPath(t reflect.Type) string {
	parts := strings.Split(t.PkgPath(), "/")
	return parts[len(parts)-1]
}

func GetServiceName(serviceBackend interface{}) (string, error) {
	t := reflect.TypeOf(serviceBackend)
	if t == nil {
		return "", fmt.Errorf("expected non-nil service backend")
	}
	if t.Kind() != reflect.Ptr {
		return "", fmt.Errorf("expected pointer type")
	}
	if t.NumMethod() == 0 {
		return "", fmt.Errorf("no methods found")
	}
	methodType := t.Method(0).Type
	if methodType.NumIn() != 2 {
		return "", fmt.Errorf(
			"expected method with receiver plus single argument, instead got: %+v",
			methodType)
	}
	argType := methodType.In(1)
	if argType.Kind() != reflect.Ptr {
		return "", fmt.Errorf("expected argument to be pointer type")
	}

	pkgPath := getShortPkgPath(argType.Elem())
	if pkgPath == "" {
		return "", fmt.Errorf("expected argument to be pointer to non-basic type")
	}

	return pkgPath, nil
}

func getProtocol(serviceName string) (string, error) {
	return "", nil
}
