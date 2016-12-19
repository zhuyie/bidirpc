package bidirpc

import (
	"net"
	"testing"
)

type Args struct {
	Name string
}

type Reply struct {
	Msg string
}

type Demo int

func (d *Demo) SayHi(args Args, reply *Reply) error {
	reply.Msg = "Hi " + args.Name
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

	d0 := new(Demo)
	sessionYin.Register(d0)
	d1 := new(Demo)
	sessionYang.Register(d1)

	args := Args{"Mac"}
	reply := new(Reply)
	err = sessionYin.Call("Demo.SayHi", args, reply)
	if err != nil {
		t.Fatalf("Call error: %v", err)
	}
	t.Logf("reply = %v", reply.Msg)

	sessionYin.Close()
	sessionYang.Close()
}
