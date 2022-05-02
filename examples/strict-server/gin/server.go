//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --package=api --generate types,gin,spec,strict-server -o server.gen.go ../strict-schema.yaml

package api

import (
	"context"
	"io"
	"mime/multipart"
)

type StrictServer struct {
}

func (s StrictServer) JSONExample(ctx context.Context, request JSONExampleRequestObject) interface{} {
	return JSONExample200JSONResponse(*request.Body)
}

func (s StrictServer) MultipartExample(ctx context.Context, request MultipartExampleRequestObject) interface{} {
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
	})
}

func (s StrictServer) MultipleRequestAndResponseTypes(ctx context.Context, request MultipleRequestAndResponseTypesRequestObject) interface{} {
	switch {
	case request.Body != nil:
		return MultipleRequestAndResponseTypes200ImagepngResponse{Body: request.Body}
	case request.JSONBody != nil:
		return MultipleRequestAndResponseTypes200JSONResponse(*request.JSONBody)
	case request.FormdataBody != nil:
		return MultipleRequestAndResponseTypes200FormdataResponse(*request.FormdataBody)
	case request.TextBody != nil:
		return MultipleRequestAndResponseTypes200TextResponse(*request.TextBody)
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
		})
	default:
		return MultipleRequestAndResponseTypes400TextResponse("content type is not supported")
	}
}

func (s StrictServer) TextExample(ctx context.Context, request TextExampleRequestObject) interface{} {
	return TextExample200TextResponse(*request.Body)
}

func (s StrictServer) UnknownExample(ctx context.Context, request UnknownExampleRequestObject) interface{} {
	return UnknownExample200Videomp4Response{Body: request.Body}
}

func (s StrictServer) UnspecifiedContentType(ctx context.Context, request UnspecifiedContentTypeRequestObject) interface{} {
	return UnspecifiedContentType200VideoResponse{Body: request.Body, ContentType: request.ContentType}
}

func (s StrictServer) URLEncodedExample(ctx context.Context, request URLEncodedExampleRequestObject) interface{} {
	return URLEncodedExample200FormdataResponse(*request.Body)
}

func (s StrictServer) HeadersExample(ctx context.Context, request HeadersExampleRequestObject) interface{} {
	return HeadersExample200JSONResponse{Body: Example(*request.Body), Headers: HeadersExample200ResponseHeaders{Header1: request.Params.Header1, Header2: *request.Params.Header2}}
}
