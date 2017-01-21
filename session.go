package bidirpc

import (
	"bytes"
	"fmt"
	"io"
	"net/rpc"
	"sync"
)

var (
	streamTypeYin  byte = 1
	streamTypeYang byte = 2
)

const (
	defaultBufferPoolSize = 16
)

// Session is a bi-direction RPC connection.
type Session struct {
	conn      io.ReadWriteCloser
	yinOrYang bool
	writeLock sync.Mutex
	bp        *bufferPool

	streamYin  *stream
	streamYang *stream

	client   *rpc.Client
	registry *Registry

	closeLock sync.Mutex
	closed    bool
	closedC   chan struct{}
}

// NewSession creates a new session.
func NewSession(conn io.ReadWriteCloser, yinOrYang bool, registry *Registry, bufferPoolSize int) (*Session, error) {
	if bufferPoolSize == 0 {
		bufferPoolSize = defaultBufferPoolSize
	}
	s := &Session{
		conn:      conn,
		yinOrYang: yinOrYang,
		bp:        newBufferPool(bufferPoolSize),
		closedC:   make(chan struct{}),
	}

	s.streamYin = newStream(s, streamTypeYin)
	s.streamYang = newStream(s, streamTypeYang)

	var cliCodec *clientCodec
	var svrCodec *serverCodec
	if yinOrYang {
		cliCodec = newClientCodec(s.streamYin)
		svrCodec = newServerCodec(s.streamYang)
	} else {
		cliCodec = newClientCodec(s.streamYang)
		svrCodec = newServerCodec(s.streamYin)
	}
	s.client = rpc.NewClientWithCodec(cliCodec)
	s.registry = registry

	go s.registry.server.ServeCodec(svrCodec)

	return s, nil
}

// Serve starts the event loop, this is a blocking call.
func (s *Session) Serve() error {
	err := s.readLoop()
	if err != nil && err != io.ErrClosedPipe && err != io.EOF {
		return err
	}

	return nil
}

// Go invokes the function asynchronously. It returns the Call structure representing
// the invocation. The done channel will signal when the call is complete by returning
// the same Call object. If done is nil, Go will allocate a new channel.
// If non-nil, done must be buffered or Go will deliberately crash.
func (s *Session) Go(serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	return s.client.Go(serviceMethod, args, reply, done)
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (s *Session) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return s.client.Call(serviceMethod, args, reply)
}

// Close closes the session.
func (s *Session) Close() error {
	return s.doClose()
}

func (s *Session) readLoop() error {
	var err error
	var header [4]byte
	var streamType byte
	var bodyLen int
	reader := io.LimitedReader{R: s.conn}
	defer func() {
		// Swallow the close error
		_ = s.doClose()
	}()

	for {
		_, err = io.ReadFull(s.conn, header[:])
		if err != nil {
			return err
		}

		streamType, bodyLen = decodeHeader(header[:])
		if (streamType != streamTypeYin && streamType != streamTypeYang) || (bodyLen <= 0) {
			return fmt.Errorf("read a invalid header")
		}

		body := s.bp.Get()
		body.Grow(bodyLen)
		reader.N = int64(bodyLen)
		_, err = io.Copy(body, &reader)
		if err != nil {
			s.bp.Put(body)
			return err
		}

		var inC *chan *bytes.Buffer
		switch streamType {
		case streamTypeYin:
			inC = &s.streamYin.inC
		case streamTypeYang:
			inC = &s.streamYang.inC
		}
		select {
		case <-s.closedC:
			return nil
		case *inC <- body:
			// do nothing
		}
	}
}

func (s *Session) write(bytes []byte) error {
	s.writeLock.Lock()
	defer s.writeLock.Unlock()

	_, err := s.conn.Write(bytes)
	if err != nil {
		if closeErr := s.doClose(); closeErr != nil {
			return closeErr
		}
	}
	return err
}

func (s *Session) doClose() error {
	s.closeLock.Lock()
	defer s.closeLock.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	close(s.closedC)
	connErr := s.conn.Close()
	err := s.client.Close()
	if connErr != nil {
		return connErr
	}
	return err
}
