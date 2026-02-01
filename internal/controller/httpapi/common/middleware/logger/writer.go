package logger

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

type writer struct {
	gin.ResponseWriter
	buffer    *bytes.Buffer
	limit     int
	truncated bool
}

func newWriter(w gin.ResponseWriter, limit int) *writer {
	return &writer{
		ResponseWriter: w,
		buffer:         bytes.NewBuffer(nil),
		limit:          limit,
	}
}

func (writer *writer) Write(bytes []byte) (int, error) {
	writer.capture(bytes)
	return writer.ResponseWriter.Write(bytes)
}

func (writer *writer) WriteString(str string) (int, error) {
	writer.capture([]byte(str))
	return writer.ResponseWriter.WriteString(str)
}

func (writer *writer) capture(bytes []byte) {
	if len(bytes) == 0 || writer.truncated || writer.limit == 0 {
		return
	}

	if writer.limit < 0 {
		_, _ = writer.buffer.Write(bytes)
		return
	}

	remain := writer.limit - writer.buffer.Len()
	if remain <= 0 {
		writer.truncated = true
		return
	}

	if len(bytes) > remain {
		_, _ = writer.buffer.Write(bytes[:remain])
		writer.truncated = true
		return
	}

	_, _ = writer.buffer.Write(bytes)
}

func (writer *writer) bytes() []byte {
	if writer.buffer == nil || writer.buffer.Len() == 0 {
		return nil
	}
	out := make([]byte, writer.buffer.Len())
	copy(out, writer.buffer.Bytes())
	return out
}
