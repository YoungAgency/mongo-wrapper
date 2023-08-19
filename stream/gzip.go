package stream

import (
	"bytes"
	"compress/gzip"
	"io"
)

type GzipStreamDecoder struct {
}

func (d *GzipStreamDecoder) Decode(raw []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GzipMongoStreamEncoder implements StreamEncoder
type GzipMongoStreamEncoder struct {
}

func (d *GzipMongoStreamEncoder) Encode(raw []byte) ([]byte, error) {
	var b bytes.Buffer
	writer := gzip.NewWriter(&b)
	_, err := writer.Write(raw)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
