package main

import (
	"flag"
	"os"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var (
		target string
	)

	flag.StringVar(&target, "target", ".", "target path.")
	flag.Parse()

	_, err := smn_file.DeepTraversalDir(target, func(path string, info os.FileInfo) smn_file.FileDoFuncResult {
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
		}

		datas, err := smn_file.FileReadAll(path)
		check(err)

		f, err := smn_file.CreateNewFile(path)
		check(err)
		_, err = f.WriteString(strings.ReplaceAll(string(datas), "google.golang.org/protobuf/proto",
			"google.golang.org/protobuf/proto"))
		check(err)
		f.Close()
		return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
	})

	check(err)
}
