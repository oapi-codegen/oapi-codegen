//go:build go1.22

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=server.cfg.yaml ../strict-schema.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=types.cfg.yaml ../strict-schema.yaml

package api

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
)

type StrictServer struct{}

type MockServer struct {
	JSONExampleMock                     func(w http.ResponseWriter, r *http.Request)
	MultipartExampleMock                func(w http.ResponseWriter, r *http.Request)
	MultipartRelatedExampleMock         func(w http.ResponseWriter, r *http.Request)
	MultipleRequestAndResponseTypesMock func(w http.ResponseWriter, r *http.Request)
	ReservedGoKeywordParametersMock     func(w http.ResponseWriter, r *http.Request, pType string)
	ReusableResponsesMock               func(w http.ResponseWriter, r *http.Request)
	TextExampleMock                     func(w http.ResponseWriter, r *http.Request)
	UnknownExampleMock                  func(w http.ResponseWriter, r *http.Request)
	UnspecifiedContentTypeMock          func(w http.ResponseWriter, r *http.Request)
	URLEncodedExampleMock               func(w http.ResponseWriter, r *http.Request)
	HeadersExampleMock                  func(w http.ResponseWriter, r *http.Request, params HeadersExampleParams)
	UnionExampleMock                    func(w http.ResponseWriter, r *http.Request)
}

func (s StrictServer) JSONExample(ctx context.Context, request JSONExampleRequestObject) (JSONExampleResponseObject, error) {
	return JSONExample200JSONResponse(*request.Body), nil
}

func (s StrictServer) MultipartExample(ctx context.Context, request MultipartExampleRequestObject) (MultipartExampleResponseObject, error) {
	return MultipartExample200MultipartResponse(func(writer *multipart.Writer) error {
		for {
			part, err := request.Body.NextPart()
			if err == io.EOF {
				return nil
			} else if err != nil {
				return err
			}
			w, err := writer.CreatePart(part.Header)
			if err != nil {
				return err
			}
			_, err = io.Copy(w, part)
			if err != nil {
				return err
			}
			if err = part.Close(); err != nil {
				return err
			}
		}
	}), nil
}

func (s StrictServer) MultipartRelatedExample(ctx context.Context, request MultipartRelatedExampleRequestObject) (MultipartRelatedExampleResponseObject, error) {
	return MultipartRelatedExample200MultipartResponse(func(writer *multipart.Writer) error {
		for {
			part, err := request.Body.NextPart()
			if err == io.EOF {
				return nil
			} else if err != nil {
				return err
			}
			w, err := writer.CreatePart(part.Header)
			if err != nil {
				return err
			}
			_, err = io.Copy(w, part)
			if err != nil {
				return err
			}
			if err = part.Close(); err != nil {
				return err
			}
		}
	}), nil
}

func (s StrictServer) MultipleRequestAndResponseTypes(ctx context.Context, request MultipleRequestAndResponseTypesRequestObject) (MultipleRequestAndResponseTypesResponseObject, error) {
	switch {
	case request.Body != nil:
		return MultipleRequestAndResponseTypes200ImagepngResponse{Body: request.Body}, nil
	case request.JSONBody != nil:
		return MultipleRequestAndResponseTypes200JSONResponse(*request.JSONBody), nil
	case request.FormdataBody != nil:
		return MultipleRequestAndResponseTypes200FormdataResponse(*request.FormdataBody), nil
	case request.TextBody != nil:
		return MultipleRequestAndResponseTypes200TextResponse(*request.TextBody), nil
	case request.MultipartBody != nil:
		return MultipleRequestAndResponseTypes200MultipartResponse(func(writer *multipart.Writer) error {
			for {
				part, err := request.MultipartBody.NextPart()
				if err == io.EOF {
					return nil
				} else if err != nil {
					return err
				}
				w, err := writer.CreatePart(part.Header)
				if err != nil {
					return err
				}
				_, err = io.Copy(w, part)
				if err != nil {
					return err
				}
				if err = part.Close(); err != nil {
					return err
				}
			}
		}), nil
	default:
		return MultipleRequestAndResponseTypes400Response{}, nil
	}
}

func (s StrictServer) TextExample(ctx context.Context, request TextExampleRequestObject) (TextExampleResponseObject, error) {
	return TextExample200TextResponse(*request.Body), nil
}

func (s StrictServer) UnknownExample(ctx context.Context, request UnknownExampleRequestObject) (UnknownExampleResponseObject, error) {
	return UnknownExample200Videomp4Response{Body: request.Body}, nil
}

