package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_net"
	"github.com/ProtossGenius/SureMoonNet/smn/net_libs/smn_port_forward"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var errLogFile *os.File
var errCount = 0

func logErr(err error) bool {
	if err == nil {
		return false
	}
	errStr := fmt.Sprintf("%s[ERROR]: %s\n", now(), err.Error())
	errLogFile.WriteString(errStr)
	fmt.Println(errStr)
	return true
}

func logErr2(err error) {
	logErr(err)
}

var server string

func now() string {
	n := time.Now()
	return fmt.Sprintf("%d%02d%02d%02d%02d%02d", n.Year(), n.Month(), n.Day(),
		n.Hour(), n.Minute(), n.Second())
}

func handleConn(c net.Conn) {
	fmt.Println("accept connect ...", c.RemoteAddr().String())
	raddr, err := net.ResolveTCPAddr("tcp", server)
	if logErr(err) {
		return
	}
	lConn, err := net.DialTCP("tcp", nil, raddr)
	if logErr(err) {
		return
	}
	pf := smn_port_forward.NewPortForwardWorker()
	pf.SetInOut(c, lConn)
	pf.DoWork(logErr2)
}

func main() {
	localPort := flag.Int("l", 900, "local port")
	flag.StringVar(&server, "s", "127.0.0.1:901", "server host")
	flag.Parse()
	var err error
	errLogFile, err = smn_file.SafeOpenFile("./smpf.err.log")
	check(err)
	defer errLogFile.Close()
	svr, err := smn_net.NewTcpServer(*localPort, handleConn)
	check(err)
	svr.OnErr = logErr2
	svr.Run()
}
