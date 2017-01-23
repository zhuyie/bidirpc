package bidirpc

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"testing"
)

type Args struct {
	Name string
}

type Reply struct {
	Msg string
}

type Service struct {
	name      string
	callCount int32
}

func (s *Service) SayHi(args Args, reply *Reply) error {
	reply.Msg = fmt.Sprintf("[%v] Hi %v, from %v", atomic.LoadInt32(&s.callCount), args.Name, s.name)
	atomic.AddInt32(&s.callCount, 1)
	return nil
}

func TestBasic(t *testing.T) {
	connYin, connYang := net.Pipe()

	registryYin := NewRegistry()
	registryYang := NewRegistry()

	serviceYin := &Service{name: "Yin"}
	err := registryYin.Register(serviceYin)
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}

	serviceYang := &Service{name: "Yang"}
	err = registryYang.Register(serviceYang)
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}

	sessionYin, err := NewSession(connYin, Yin, registryYin, 0)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}
	sessionYang, err := NewSession(connYang, Yang, registryYang, 0)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	sessionWait := sync.WaitGroup{}
	sessionWait.Add(2)
	go func() {
		if err := sessionYin.Serve(); err != nil {
			t.Fatalf("Eventloop error: %v", err)
		}
		sessionWait.Done()
	}()
	go func() {
		if err := sessionYang.Serve(); err != nil {
			t.Fatalf("Eventloop error: %v", err)
		}
		sessionWait.Done()
	}()

	for i := 0; i < 3; i++ {
		args := Args{"Windows"}
		reply := new(Reply)
		err = sessionYin.Call("Service.SayHi", args, reply)
		if err != nil {
			t.Fatalf("Call error: %v", err)
		}
		t.Logf("reply = %v\n", reply.Msg)

		args.Name = "OSX"
		err = sessionYang.Call("Service.SayHi", args, reply)
		if err != nil {
			t.Fatalf("Call error: %v", err)
		}
		t.Logf("reply = %v\n", reply.Msg)

		args.Name = "iOS"
		err = sessionYin.Call("Service.SayHi", args, reply)
		if err != nil {
			t.Fatalf("Call error: %v", err)
		}
		t.Logf("reply = %v\n", reply.Msg)
	}

	err = registryYang.RegisterName("NewService", serviceYang)
	if err != nil {
		t.Fatalf("RegisterName error: %v", err)
	}
	args := Args{"Linux"}
	reply := new(Reply)
	call := sessionYin.Go("NewService.SayHi", args, reply, nil)
	<-call.Done
	if call.Error != nil {
		t.Fatalf("Go error: %v", call.Error)
	}
	t.Logf("reply = %v\n", reply.Msg)

	sessionYin.Close()
	sessionYang.Close()
	sessionWait.Wait()
}

func TestReadError(t *testing.T) {
	connYin, connYang := net.Pipe()
	connYang.Close()

	sessionYin, err := NewSession(connYin, Yin, NewRegistry(), 0)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	sessionWait := sync.WaitGroup{}
	sessionWait.Add(1)
	go func() {
		if err := sessionYin.Serve(); err != nil {
			t.Fatalf("Eventloop error: %v", err)
		}
		sessionWait.Done()
	}()

	args := Args{"Windows"}
	reply := new(Reply)
	err = sessionYin.Call("Service.SayHi", args, reply)
	if err == nil {
		t.Fatal("Call should return error, got nil")
	}

	sessionYin.Close()
	sessionWait.Wait()
}

func TestWriteError(t *testing.T) {
	connYin, _ := net.Pipe()
	connYin.Close()

	sessionYin, err := NewSession(connYin, Yin, NewRegistry(), 0)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	sessionWait := sync.WaitGroup{}
	sessionWait.Add(1)
	go func() {
		if err := sessionYin.Serve(); err != nil {
			t.Fatalf("Eventloop error: %v", err)
		}
		sessionWait.Done()
	}()

	args := Args{"Windows"}
	reply := new(Reply)
	err = sessionYin.Call("Service.SayHi", args, reply)
	if err == nil {
		t.Fatal("Call should return error, got nil")
	}

	sessionYin.Close()
	sessionWait.Wait()
}