func (s StrictServer) UnspecifiedContentType(ctx context.Context, request UnspecifiedContentTypeRequestObject) (UnspecifiedContentTypeResponseObject, error) {
	return UnspecifiedContentType200VideoResponse{Body: request.Body, ContentType: request.ContentType}, nil
}

func (s StrictServer) URLEncodedExample(ctx context.Context, request URLEncodedExampleRequestObject) (URLEncodedExampleResponseObject, error) {
	return URLEncodedExample200FormdataResponse(*request.Body), nil
}

func (s StrictServer) HeadersExample(ctx context.Context, request HeadersExampleRequestObject) (HeadersExampleResponseObject, error) {
	return HeadersExample200JSONResponse{Body: *request.Body, Headers: HeadersExample200ResponseHeaders{Header1: request.Params.Header1, Header2: *request.Params.Header2}}, nil
}

func (s StrictServer) ReusableResponses(ctx context.Context, request ReusableResponsesRequestObject) (ReusableResponsesResponseObject, error) {
	return ReusableResponses200JSONResponse{ReusableresponseJSONResponse: ReusableresponseJSONResponse{Body: *request.Body}}, nil
}

func (s StrictServer) ReservedGoKeywordParameters(ctx context.Context, request ReservedGoKeywordParametersRequestObject) (ReservedGoKeywordParametersResponseObject, error) {
	return ReservedGoKeywordParameters200TextResponse(""), nil
}

func (s StrictServer) UnionExample(ctx context.Context, request UnionExampleRequestObject) (UnionExampleResponseObject, error) {
	union, err := json.Marshal(*request.Body)
	if err != nil {
		return nil, err
	}

	return UnionExample200JSONResponse{
		Body: struct{ union json.RawMessage }{
			union: union,
		},
	}, nil
}

func (m *MockServer) JSONExample(w http.ResponseWriter, r *http.Request) {
	if m.JSONExampleMock != nil {
		m.JSONExampleMock(w, r)
	}
}

func (m *MockServer) MultipartExample(w http.ResponseWriter, r *http.Request) {
	if m.MultipartExampleMock != nil {
		m.MultipartExampleMock(w, r)
	}
}

func (m *MockServer) MultipartRelatedExample(w http.ResponseWriter, r *http.Request) {
	if m.MultipartRelatedExampleMock != nil {
		m.MultipartRelatedExampleMock(w, r)
	}
}

func (m *MockServer) MultipleRequestAndResponseTypes(w http.ResponseWriter, r *http.Request) {
	if m.MultipleRequestAndResponseTypesMock != nil {
		m.MultipleRequestAndResponseTypesMock(w, r)
	}
}

func (m *MockServer) ReservedGoKeywordParameters(w http.ResponseWriter, r *http.Request, pType string) {
	if m.ReservedGoKeywordParametersMock != nil {
		m.ReservedGoKeywordParametersMock(w, r, pType)
	}
}

func (m *MockServer) ReusableResponses(w http.ResponseWriter, r *http.Request) {
	if m.ReusableResponsesMock != nil {
		m.ReusableResponsesMock(w, r)
	}
}

func (m *MockServer) TextExample(w http.ResponseWriter, r *http.Request) {
	if m.TextExampleMock != nil {
		m.TextExampleMock(w, r)
	}
}

func (m *MockServer) UnknownExample(w http.ResponseWriter, r *http.Request) {
	if m.UnknownExampleMock != nil {
		m.UnknownExampleMock(w, r)
	}
}

func (m *MockServer) UnspecifiedContentType(w http.ResponseWriter, r *http.Request) {
	if m.UnspecifiedContentTypeMock != nil {
		m.UnspecifiedContentTypeMock(w, r)
	}
}

func (m *MockServer) URLEncodedExample(w http.ResponseWriter, r *http.Request) {
	if m.URLEncodedExampleMock != nil {
		m.URLEncodedExampleMock(w, r)
	}
}

func (m *MockServer) HeadersExample(w http.ResponseWriter, r *http.Request, params HeadersExampleParams) {
	if m.HeadersExampleMock != nil {
		m.HeadersExampleMock(w, r, params)
	}
}

func (m *MockServer) UnionExample(w http.ResponseWriter, r *http.Request) {
	if m.UnionExampleMock != nil {
		m.UnionExampleMock(w, r)
	}
}
