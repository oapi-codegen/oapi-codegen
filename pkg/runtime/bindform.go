package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/types"
)

const tagName = "json"
const jsonContentType = "application/json"

type RequestBodyEncoding struct {
	ContentType string
	Style       string
	Explode     *bool
}

func BindForm(ptr interface{}, form map[string][]string, files map[string][]*multipart.FileHeader, encodings map[string]RequestBodyEncoding) error {
	ptrVal := reflect.Indirect(reflect.ValueOf(ptr))
	if ptrVal.Kind() != reflect.Struct {
		return errors.New("form data body should be a struct")
	}
	tValue := ptrVal.Type()

	for i := 0; i < tValue.NumField(); i++ {
		field := tValue.Field(i)
		tag := field.Tag.Get(tagName)
		if !field.IsExported() || tag == "-" {
			continue
		}
		tag = strings.Split(tag, ",")[0] // extract the name of the tag
		if encoding, ok := encodings[tag]; ok {
			// custom encoding
			values := form[tag]
			if len(values) == 0 {
				continue
			}
			value := values[0]
			if encoding.ContentType != "" {
				if strings.HasPrefix(encoding.ContentType, jsonContentType) {
					if err := json.Unmarshal([]byte(value), ptr); err != nil {
						return err
					}
				}
				return errors.New("unsupported encoding, only application/json is supported")
			} else {
				var explode bool
				if encoding.Explode != nil {
					explode = *encoding.Explode
				}
				if err := BindStyledParameterWithLocation(encoding.Style, explode, tag, ParamLocationUndefined, value, ptrVal.Field(i).Addr().Interface()); err != nil {
					return err
				}
			}
		} else {
			// regular form data
			if _, err := bindFormImpl(ptrVal.Field(i), form, files, tag); err != nil {
				return err
			}
		}
	}

	return nil
}

func MarshalForm(ptr interface{}, encodings map[string]RequestBodyEncoding) (url.Values, error) {
	ptrVal := reflect.Indirect(reflect.ValueOf(ptr))
	if ptrVal.Kind() != reflect.Struct {
		return nil, errors.New("form data body should be a struct")
	}
	tValue := ptrVal.Type()
	result := make(url.Values)
	for i := 0; i < tValue.NumField(); i++ {
		field := tValue.Field(i)
		tag := field.Tag.Get(tagName)
		if !field.IsExported() || tag == "-" {
			continue
		}
		omitEmpty := strings.HasSuffix(tag, ",omitempty")
		if omitEmpty && ptrVal.Field(i).IsZero() {
			continue
		}
		tag = strings.Split(tag, ",")[0] // extract the name of the tag
		if encoding, ok := encodings[tag]; ok && encoding.ContentType != "" {
			if strings.HasPrefix(encoding.ContentType, jsonContentType) {
				if data, err := json.Marshal(ptrVal.Field(i)); err != nil {
					return nil, err
				} else {
					result[tag] = append(result[tag], string(data))
				}
			}
			return nil, errors.New("unsupported encoding, only application/json is supported")
		} else {
			marshalFormImpl(ptrVal.Field(i), result, tag)
		}
	}
	return result, nil
}

