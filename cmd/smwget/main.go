package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_exec"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
)

func checkerr(err error, errInfo string) {
	if err != nil {
		fmt.Println(errInfo)
		panic(err)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func md5sum(path string) string {
	oInfo, oErr, err := smn_exec.DirExecGetOut(".", "md5sum", path)
	checkerr(err, oErr)

	md5 := strings.TrimSpace(strings.Split(oInfo, " ")[0])

	return md5
}

func main() {
	help := flag.Bool("h", false, "show help")
	flag.Parse()

	if *help {
		fmt.Println("usage:  smwget md5 file_name  url.")
		fmt.Println("if md5 is 0 will download with out check. always download current dir.")

		return
	}

	args := flag.Args()

	const MinArgNum = 3

	if len(args) < MinArgNum {
		fmt.Println("no enough arg to call smwget.")
	}

	md5 := args[0]
	fileName := args[1]
	url := args[2]

	if smn_file.IsFileExist(fileName) && md5sum(fileName) == md5 {
		fmt.Println("download success")
		return
	}

	check(smn_exec.EasyDirExec("./", "wget", url, "-O", fileName))
}
