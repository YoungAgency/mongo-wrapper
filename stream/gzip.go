package stream

import (
	"bytes"
	"compress/gzip"
	"io"
)

type GzipEncoderDecoder struct {
}

func (d *GzipEncoderDecoder) Encode(raw []byte) ([]byte, error) {
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

func (d *GzipEncoderDecoder) Decode(raw []byte) ([]byte, error) {
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