func bindFormImpl(v reflect.Value, form map[string][]string, files map[string][]*multipart.FileHeader, name string) (bool, error) {
	var hasData bool
	switch v.Kind() {
	case reflect.Interface:
		return bindFormImpl(v.Elem(), form, files, name)
	case reflect.Ptr:
		ptrData := v.Elem()
		if !ptrData.IsValid() {
			ptrData = reflect.New(v.Type().Elem())
		}
		ptrHasData, err := bindFormImpl(ptrData, form, files, name)
		if err == nil && ptrHasData && !v.Elem().IsValid() {
			v.Set(ptrData)
		}
		return ptrHasData, err
	case reflect.Slice:
		if files := append(files[name], files[name+"[]"]...); len(files) != 0 {
			if _, ok := v.Interface().([]types.File); ok {
				result := make([]types.File, len(files), len(files))
				for i, file := range files {
					result[i].InitFromMultipart(file)
				}
				v.Set(reflect.ValueOf(result))
				hasData = true
			}
		}
		indexedElementsCount := indexedElementsCount(form, files, name)
		items := append(form[name], form[name+"[]"]...)
		if indexedElementsCount+len(items) != 0 {
			result := reflect.MakeSlice(v.Type(), indexedElementsCount+len(items), indexedElementsCount+len(items))
			for i := 0; i < indexedElementsCount; i++ {
				if _, err := bindFormImpl(result.Index(i), form, files, fmt.Sprintf("%s[%v]", name, i)); err != nil {
					return false, err
				}
			}
			for i, item := range items {
				if err := BindStringToObject(item, result.Index(indexedElementsCount+i).Addr().Interface()); err != nil {
					return false, err
				}
			}
			v.Set(result)
			hasData = true
		}
	case reflect.Struct:
		if files := files[name]; len(files) != 0 {
			if file, ok := v.Interface().(types.File); ok {
				file.InitFromMultipart(files[0])
				v.Set(reflect.ValueOf(file))
				return true, nil
			}
		}
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			tag := field.Tag.Get(tagName)
			if field.Name == "AdditionalProperties" && tag == "-" {
				additionalPropertiesHasData, err := bindAdditionalProperties(v.Field(i), form, files, name)
				if err != nil {
					return false, err
				}
				hasData = hasData || additionalPropertiesHasData
			}
			if !field.IsExported() || tag == "-" {
				continue
			}
			tag = strings.Split(tag, ",")[0] // extract the name of the tag
			fieldHasData, err := bindFormImpl(v.Field(i), form, files, fmt.Sprintf("%s[%s]", name, tag))
			if err != nil {
				return false, err
			}
			hasData = hasData || fieldHasData
		}
		return hasData, nil
	default:
		value := form[name]
		if len(value) != 0 {
			return true, BindStringToObject(value[0], v.Addr().Interface())
		}
	}
	return hasData, nil
}

func indexedElementsCount(form map[string][]string, files map[string][]*multipart.FileHeader, name string) int {
	name += "["
	maxIndex := -1
	for k := range form {
		if strings.HasPrefix(k, name) {
			str := strings.TrimPrefix(k, name)
			str = str[:strings.Index(str, "]")]
			if idx, err := strconv.Atoi(str); err == nil {
				if idx > maxIndex {
					maxIndex = idx
				}
			}
		}
	}
	for k := range files {
		if strings.HasPrefix(k, name) {
			str := strings.TrimPrefix(k, name)
			str = str[:strings.Index(str, "]")]
			if idx, err := strconv.Atoi(str); err == nil {
				if idx > maxIndex {
					maxIndex = idx
				}
			}
		}
	}
	return maxIndex + 1
}

func bindAdditionalProperties(additionalProperties reflect.Value, form map[string][]string, files map[string][]*multipart.FileHeader, name string) (bool, error) {
	// TODO: support additional properties
	return false, nil
}

func marshalFormImpl(v reflect.Value, result url.Values, name string) {
	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		marshalFormImpl(v.Elem(), result, name)
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			marshalFormImpl(elem, result, fmt.Sprintf("%s[%v]", name, i))
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			tag := field.Tag.Get(tagName)
			if field.Name == "AdditionalProperties" && tag == "-" {
				iter := v.MapRange()
				for iter.Next() {
					marshalFormImpl(iter.Value(), result, fmt.Sprintf("%s[%s]", name, iter.Key().String()))
				}
				continue
			}
			if !field.IsExported() || tag == "-" {
				continue
			}
			tag = strings.Split(tag, ",")[0] // extract the name of the tag
			marshalFormImpl(v.Field(i), result, fmt.Sprintf("%s[%s]", name, tag))
		}
	default:
		result[name] = append(result[name], fmt.Sprint(v.Interface()))
	}
}
