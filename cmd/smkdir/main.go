package main

import (
	"flag"
	"fmt"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_exec"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
)

// mkdir mkdir and create README.md .
func mkdir(dir string) {
	if err := smn_file.MakeSureDirExist(dir); err != nil {
		fmt.Println("when create dir fail, err is ", err.Error())
		return
	}

	if err := smn_exec.EasyDirExec(".", "vim", dir+"/README.md"); err != nil {
		fmt.Println("when try create README.md fail, err is ", err.Error())
	}
}

func main() {
	flag.Parse()
	args := flag.Args()

	for _, arg := range args {
		mkdir(arg)
	}
}
