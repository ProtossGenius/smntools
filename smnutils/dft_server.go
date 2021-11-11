package smnutils

import (
	"fmt"
	"net"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_err"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_net"
	"github.com/ProtossGenius/smnric/net_libs/smn_rpc"
	"google.golang.org/protobuf/proto"
)

// RICRunner remote interface call runner.
type RICRunner struct {
	ric     smn_rpc.RICSvrItf
	retChan chan RPCRet
	OnErr   smn_err.OnErr
}

// RPCRet rpc result.
type RPCRet struct {
	Dict int32
	Res  proto.Message
	Err  error
}

// RunRICExample run ric example.
func RunRICExample(port int, retChanSize int, ric smn_rpc.RICSvrItf) (res *RICRunner, err error) {
	res = &RICRunner{
		ric:     ric,
		retChan: make(chan RPCRet, retChanSize),
		OnErr:   smn_err.DftOnErr,
	}

	var svr *smn_net.TcpServer

	if svr, err = smn_net.NewTcpServer(port, func(conn net.Conn) {
		adapter := smn_rpc.NewMessageAdapter(conn)
		// async write result.
		go func() {
			for {
				_ret := <-res.retChan
				_, err := adapter.WriteRet(_ret.Dict, _ret.Res, _ret.Err)
				if err != nil {
					res.OnErr(err)
				}
			}
		}()

		// async read message.
		for {
			msg, err := adapter.ReadCall()
			res.OnErr(err)
			dict, _res, err := res.ric.OnMessage(msg, adapter.GetConn())
			res.retChan <- RPCRet{dict, _res, err}
		}
	}); err != nil {
		return nil, fmt.Errorf("RunRICExample:%w", err)
	}

	go svr.Run()

	return res, nil
}
