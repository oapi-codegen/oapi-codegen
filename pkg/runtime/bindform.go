package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/codegen"
	"github.com/deepmap/oapi-codegen/pkg/types"
)

const tagName = "json"
const jsonContentType = "application/json"

func BindForm(ptr interface{}, form map[string][]string, files map[string][]*multipart.FileHeader, encodings map[string]codegen.RequestBodyEncoding) error {
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
	//TODO: support additional properties
	return false, nil
}
