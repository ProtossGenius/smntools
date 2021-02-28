package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_net"
	"github.com/ProtossGenius/smnric/net_libs/smn_rpc"
	"github.com/ProtossGenius/smntools/rpc_nitf/cltrpc/clt_rpc_smntitf"
	"github.com/ProtossGenius/smntools/rpc_nitf/svrrpc/svr_rpc_smntitf"
	"google.golang.org/protobuf/proto"
)

type HelloStruct struct {
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func (h *HelloStruct) Hello(a string) string {
	return "anything"
}

type RPCRet struct {
	Dict int32
	Res  proto.Message
	Err  error
}

func AccpterRun(adapter smn_rpc.MessageAdapterItf) {
	rpcSvr := svr_rpc_smntitf.NewSvrRpcHelloItf(&HelloStruct{})
	retChan := make(chan *RPCRet, 10000)

	go func() {
		for {
			_ret := <-retChan
			if _ret == nil {
				break
			}

			_, err := adapter.WriteRet(_ret.Dict, _ret.Res, _ret.Err)
			check(err)
		}
	}()

	for {
		msg, err := adapter.ReadCall()
		check(err)
		dict, res, err := rpcSvr.OnMessage(msg, adapter.GetConn())
		retChan <- &RPCRet{dict, res, err}
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
	start := time.Now().UnixNano()
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()

	var wg sync.WaitGroup

	go RunSvr()

	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:900")
	check(err)

	client := clt_rpc_smntitf.NewCltRpcHelloItf(smn_rpc.NewMessageAdapter(conn), 10000)

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		client.Hello("world", func(str string) {
			wg.Done()
		})
	}

	wg.Wait()

	fmt.Println("time use = ", (time.Now().UnixNano()-start)/1e6, "ms")
}
