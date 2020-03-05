package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
)

func check(err error){
	if err != nil{
		panic(err)
	}
}


func analysisGoLine(line string, inImport bool) (isImport, mutiImport bool, pkgName string) {
	if strings.HasPrefix(line, ")") {
		return false, false, ""
	}
	if inImport || strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "import(") {
		mutiImport = inImport
		if strings.Contains(line, "(") {
			mutiImport = true
		}
		if strings.Index(line, "\"") == -1 {
			return false, mutiImport, ""
		}
		pkgName = line[strings.Index(line, "\"")+1:]
		if strings.Index(pkgName, "\"") == -1 {
			return false, mutiImport, ""
		}
		pkgName = pkgName[:strings.Index(pkgName, "\"")]
		fmt.Println(pkgName)
		isImport = true
	}
	if strings.Contains(line, ")") {
		mutiImport = false
	}
	return
}

func analysis(responsePath string) {
	smn_file.DeepTraversalDir(responsePath, func(path string, info os.FileInfo) smn_file.FileDoFuncResult {
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
		}
		data, err := smn_file.FileReadAll(path)
		check(err)
		inImport := false
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(strings.Split(line, "//")[0])
			var isImport bool
			var pkg string
			isImport, inImport, pkg = analysisGoLine(line, inImport)
			fmt.Println(line, "+++ ", isImport, inImport, pkg)
		}
		return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
	})
}

func main(){
	analysis("./")
}