func TestWriteError2(t *testing.T) {
	_, connYang := net.Pipe()
	connYang.Close()

	sessionYang, err := NewSession(connYang, Yang, NewRegistry(), 0)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	sessionWait := sync.WaitGroup{}
	sessionWait.Add(1)
	go func() {
		if err := sessionYang.Serve(); err != nil {
			t.Fatalf("Eventloop error: %v", err)
		}
		sessionWait.Done()
	}()

	args := Args{"Windows"}
	reply := new(Reply)
	err = sessionYang.Call("Service.SayHi", args, reply)
	if err == nil {
		t.Fatal("Call should return error, got nil")
	}

	sessionYang.Close()
	sessionWait.Wait()
}

func TestReadInvalidHeader(t *testing.T) {
	connYin, connYang := net.Pipe()

	sessionYin, err := NewSession(connYin, Yin, NewRegistry(), 0)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	sessionWait := sync.WaitGroup{}
	sessionWait.Add(1)
	go func() {
		if err := sessionYin.Serve(); err != nil {
			if err == nil {
				t.Fatal("Call should return error, got nil")
			}
		}
		sessionWait.Done()
	}()

	var header [4]byte
	connYang.Write(header[:])

	args := Args{"Windows"}
	reply := new(Reply)
	err = sessionYin.Call("Service.SayHi", args, reply)
	if err == nil {
		t.Fatal("Call should return error, got nil")
	}

	sessionYin.Close()
	sessionWait.Wait()
}

func TestReadBodyError(t *testing.T) {
	connYin, connYang := net.Pipe()

	sessionYin, err := NewSession(connYin, Yin, NewRegistry(), 0)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	sessionWait := sync.WaitGroup{}
	sessionWait.Add(1)
	go func() {
		if err := sessionYin.Serve(); err != nil {
			t.Fatalf("Eventloop error: %v", err)
		}
		sessionWait.Done()
	}()

	var header [4]byte
	encodeHeader(header[:], byte(Yang), 10)
	connYang.Write(header[:])
	connYang.Close()

	args := Args{"Windows"}
	reply := new(Reply)
	err = sessionYin.Call("Service.SayHi", args, reply)
	if err == nil {
		t.Fatal("Call should return error, got nil")
	}

	sessionYin.Close()
	sessionWait.Wait()
}

func TestConcurrent(t *testing.T) {
	connYin, connYang := net.Pipe()

	registryYin := NewRegistry()
	registryYang := NewRegistry()

	serviceYin := &Service{name: "Yin"}
	err := registryYin.Register(serviceYin)
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}

	serviceYang := &Service{name: "Yang"}
	err = registryYang.Register(serviceYang)
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}

	sessionYin, err := NewSession(connYin, Yin, registryYin, 0)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}
	sessionYang, err := NewSession(connYang, Yang, registryYang, 0)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	sessionWait := sync.WaitGroup{}
	sessionWait.Add(2)
	go func() {
		if err := sessionYin.Serve(); err != nil {
			t.Fatalf("Eventloop error: %v", err)
		}
		sessionWait.Done()
	}()
	go func() {
		if err := sessionYang.Serve(); err != nil {
			t.Fatalf("Eventloop error: %v", err)
		}
		sessionWait.Done()
	}()

	var GoroutineCount = 6
	var CallCount = 2
	var wg sync.WaitGroup
	wg.Add(GoroutineCount * 2)
	for i := 0; i < GoroutineCount; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i <= CallCount; i++ {
				args := Args{"Anakin Skywalker"}
				reply := new(Reply)
				err := sessionYin.Call("Service.SayHi", args, reply)
				if err != nil {
					t.Fatalf("Call error: %v", err)
				}
				t.Logf("reply = %v\n", reply.Msg)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i <= CallCount; i++ {
				args := Args{"Darth Vader"}
				reply := new(Reply)
				err := sessionYang.Call("Service.SayHi", args, reply)
				if err != nil {
					t.Fatalf("Call error: %v", err)
				}
				t.Logf("reply = %v\n", reply.Msg)
			}
		}()

	}
	wg.Wait()

	sessionYin.Close()
	sessionYang.Close()
	sessionWait.Wait()
}

func ExampleSession() {
	var conn io.ReadWriteCloser

	// Create a registry, and register your available services
	registry := NewRegistry()
	registry.Register(&Service{})

	// TODO: Establish your connection before passing it to the session

	// Create a new session
	session, err := NewSession(conn, Yin, registry, 0)
	if err != nil {
		log.Fatal(err)
	}
	// Clean up session resources
	defer func() {
		if err := session.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Start the event loop, this is a blocking call, so place it in a goroutine
	// if you need to move on.  The call will return when the connection is
	// terminated.
	if err = session.Serve(); err != nil {
		log.Fatal(err)
	}
}
