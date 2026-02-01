package gzip

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	contentEncoding = "Content-Encoding"
	contentLength   = "Content-Length"
	vary            = "Vary"
	noWritten       = -1
)

type gzipMinWriter struct {
	underlyingWriter gin.ResponseWriter
	header           http.Header
	status           int
	size             int
	buf              bytes.Buffer
	gzWriter         *gzip.Writer
	passthrough      bool
	streaming        bool
	finalized        bool
	minSize          int
}

func newGzipMinWriter(underlyingWriter gin.ResponseWriter, minSize int) *gzipMinWriter {
	header := make(http.Header)
	if underlyingWriter != nil {
		for key, value := range underlyingWriter.Header() {
			copyValue := make([]string, len(value))
			copy(copyValue, value)
			header[key] = copyValue
		}
	}
	return &gzipMinWriter{
		underlyingWriter: underlyingWriter,
		header:           header,
		status:           defaultStatus,
		size:             noWritten,
		minSize:          minSize,
	}
}

func (writer *gzipMinWriter) Header() http.Header {
	if writer == nil {
		return nil
	}
	if writer.passthrough {
		return writer.underlyingWriter.Header()
	}
	return writer.header
}

func (writer *gzipMinWriter) WriteHeader(code int) {
	if writer == nil || writer.passthrough {
		if writer != nil {
			writer.underlyingWriter.WriteHeader(code)
		}
		return
	}

	if !writer.Written() && writer.minSize > 0 {
		if contentLengthStr := writer.header.Get(contentLength); contentLengthStr != "" {
			if size, err := strconv.Atoi(contentLengthStr); err == nil && size > 0 && size < writer.minSize {
				writer.passthrough = true
				for key, values := range writer.header {
					for _, value := range values {
						writer.underlyingWriter.Header().Add(key, value)
					}
				}
				writer.underlyingWriter.WriteHeader(code)
				return
			}
		}
	}

	if code > 0 && writer.status != code {
		if writer.Written() {
			return
		}
		writer.status = code
	}
}

func (writer *gzipMinWriter) WriteHeaderNow() {
	if writer == nil || writer.passthrough {
		if writer != nil {
			writer.underlyingWriter.WriteHeaderNow()
		}
		return
	}
	if !writer.Written() {
		writer.size = 0
	}
}

func (writer *gzipMinWriter) Write(data []byte) (int, error) {
	if writer == nil {
		return 0, nil
	}
	if writer.passthrough {
		return writer.underlyingWriter.Write(data)
	}
	writer.WriteHeaderNow()

	if len(data) == 0 {
		return 0, nil
	}

	if writer.streaming {
		size, err := writer.gzWriter.Write(data)
		writer.size += size
		return size, err
	}

	writer.buf.Write(data)
	bufferedSize := writer.buf.Len()

	if bufferedSize >= writer.minSize {
		writer.enableStreaming()
	}

	return len(data), nil
}

func (writer *gzipMinWriter) WriteString(stringValue string) (int, error) {
	if writer == nil {
		return 0, nil
	}

	if writer.passthrough {
		return writer.underlyingWriter.WriteString(stringValue)
	}

	writer.WriteHeaderNow()

	if stringValue == "" {
		return 0, nil
	}

	if writer.streaming {
		size, err := io.WriteString(writer.gzWriter, stringValue)
		writer.size += size
		return size, err
	}

	writer.buf.WriteString(stringValue)
	bufferedSize := writer.buf.Len()

	if bufferedSize >= writer.minSize {
		writer.enableStreaming()
	}

	return len(stringValue), nil
}

func (writer *gzipMinWriter) Status() int {
	if writer == nil {
		return defaultStatus
	}
	return writer.status
}

func (writer *gzipMinWriter) Size() int {
	if writer == nil {
		return 0
	}
	return writer.size
}

func (writer *gzipMinWriter) Written() bool {
	if writer == nil {
		return false
	}
	return writer.size != noWritten
}

func (writer *gzipMinWriter) Flush() {
	if writer == nil {
		return
	}
	if writer.passthrough {
		writer.underlyingWriter.Flush()
		return
	}
	if writer.streaming {
		// Flush gzip writer
		if writer.gzWriter != nil {
			_ = writer.gzWriter.Flush()
		}
		writer.underlyingWriter.Flush()
		return
	}
	writer.finalizeUncompressed()
	writer.passthrough = true
	writer.underlyingWriter.Flush()
}

