package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_err"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_exec"
)

func check(err error) {
	smn_err.DftOnErr(err)
}

func main() {
	fmt.Println("Start build.go")
	defer fmt.Println("Finish build.go")

	scriptsPath := "./build-scrips/"
	hasError := false
	fileInfos, err := ioutil.ReadDir(scriptsPath)
	check(err)

	for _, info := range fileInfos {
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			continue
		}

		fmt.Println("running build-script: ", info.Name(), "......")
		err = smn_exec.EasyDirExec("./", "go", "run", scriptsPath+info.Name())

		if err != nil {
			fmt.Println("Error when run ", info.Name(), ", error is : ", err.Error())

			hasError = true
		}
	}

	if hasError {
		panic("some error happened.")
	}
}
