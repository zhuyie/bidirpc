package bidirpc

import (
	"net"
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
)

func init() {
	connYin, connYang := net.Pipe()
	sessionYin, _ = NewSession(connYin, true)
	sessionYang, _ = NewSession(connYang, false)
	service := &BenchService{}
	sessionYin.Register(service)
}

func BenchmarkEchoInt(b *testing.B) {
	args := IntArgs{}
	reply := new(IntReply)
	for i := 0; i < b.N; i++ {
		args.V = int32(i)
		sessionYang.Call("BenchService.EchoInt", args, reply)
	}
}

func BenchmarkEchoString(b *testing.B) {
	args := StringArgs{"abcdefghijklmnopqrstuvwxyz"}
	reply := new(StringReply)
	for i := 0; i < b.N; i++ {
		sessionYang.Call("BenchService.EchoString", args, reply)
	}
}
