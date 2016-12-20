package bidirpc

import (
	"fmt"
	"net"
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
	callCount int
}

func (s *Service) SayHi(args Args, reply *Reply) error {
	reply.Msg = fmt.Sprintf("[%v] Hi %v, from %v", s.callCount, args.Name, s.name)
	s.callCount++
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
