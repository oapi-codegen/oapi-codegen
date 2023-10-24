package types

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
)

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/types#File
type File struct {
	multipart *multipart.FileHeader
	data      []byte
	filename  string
}

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/types#InitFromMultipart
func (file *File) InitFromMultipart(header *multipart.FileHeader) {
	file.multipart = header
	file.data = nil
	file.filename = ""
}

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/types#InitFromBytes
func (file *File) InitFromBytes(data []byte, filename string) {
	file.data = data
	file.filename = filename
	file.multipart = nil
}

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/types#MarshalJSON
func (file File) MarshalJSON() ([]byte, error) {
	b, err := file.Bytes()
	if err != nil {
		return nil, err
	}
	return json.Marshal(b)
}

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/types#UnmarshalJSON
func (file *File) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &file.data)
}

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/types#Bytes
func (file File) Bytes() ([]byte, error) {
	if file.multipart != nil {
		f, err := file.multipart.Open()
		if err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
		return io.ReadAll(f)
	}
	return file.data, nil
}

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/types#Reader
func (file File) Reader() (io.ReadCloser, error) {
	if file.multipart != nil {
		return file.multipart.Open()
	}
	return io.NopCloser(bytes.NewReader(file.data)), nil
}

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/types#Filename
func (file File) Filename() string {
	if file.multipart != nil {
		return file.multipart.Filename
	}
	return file.filename
}

// Deprecated: This has been replaced by https://pkg.go.dev/github.com/oapi-codegen/runtime/types#FileSize
func (file File) FileSize() int64 {
	if file.multipart != nil {
		return file.multipart.Size
	}
	return int64(len(file.data))
}
