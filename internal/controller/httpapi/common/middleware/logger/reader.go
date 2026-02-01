package logger

import (
	"bytes"
	"io"
	"strings"
)

type reader struct {
	readCloser io.ReadCloser
	buffer     *bytes.Buffer
	limit      int
}

func newReader(readCloser io.ReadCloser, limit int) *reader {
	if readCloser == nil {
		readCloser = io.NopCloser(strings.NewReader(""))
	}
	return &reader{
		readCloser: readCloser,
		buffer:     bytes.NewBuffer(nil),
		limit:      limit,
	}
}

func (reader *reader) Read(bytes []byte) (int, error) {
	bytesRead, err := reader.readCloser.Read(bytes)
	if bytesRead <= 0 || reader.limit == 0 {
		return bytesRead, err
	}

	if reader.limit < 0 {
		_, _ = reader.buffer.Write(bytes[:bytesRead])
		return bytesRead, err
	}

	if reader.buffer.Len() < reader.limit {
		remain := reader.limit - reader.buffer.Len()
		if bytesRead > remain {
			_, _ = reader.buffer.Write(bytes[:remain])
		} else {
			_, _ = reader.buffer.Write(bytes[:bytesRead])
		}
	}

	return bytesRead, err
}

func (reader *reader) Close() error { return reader.readCloser.Close() }

func (reader *reader) bytes() []byte {
	if reader.buffer == nil || reader.buffer.Len() == 0 {
		return nil
	}
	out := make([]byte, reader.buffer.Len())
	copy(out, reader.buffer.Bytes())
	return out
}
