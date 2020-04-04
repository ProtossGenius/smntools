package main

import (
	"crypto/rand"
	"crypto/rsa"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_net"
	"github.com/ProtossGenius/SureMoonNet/smn/net_libs/smn_port_forward"
	"github.com/ProtossGenius/smntools/auto_code/smntac_asppl"
	_ "github.com/ProtossGenius/smntools/auto_code/smntac_asppl"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var nofile bool
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
	if err == nil || err.Error() == "EOF" {
		return false
	}
	errStr := fmt.Sprintf("%s[ERROR]: %s\n", now(), err.Error())
	fmt.Println(errStr)
	if nofile {
		return true
	}
	if errCount >= MAX_COUNT {
		errLogFile.Close()
		errLogFile = newLogFile()
	}
	if errLogFile != nil {
		_, _e := errLogFile.WriteString(errStr)
		if _e != nil {
			fmt.Println(_e)
			errLogFile.Close()
			errLogFile = newLogFile()
		}
		errCount++
	}
	return true
}

func logErr2(err error) {
	logErr(err)
}

var server string
var asclt bool

func now() string {
	n := time.Now()
	return fmt.Sprintf("%d%02d%02d%02d%02d%02d", n.Year(), n.Month(), n.Day(),
		n.Hour(), n.Minute(), n.Second())
}
func encode(bytes []byte) []byte {
	res, err := rsa.EncryptPKCS1v15(rand.Reader, smntac_asppl.PubKey(), bytes)
	check(err)
	return res
}

func decode(bytes []byte) []byte {
	res, err := rsa.DecryptPKCS1v15(rand.Reader, smntac_asppl.PriKey(), bytes)
	check(err)
	return res
}

func copyDescode(dst io.Writer, src io.Reader) (int64, error) {
	bts := make([]byte, smntac_asppl.OutLen)
	for {
		btl, err := src.Read(bts)
		if err != nil {
			return 0, err
		}
		fmt.Println("decode : len = ", btl)
		ecbts := decode(bts[:btl])
		_, err = dst.Write(ecbts)
		if err != nil {
			return 0, err
		}
	}
}

func copyEncode(dst io.Writer, src io.Reader) (int64, error) {
	bts := make([]byte, smntac_asppl.ReadLen)
	for {
		btl, err := src.Read(bts)
		if err != nil {
			return 0, err
		}
		fmt.Println("encode : len = ", btl)
		ecbts := encode(bts[:btl])
		_, err = dst.Write(ecbts)
		if err != nil {
			return 0, err
		}
	}
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
	if asclt {
		pf.CopyFromIn = copyEncode
		pf.CopyFromOut = copyDescode
	} else {
		pf.CopyFromOut = copyEncode
		pf.CopyFromIn = copyDescode
	}
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
	flag.BoolVar(&nofile, "nofile", true, "not error to logfile.")
	flag.BoolVar(&asclt, "asclt", false, "run as a client. default run as a server.")
	flag.Parse()
	var err error
	errLogFile = newLogFile()
	defer closeErrLogFile()
	svr, err := smn_net.NewTcpServer(*localPort, handleConn)
	check(err)
	svr.OnErr = logErr2
	svr.Run()
}
