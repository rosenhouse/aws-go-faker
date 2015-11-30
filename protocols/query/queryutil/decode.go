// Package queryutil supports encoding and decoding between Go structs and the
// AWS query protocol.
//
// The query protocol is used by many AWS APIs, including CloudFormation,
// IAM, and RDS.  This package also supports the variant used by EC2.
//
// The encoder is copied from github.com/aws/aws-sdk-go/private/protocol/query/queryutil/
// and the decoder is based on that as well.
package queryutil

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Decode decodes from url.Values into an output object. The isEC2 flag
// indicates if this is the EC2 Query sub-protocol.
func Decode(encoded url.Values, output interface{}, isEC2 bool) error {
	q := queryDecoder{isEC2: isEC2, encoded: encoded}
	return q.decodeValue(reflect.ValueOf(output), "", "")
}

type queryDecoder struct {
	isEC2   bool
	encoded url.Values
}

func (q *queryDecoder) containsPrefix(prefix string) bool {
	for k, _ := range q.encoded {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

func getFieldType(tag reflect.StructTag, kind reflect.Kind) string {
	t := tag.Get("type")
	if t != "" {
		return t
	}
	switch kind {
	case reflect.Struct:
		return "structure"
	case reflect.Slice:
		return "list"
	case reflect.Map:
		return "map"
	}
	return ""
}

func (q *queryDecoder) decodeValue(output reflect.Value, prefix string, tag reflect.StructTag) error {
	if !q.containsPrefix(prefix) {
		return nil
	}

	if output.Kind() == reflect.Ptr {
		if output.IsNil() {
			output.Set(reflect.New(output.Type().Elem()))
		}
	}
	output = elemOf(output)

	if !output.CanSet() {
		panic("can't set " + prefix)
	}

	switch getFieldType(tag, output.Kind()) {
	case "structure":
		return q.decodeStruct(output, prefix)
	case "list":
		return q.decodeList(output, prefix, tag)
	case "map":
		return q.decodeMap(output, prefix, tag)
	default:
		return q.decodeScalar(output, prefix)
	}
}

func (q *queryDecoder) getStructFieldName(prefix string, field reflect.StructField) string {
	var name string

	if q.isEC2 {
		name = field.Tag.Get("queryName")
	}
	if name == "" {
		if field.Tag.Get("flattened") != "" && field.Tag.Get("locationNameList") != "" {
			name = field.Tag.Get("locationNameList")
		} else if locName := field.Tag.Get("locationName"); locName != "" {
			name = locName
		}
		if name != "" && q.isEC2 {
			name = strings.ToUpper(name[0:1]) + name[1:]
		}
	}
	if name == "" {
		name = field.Name
	}

	if prefix != "" {
		name = prefix + "." + name
	}

	return name
}

func (q *queryDecoder) decodeStruct(output reflect.Value, prefix string) error {
	if !output.IsValid() {
		return nil
	}

	t := output.Type()
	for i := 0; i < output.NumField(); i++ {
		if c := t.Field(i).Name[0:1]; strings.ToLower(c) == c {
			continue // ignore unexported fields
		}

		elemValue := output.Field(i)
		field := t.Field(i)
		name := q.getStructFieldName(prefix, field)

		if err := q.decodeValue(elemValue, name, field.Tag); err != nil {
			return err
		}
	}
	return nil
}

func (q *queryDecoder) getValue(keyName string, valType reflect.Type) (reflect.Value, error) {
	toSet := reflect.New(valType).Elem()
	if err := q.decodeValue(toSet, keyName, ""); err != nil {
		return reflect.Value{}, err
	}
	return toSet, nil
}

func (q *queryDecoder) decodeList(output reflect.Value, prefix string, tag reflect.StructTag) error {
	// check for unflattened list member
	if !q.isEC2 && tag.Get("flattened") == "" {
		prefix += ".member"
	}
	namer := newElementNamer(prefix)

	for i := 0; ; i++ {
		elementPrefix := namer.Name(i)
		if !q.containsPrefix(elementPrefix) {
			break
		}

		val, err := q.getValue(elementPrefix, output.Type().Elem())
		if err != nil {
			return err
		}
		output.Set(reflect.Append(output, val))
	}

	return nil
}

func (q *queryDecoder) decodeMap(output reflect.Value, prefix string, tag reflect.StructTag) error {
	// check for unflattened list member
	if !q.isEC2 && tag.Get("flattened") == "" {
		prefix += ".entry"
	}
	namer := newMapNamer(prefix, tag)

	mapType := output.Type()
	output.Set(reflect.MakeMap(mapType))

	keyType := mapType.Key()
	valueType := mapType.Elem()

	for i := 0; ; i++ {
		keyName := namer.KeyName(i)
		if !q.containsPrefix(keyName) {
			break
		}

		key, err := q.getValue(keyName, keyType)
		if err != nil {
			return err
		}

		val, err := q.getValue(namer.ValueName(i), valueType)
		if err != nil {
			return err
		}

		output.SetMapIndex(key, val)
	}

	return nil
}

func (q *queryDecoder) decodeScalar(output reflect.Value, name string) error {
	encodedValue := q.encoded.Get(name)
	if encodedValue == "" {
		return nil
	}
	switch output.Interface().(type) {
	case string:
		output.SetString(encodedValue)
	case []byte:
		decoded, err := base64.StdEncoding.DecodeString(encodedValue)
		if err != nil {
			return &decodeError{Field: name, Value: encodedValue, Inner: err}
		}
		output.SetBytes(decoded)
	case bool:
		value, err := strconv.ParseBool(encodedValue)
		if err != nil {
			return &decodeError{Field: name, Value: encodedValue, Inner: err}
		}
		output.SetBool(value)
	case int64, int, int32:
		value, err := strconv.ParseInt(encodedValue, 10, 64)
		if err != nil {
			return &decodeError{Field: name, Value: encodedValue, Inner: err}
		}
		output.SetInt(value)
	case float64, float32:
		value, err := strconv.ParseFloat(encodedValue, 64)
		if err != nil {
			return &decodeError{Field: name, Value: encodedValue, Inner: err}
		}
		output.SetFloat(value)
	case time.Time:
		const ISO8601UTC = "2006-01-02T15:04:05Z"
		value, err := time.Parse(ISO8601UTC, encodedValue)
		if err != nil {
			return &decodeError{Field: name, Value: encodedValue, Inner: err}
		}
		output.Set(reflect.ValueOf(value.UTC()))
	default:
		return fmt.Errorf("unsupported value for param %s: %v (%s)", name, output.Interface(), output.Type().Name())
	}
	return nil
}
