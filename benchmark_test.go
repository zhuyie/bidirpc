package bidirpc

import (
	"net"
	"net/rpc"
	"testing"
)

type IntArgs struct {
	V int32
}

type IntReply struct {
	V int32
}

type StringArgs struct {
	Str string
}

type StringReply struct {
	Str string
}

type BenchService struct{}

func (s *BenchService) EchoInt(args IntArgs, reply *IntReply) error {
	reply.V = args.V
	return nil
}

func (s *BenchService) EchoString(args StringArgs, reply *StringReply) error {
	reply.Str = args.Str
	return nil
}

var (
	sessionYin  *Session
	sessionYang *Session
	client      *rpc.Client
	server      *rpc.Server
)

func init() {
	service := &BenchService{}

	connYin, connYang := net.Pipe()
	sessionYin, _ = NewSession(connYin, true, 0)
	sessionYang, _ = NewSession(connYang, false, 0)
	sessionYin.Register(service)
	go func() {
		_ = sessionYin.Serve()
	}()
	go func() {
		_ = sessionYang.Serve()
	}()

	connServer, connClient := net.Pipe()
	client = rpc.NewClient(connClient)
	server = rpc.NewServer()
	server.Register(service)
	go server.ServeConn(connServer)
}

func BenchmarkEchoInt(b *testing.B) {
	args := IntArgs{}
	reply := new(IntReply)
	for i := 0; i < b.N; i++ {
		args.V = int32(i)
		sessionYang.Call("BenchService.EchoInt", args, reply)
	}
}

func BenchmarkBuiltinEchoInt(b *testing.B) {
	args := IntArgs{}
	reply := new(IntReply)
	for i := 0; i < b.N; i++ {
		args.V = int32(i)
		client.Call("BenchService.EchoInt", args, reply)
	}
}

func BenchmarkEchoString(b *testing.B) {
	args := StringArgs{"abcdefghijklmnopqrstuvwxyz"}
	reply := new(StringReply)
	for i := 0; i < b.N; i++ {
		sessionYang.Call("BenchService.EchoString", args, reply)
	}
}

func BenchmarkBuiltinEchoString(b *testing.B) {
	args := StringArgs{"abcdefghijklmnopqrstuvwxyz"}
	reply := new(StringReply)
	for i := 0; i < b.N; i++ {
		client.Call("BenchService.EchoString", args, reply)
	}
}
