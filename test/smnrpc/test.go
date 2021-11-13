package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ProtossGenius/smnric/net_libs/smn_rpc"
	"github.com/ProtossGenius/smntools/rpc_nitf/cltrpc/clt_rpc_smnitf"
	"github.com/ProtossGenius/smntools/rpc_nitf/svrrpc/svr_rpc_smnitf"
	"github.com/ProtossGenius/smntools/smnutils"
)

// HelloStruct impl.
type HelloStruct struct{}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Hello hello.
func (h *HelloStruct) Hello(a string) string {
	return "anything"
}

func main() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()

	start := time.Now().UnixNano()

	var wg sync.WaitGroup

	_, err := smnutils.RunRICExample(900, 1000, svr_rpc_smnitf.NewSvrRpcHelloItf(&HelloStruct{}))
	check(err)

	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", "127.0.0.1:900")
	check(err)

	client := clt_rpc_smnitf.NewCltRpcHelloItf(smn_rpc.NewMessageAdapter(conn), 1000)

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		client.Hello("world", func(str string) {
			wg.Done()
		})
	}

	wg.Wait()

	fmt.Println("time use = ", (time.Now().UnixNano()-start)/1e6, "ms")
}
