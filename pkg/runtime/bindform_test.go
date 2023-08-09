package runtime

import (
	"bytes"
	"mime/multipart"
	"net/url"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestBindURLForm(t *testing.T) {
	type testSubStruct struct {
		Int                  int               `json:"int"`
		String               string            `json:"string"`
		AdditionalProperties map[string]string `json:"-"`
	}
	type testStruct struct {
		Int             int              `json:"int"`
		Bool            bool             `json:"bool,omitempty"`
		String          string           `json:"string"`
		IntSlice        []int            `json:"int_slice"`
		Struct          testSubStruct    `json:"struct"`
		StructSlice     []testSubStruct  `json:"struct_slice"`
		OptInt          *int             `json:"opt_int,omitempty"`
		OptBool         *bool            `json:"opt_bool,omitempty"`
		OptString       *string          `json:"opt_string,omitempty"`
		OptStruct       *testSubStruct   `json:"opt_struct,omitempty"`
		OptStructSlice  *[]testSubStruct `json:"opt_struct_slice,omitempty"`
		NotSerializable int              `json:"-"`
		unexported      int              //nolint:unused
	}

	testCases := map[string]testStruct{
		"int=123":                             {Int: 123},
		"bool=true":                           {Bool: true},
		"string=example":                      {String: "example"},
		"int_slice=1&int_slice=2&int_slice=3": {IntSlice: []int{1, 2, 3}},
		"int_slice[]=1&int_slice[]=2&int_slice[]=3":    {IntSlice: []int{1, 2, 3}},
		"int_slice[2]=3&int_slice[1]=2&int_slice[0]=1": {IntSlice: []int{1, 2, 3}},
		"struct[int]=789&struct[string]=abc":           {Struct: testSubStruct{Int: 789, String: "abc"}},
		"struct_slice[0][int]=3&struct_slice[0][string]=a&struct_slice[1][int]=2&struct_slice[1][string]=b&struct_slice[2][int]=1&struct_slice[2][string]=c": {
			StructSlice: []testSubStruct{{Int: 3, String: "a"}, {Int: 2, String: "b"}, {Int: 1, String: "c"}},
		},
		"opt_int=456":    {OptInt: func(v int) *int { return &v }(456)},
		"opt_bool=true":  {OptBool: func(v bool) *bool { return &v }(true)},
		"opt_string=def": {OptString: func(v string) *string { return &v }("def")},
		"opt_struct[int]=456&opt_struct[string]=def": {OptStruct: &testSubStruct{Int: 456, String: "def"}},
		"opt_struct_slice[0][int]=123&opt_struct_slice[0][string]=abc&opt_struct_slice[1][int]=456&opt_struct_slice[1][string]=def": {
			OptStructSlice: &([]testSubStruct{{Int: 123, String: "abc"}, {Int: 456, String: "def"}}),
		},
		"opt_struct[additional_property]=123": {
			OptStruct: &testSubStruct{AdditionalProperties: map[string]string{"additional_property": "123"}},
		},
	}

	for k, v := range testCases {
		values, err := url.ParseQuery(k)
		assert.NoError(t, err)
		var result testStruct
		err = BindForm(&result, values, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, v, result)
	}
}

func TestBindMultipartForm(t *testing.T) {
	var testStruct struct {
		File     types.File    `json:"file"`
		OptFile  *types.File   `json:"opt_file,omitempty"`
		Files    []types.File  `json:"files"`
		OptFiles *[]types.File `json:"opt_files"`
	}

	form, err := makeMultipartFilesForm([]fileData{{field: "file", filename: "123.txt", content: []byte("123")}})
	assert.NoError(t, err)
	err = BindForm(&testStruct, form.Value, form.File, nil)
	assert.NoError(t, err)
	assert.Equal(t, "123.txt", testStruct.File.Filename())
	content, err := testStruct.File.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, []byte("123"), content)

	form, err = makeMultipartFilesForm([]fileData{
		{field: "files", filename: "123.pdf", content: []byte("123")},
		{field: "files", filename: "456.pdf", content: []byte("456")},
		{field: "files", filename: "789.pdf", content: []byte("789")},
	})
	assert.NoError(t, err)
	err = BindForm(&testStruct, form.Value, form.File, nil)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(testStruct.Files))
	assert.Equal(t, "123.pdf", testStruct.Files[0].Filename())
	assert.Equal(t, "456.pdf", testStruct.Files[1].Filename())
	assert.Equal(t, "789.pdf", testStruct.Files[2].Filename())

	form, err = makeMultipartFilesForm([]fileData{{field: "opt_file", filename: "456.png", content: []byte("456")}})
	assert.NoError(t, err)
	err = BindForm(&testStruct, form.Value, form.File, nil)
	assert.NoError(t, err)
	assert.Equal(t, "456.png", testStruct.OptFile.Filename())
	content, err = testStruct.OptFile.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, []byte("456"), content)

	form, err = makeMultipartFilesForm([]fileData{
		{field: "opt_files[2]", filename: "123.pdf", content: []byte("123")},
		{field: "opt_files[1]", filename: "456.pdf", content: []byte("456")},
		{field: "opt_files[0]", filename: "789.pdf", content: []byte("789")},
	})
	assert.NoError(t, err)
	err = BindForm(&testStruct, form.Value, form.File, nil)
	assert.NoError(t, err)
	assert.NotNil(t, testStruct.OptFiles)
	assert.Equal(t, 3, len(*testStruct.OptFiles))
	assert.Equal(t, "789.pdf", (*testStruct.OptFiles)[0].Filename())
	assert.Equal(t, "456.pdf", (*testStruct.OptFiles)[1].Filename())
	assert.Equal(t, "123.pdf", (*testStruct.OptFiles)[2].Filename())
}