func (writer *gzipMinWriter) CloseNotify() <-chan bool {
	if writer == nil {
		ch := make(chan bool)
		close(ch)
		return ch
	}
	return writer.underlyingWriter.CloseNotify()
}

func (writer *gzipMinWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if writer == nil {
		return nil, nil, nil
	}
	if !writer.passthrough && !writer.finalized && !writer.streaming {
		writer.finalizeUncompressed()
		writer.passthrough = true
	}
	if writer.streaming && writer.gzWriter != nil {
		_ = writer.gzWriter.Close()
	}
	return writer.underlyingWriter.Hijack()
}

func (writer *gzipMinWriter) Pusher() http.Pusher {
	if writer == nil {
		return nil
	}
	return writer.underlyingWriter.Pusher()
}

func (writer *gzipMinWriter) enableStreaming() {
	if writer.streaming || writer.finalized {
		return
	}
	writer.streaming = true

	for key, values := range writer.header {
		for _, value := range values {
			writer.underlyingWriter.Header().Add(key, value)
		}
	}
	writer.underlyingWriter.Header().Set(contentEncoding, gzipValue)
	writer.underlyingWriter.Header().Del(contentLength)
	writer.underlyingWriter.Header().Add(vary, acceptEncoding)
	writer.underlyingWriter.WriteHeader(writer.status)

	writer.gzWriter = gzip.NewWriter(writer.underlyingWriter)
	buffered := writer.buf.Bytes()
	if len(buffered) > 0 {
		size, _ := writer.gzWriter.Write(buffered)
		writer.size = size
	}
	writer.buf.Reset()
}

func (writer *gzipMinWriter) finalize(minSizeBytes int) {
	if writer == nil || writer.finalized {
		return
	}
	writer.finalized = true

	if writer.streaming {
		if writer.gzWriter != nil {
			_ = writer.gzWriter.Close()
		}
		return
	}

	body := writer.buf.Bytes()
	if writer.size == noWritten {
		writer.size = 0
	}

	if writer.status == http.StatusNoContent || len(body) == 0 {
		writer.finalizeUncompressed()
		return
	}

	if enc := strings.TrimSpace(writer.header.Get(contentEncoding)); enc != "" {
		writer.finalizeUncompressed()
		return
	}

	if minSizeBytes > 0 && len(body) < minSizeBytes {
		writer.finalizeUncompressed()
		return
	}

	var buffer bytes.Buffer
	writerLevel, err := gzip.NewWriterLevel(&buffer, gzip.BestSpeed)
	if err != nil {
		writer.finalizeUncompressed()
		return
	}
	if _, err := writerLevel.Write(body); err != nil {
		_ = writerLevel.Close()
		writer.finalizeUncompressed()
		return
	}
	if err := writerLevel.Close(); err != nil {
		writer.finalizeUncompressed()
		return
	}

	writer.header.Del(contentLength)
	writer.header.Set(contentEncoding, gzipValue)

	addVaryAcceptEncoding(writer.header)
	writer.header.Set(contentLength, strconv.Itoa(buffer.Len()))

	copyHeaders(writer.underlyingWriter.Header(), writer.header)
	writer.underlyingWriter.WriteHeader(writer.status)
	n, _ := io.Copy(writer.underlyingWriter, &buffer)
	writer.size = int(n)
}

func (writer *gzipMinWriter) finalizeUncompressed() {
	if writer == nil {
		return
	}

	if !writer.finalized {
		writer.finalized = true
	}

	body := writer.buf.Bytes()
	if writer.size == noWritten {
		writer.size = 0
	}

	writer.header.Del(contentLength)
	writer.header.Set(contentLength, strconv.Itoa(len(body)))
	copyHeaders(writer.underlyingWriter.Header(), writer.header)

	writer.underlyingWriter.WriteHeader(writer.status)
	writeSize, _ := writer.underlyingWriter.Write(body)
	writer.size = writeSize
}

func copyHeaders(destination, src http.Header) {
	for key := range destination {
		destination.Del(key)
	}
	for key, value := range src {
		copyValue := make([]string, len(value))
		copy(copyValue, value)
		destination[key] = copyValue
	}
}

func addVaryAcceptEncoding(header http.Header) {
	if header == nil {
		return
	}
	existingVary := header.Values(vary)
	for _, value := range existingVary {
		for _, part := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(part), acceptEncoding) {
				return
			}
		}
	}
	header.Add(vary, acceptEncoding)
}
