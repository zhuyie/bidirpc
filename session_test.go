package bidirpc

import (
	"fmt"
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

	sessionYin, err := NewSession(connYin, true)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}
	sessionYang, err := NewSession(connYang, false)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	serviceYin := &Service{name: "Yin"}
	err = sessionYin.Register(serviceYin)
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}

	serviceYang := &Service{name: "Yang"}
	sessionYang.Register(serviceYang)
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}

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

	sessionYang.RegisterName("NewService", serviceYang)
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
}

func TestReadError(t *testing.T) {
	connYin, connYang := net.Pipe()
	connYang.Close()

	sessionYin, err := NewSession(connYin, true)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	args := Args{"Windows"}
	reply := new(Reply)
	err = sessionYin.Call("Service.SayHi", args, reply)
	if err == nil {
		t.Fatal("Call should return error, got nil")
	}

	sessionYin.Close()
}

func TestWriteError(t *testing.T) {
	connYin, _ := net.Pipe()
	connYin.Close()

	sessionYin, err := NewSession(connYin, true)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	args := Args{"Windows"}
	reply := new(Reply)
	err = sessionYin.Call("Service.SayHi", args, reply)
	if err == nil {
		t.Fatal("Call should return error, got nil")
	}

	sessionYin.Close()
}

func TestWriteError2(t *testing.T) {
	_, connYang := net.Pipe()
	connYang.Close()

	sessionYang, err := NewSession(connYang, false)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	args := Args{"Windows"}
	reply := new(Reply)
	err = sessionYang.Call("Service.SayHi", args, reply)
	if err == nil {
		t.Fatal("Call should return error, got nil")
	}

	sessionYang.Close()
}

func TestReadInvalidHeader(t *testing.T) {
	connYin, connYang := net.Pipe()

	sessionYin, err := NewSession(connYin, true)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	var header [4]byte
	connYang.Write(header[:])

	args := Args{"Windows"}
	reply := new(Reply)
	err = sessionYin.Call("Service.SayHi", args, reply)
	if err == nil {
		t.Fatal("Call should return error, got nil")
	}

	sessionYin.Close()
}

func TestReadBodyError(t *testing.T) {
	connYin, connYang := net.Pipe()

	sessionYin, err := NewSession(connYin, true)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	var header [4]byte
	encodeHeader(header[:], streamTypeYang, 10)
	connYang.Write(header[:])
	connYang.Close()

	args := Args{"Windows"}
	reply := new(Reply)
	err = sessionYin.Call("Service.SayHi", args, reply)
	if err == nil {
		t.Fatal("Call should return error, got nil")
	}

	sessionYin.Close()
}

func TestConcurrent(t *testing.T) {
	connYin, connYang := net.Pipe()

	sessionYin, err := NewSession(connYin, true)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}
	sessionYang, err := NewSession(connYang, false)
	if err != nil {
		t.Fatalf("NewSession error: %v", err)
	}

	serviceYin := &Service{name: "Yin"}
	err = sessionYin.Register(serviceYin)
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}

	serviceYang := &Service{name: "Yang"}
	sessionYang.Register(serviceYang)
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}

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
}
