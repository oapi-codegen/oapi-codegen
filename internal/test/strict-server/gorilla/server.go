//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=server.cfg.yaml ../strict-schema.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=types.cfg.yaml ../strict-schema.yaml

package api

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
)

type StrictServer struct {
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