func TestMarshalForm(t *testing.T) {
	type testSubStruct struct {
		Int    int    `json:"int"`
		String string `json:"string"`
	}
	type testStruct struct {
		Int             int              `json:"int,omitempty"`
		Bool            bool             `json:"bool,omitempty"`
		String          string           `json:"string,omitempty"`
		IntSlice        []int            `json:"int_slice,omitempty"`
		Struct          testSubStruct    `json:"struct,omitempty"`
		StructSlice     []testSubStruct  `json:"struct_slice,omitempty"`
		OptInt          *int             `json:"opt_int,omitempty"`
		OptBool         *bool            `json:"opt_bool,omitempty"`
		OptString       *string          `json:"opt_string,omitempty"`
		OptStruct       *testSubStruct   `json:"opt_struct,omitempty"`
		OptStructSlice  *[]testSubStruct `json:"opt_struct_slice,omitempty"`
		NotSerializable int              `json:"-"`
		unexported      int              //nolint:unused
	}

	testCases := map[string]testStruct{
		"int=123":        {Int: 123},
		"bool=true":      {Bool: true},
		"string=example": {String: "example"},
		"int_slice[0]=1&int_slice[1]=2&int_slice[2]=3": {IntSlice: []int{1, 2, 3}},
		"struct[int]=789&struct[string]=abc":           {Struct: testSubStruct{Int: 789, String: "abc"}},
		"struct_slice[0][int]=3&struct_slice[0][string]=a&struct_slice[1][int]=2&struct_slice[1][string]=b&struct_slice[2][int]=1&struct_slice[2][string]=c": {
			StructSlice: []testSubStruct{{Int: 3, String: "a"}, {Int: 2, String: "b"}, {Int: 1, String: "c"}},
		},
		"opt_int=456":    {OptInt: func(v int) *int { return &v }(456)},
		"opt_bool=true":  {OptBool: func(v bool) *bool { return &v }(true)},
		"opt_string=def": {OptString: func(v string) *string { return &v }("def")},
		"opt_struct[int]=456&opt_struct[string]=def": {OptStruct: &testSubStruct{Int: 456, String: "def"}},
		"opt_struct_slice[0][int]=123&opt_struct_slice[0][string]=abc&opt_struct_slice[1][int]=456&opt_struct_slice[1][string]=def": {
			OptStructSlice: &([]testSubStruct{{Int: 123, String: "abc"}, {Int: 456, String: "def"}}),
		},
	}

	for k, v := range testCases {
		marshaled, err := MarshalForm(v, nil)
		assert.NoError(t, err)
		encoded, err := url.QueryUnescape(marshaled.Encode())
		assert.NoError(t, err)
		assert.Equal(t, k, encoded)
	}
}

type fileData struct {
	field    string
	filename string
	content  []byte
}

func makeMultipartFilesForm(files []fileData) (*multipart.Form, error) {
	var buffer bytes.Buffer
	mw := multipart.NewWriter(&buffer)
	for _, file := range files {
		w, err := mw.CreateFormFile(file.field, file.filename)
		if err != nil {
			return nil, err
		}
		_, err = w.Write(file.content)
		if err != nil {
			return nil, err
		}
	}
	err := mw.Close()
	if err != nil {
		return nil, err
	}
	mr := multipart.NewReader(&buffer, mw.Boundary())
	return mr.ReadForm(1024)
}
