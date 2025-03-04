package api

import (
	"net/http"
	"time"
)

// Chunk represents a piece of data with its associated timing
type writtenChunk struct {
	bytes []byte
	ts    time.Duration
}

// Custom ResponseWriter with writtenChunk tracking and timing
type streamRecordWriter struct {
	ops       []writtenChunk
	startTime time.Time
	headers   http.Header
	Code      int
}

func newStreamResponseWriter() *streamRecordWriter {
	return &streamRecordWriter{
		startTime: time.Now(),
		ops:       make([]writtenChunk, 0),
		headers:   make(http.Header),
		Code:      0,
	}
}

func (crw *streamRecordWriter) Write(data []byte) (int, error) {
	// copy the data to avoid the slice being modified if the caller modifies it
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	newChunk := writtenChunk{
		bytes: dataCopy,
		ts:    time.Since(crw.startTime),
	}
	crw.ops = append(crw.ops, newChunk)
	return len(data), nil
}

func (crw *streamRecordWriter) Header() http.Header {
	return crw.headers
}

func (crw *streamRecordWriter) WriteHeader(statusCode int) {
	crw.Code = statusCode
}

func (crw *streamRecordWriter) Flush() {
}
