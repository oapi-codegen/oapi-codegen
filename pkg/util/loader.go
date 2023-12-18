package util

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func LoadFromData(data []byte, basePath ...string) (swagger *libopenapi.DocumentModel[v3.Document], err error) {
	basePath_ := ""
	if len(basePath) > 0 {
		basePath_ = basePath[0]
	}
	document, err := libopenapi.NewDocumentWithConfiguration(data, &datamodel.DocumentConfiguration{
		AllowFileReferences:   true,
		BasePath:              basePath_,
		AllowRemoteReferences: true,
	})
	if err != nil {
		return nil, err
	}

	modelv3, errs := document.BuildV3Model()
	return modelv3, errors.Join(errs...)
}

func LoadOpenAPI(filePath string) (swagger *libopenapi.DocumentModel[v3.Document], err error) {
	var b []byte
	u, err := url.Parse(filePath)
	if err == nil && u.Scheme != "" && u.Host != "" {
		r, err := http.Get(u.String())
		if err != nil {
			return nil, err
		}
		if r.StatusCode != http.StatusOK {
			return nil, errors.New("received non 200 status code on GET request")
		}
		defer r.Body.Close()
		b, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		return LoadFromData(b)
	} else {
		b, err = os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		return LoadFromData(b, path.Dir(filePath))
	}
}
