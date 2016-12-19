package bidirpc

import (
	"bytes"
	"errors"
)

type stream struct {
	id             byte
	sessionClosedC <-chan struct{}
	reader         *bytes.Buffer
	inC            chan []byte
	writer         *bytes.Buffer
	outC           chan []byte
}

func newStream(id byte, sessionClosedC <-chan struct{}) *stream {
	s := &stream{
		id:             id,
		sessionClosedC: sessionClosedC,
		inC:            make(chan []byte),
		outC:           make(chan []byte),
	}
	return s
}

func (s *stream) Read(p []byte) (n int, err error) {
	if s.reader == nil || s.reader.Len() == 0 {
		select {
		case <-s.sessionClosedC:
			return 0, errors.New("stream read from a closed session")
		case data := <-s.inC: // only the message body
			s.reader = bytes.NewBuffer(data)
		}
	}
	return s.reader.Read(p)
}

func (s *stream) Write(p []byte) (n int, err error) {
	if s.writer == nil {
		s.writer = bytes.NewBuffer(nil)
		var dummyHeader [4]byte
		s.writer.Write(dummyHeader[:])
	}
	return s.writer.Write(p)
}

func (s *stream) flush() error {
	buffer := s.writer.Bytes()
	bodyLen := s.writer.Len() - 4
	encodeHeader(buffer, s.id, bodyLen)
	select {
	case <-s.sessionClosedC:
		return errors.New("stream flush to a closed session")
	case s.outC <- buffer:
	}

	s.writer.Truncate(4) // for dummy header
	return nil
}
