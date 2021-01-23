package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_net"
	"github.com/ProtossGenius/smnric/net_libs/smn_port_forward"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var errLogFile *os.File
var errCount = 0
var MAX_COUNT int

var loglock sync.Mutex

func newLogFile() *os.File {
	f, _ := smn_file.SafeOpenFile("./" + now() + ".err.log")
	return f
}
func logErr(err error) bool {
	loglock.Lock()
	defer loglock.Unlock()
	if err == nil {
		return false
	}
	if errCount >= MAX_COUNT {
		errLogFile.Close()
		errLogFile = newLogFile()
	}
	errStr := fmt.Sprintf("%s[ERROR]: %s\n", now(), err.Error())
	if errLogFile != nil {
		_, _e := errLogFile.WriteString(errStr)
		if _e != nil {
			fmt.Println(_e)
			errLogFile.Close()
			errLogFile = newLogFile()
		}
		errCount++
	}
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
func closeErrLogFile() {
	if errLogFile != nil {
		errLogFile.Close()
	}
}
func main() {
	localPort := flag.Int("l", 900, "local port")
	flag.StringVar(&server, "s", "127.0.0.1:901", "server host")
	flag.IntVar(&MAX_COUNT, "ec", 100000, "how many error logs in one file")
	flag.Parse()
	var err error
	errLogFile = newLogFile()
	defer closeErrLogFile()
	svr, err := smn_net.NewTcpServer(*localPort, handleConn)
	check(err)
	svr.OnErr = logErr2
	svr.Run()
}
