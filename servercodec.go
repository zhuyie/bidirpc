package bidirpc

import (
	"encoding/gob"
	"net/rpc"
)

type serverCodec struct {
	stream *stream
	dec    *gob.Decoder
	enc    *gob.Encoder
}

func newServerCodec(s *stream) *serverCodec {
	c := &serverCodec{
		stream: s,
		dec:    gob.NewDecoder(s),
		enc:    gob.NewEncoder(s),
	}
	return c
}

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	return c.dec.Decode(r)
}

func (c *serverCodec) ReadRequestBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *serverCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	if err = c.enc.Encode(r); err != nil {
		return
	}
	if err = c.enc.Encode(body); err != nil {
		return
	}
	return c.stream.flush()
}

func (c *serverCodec) Close() error {
	return nil
}
