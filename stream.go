package bidirpc

import (
	"bytes"
	"errors"
)

type stream struct {
	session *Session
	id      byte
	inC     chan *bytes.Buffer
	reader  *bytes.Buffer
	writer  bytes.Buffer
}

func newStream(session *Session, id byte) *stream {
	s := &stream{
		session: session,
		id:      id,
		inC:     make(chan *bytes.Buffer, 1),
	}
	return s
}

func (s *stream) Read(p []byte) (n int, err error) {
	if s.reader != nil && s.reader.Len() == 0 {
		s.session.bp.Put(s.reader)
		s.reader = nil
	}

	if s.reader == nil {
		select {
		case <-s.session.closedC:
			return 0, errors.New("stream read from a closed session")
		case s.reader = <-s.inC: // only the message body
		}
	}

	return s.reader.Read(p)
}

func (s *stream) Write(p []byte) (n int, err error) {
	if s.writer.Len() == 0 {
		var dummyHeader [4]byte
		s.writer.Write(dummyHeader[:])
	}
	return s.writer.Write(p)
}

func (s *stream) flush() (err error) {
	buffer := s.writer.Bytes()
	bodyLen := s.writer.Len() - 4
	encodeHeader(buffer, s.id, bodyLen)

	err = s.session.write(buffer)

	s.writer.Reset()
	return
}
