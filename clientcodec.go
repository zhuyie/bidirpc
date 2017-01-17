package bidirpc

import (
	"encoding/gob"
	"net/rpc"
)

type clientCodec struct {
	stream *stream
	dec    *gob.Decoder
	enc    *gob.Encoder
}

func newClientCodec(s *stream) *clientCodec {
	c := &clientCodec{
		stream: s,
		dec:    gob.NewDecoder(s),
		enc:    gob.NewEncoder(s),
	}
	return c
}

func (c *clientCodec) WriteRequest(r *rpc.Request, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		return
	}
	if err = c.enc.Encode(body); err != nil {
		return
	}
	return c.stream.flush()
}

func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	return c.dec.Decode(r)
}

func (c *clientCodec) ReadResponseBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *clientCodec) Close() error {
	return nil
}
