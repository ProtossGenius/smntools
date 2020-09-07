package main

import (
	"fmt"
	"net"
	"time"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_net"
	"github.com/ProtossGenius/SureMoonNet/smn/net_libs/smn_rpc"
	"github.com/ProtossGenius/smntools/rpc_nitf/cltrpc/clt_rpc_smntitf"
	"github.com/ProtossGenius/smntools/rpc_nitf/svrrpc/svr_rpc_smntitf"
)

type HelloStruct struct {
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
func (h *HelloStruct) Hello(a string) string {
	if a == "hello" {
		return "world"
	}
	return "why not hello?"
}
func AccpterRun(adapter smn_rpc.MessageAdapterItf) {
	rpcSvr := svr_rpc_smntitf.NewSvrRpcHelloItf(&HelloStruct{})
	for {
		msg, err := adapter.ReadCall()
		check(err)
		dict, res, err := rpcSvr.OnMessage(msg, adapter.GetConn())
		adapter.WriteRet(int32(dict), res, err)
	}
}

func accept(conn net.Conn) {
	adapter := smn_rpc.NewMessageAdapter(conn)
	go AccpterRun(adapter)
}

func RunSvr() {
	svr, err := smn_net.NewTcpServer(900, accept)
	check(err)
	svr.Run()
}

func main() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()

	go RunSvr()
	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:900")
	check(err)

	client := clt_rpc_smntitf.NewCltRpcHelloItf(smn_rpc.NewMessageAdapter(conn))
	client.Hello("hello", func(str string) {
		fmt.Println(str)
	})
	client.Hello("", func(str string) {
		fmt.Println(str)
	})

	time.Sleep(1 * time.Second)
}
