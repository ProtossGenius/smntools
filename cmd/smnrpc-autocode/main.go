package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_data"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_err"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_pglang"
	"github.com/ProtossGenius/SureMoonNet/smn/analysis/smn_rpc_itf"
)

/*
rem:
	1. in target language, protobuf's compile path can't change. if target-language  not go, interface's package like go package, can't change.
	2. interface's proto package name always startWith rip_ (rpc-interface-proto).
	3. maybe will delete proto-read. because auto-code. can read/write proto-bytes in function.

	work do ..

	(itf -> lang-itf) //call smn_goitf2lang

	itf -> proto.
	itf -> rpc-code

	compile proto

*/
type AutoCodeCfg struct {
	ItfPath   string            `json:"itf_path"   node:"go-interface path"`
	Target    map[string]string `json:"target"     node:"target to output path. such as {"go_c":"./clt/", "go_s":"./svr/") ...  target: [lang]_[c/s]"`
	ProtoPath string            `json:"proto_path" node:"proto file path."`
	Module    string            `json:"module"    node:"project's go-package."`

	//language suit. not must .
}

func checkerr(err error) {
	smn_err.DftOnErr(err)
}

// flag var
var (
	cfg string
	doc bool
)

func readCfg() *AutoCodeCfg {
	data, err := smn_file.FileReadAll(cfg)
	checkerr(err)
	cfgStruct := &AutoCodeCfg{}
	err = smn_data.GetDataFromStr(string(data), &cfgStruct)
	checkerr(err)
	return cfgStruct
}

func printDoc() {
	//use tag product doc.
	rt := reflect.TypeOf(AutoCodeCfg{})
	numField := rt.NumField()
	for i := 0; i < numField; i++ {
		field := rt.Field(i)

		fmt.Println(field.Tag.Get("node"))
	}

}

func itf2proto(itf *smn_pglang.ItfDef) {

}

func autocode() {
	c := readCfg()
	itfs, err := smn_rpc_itf.GetItfListFromDir(c.ItfPath)
	checkerr(err)
	for path, list := range itfs {
		fullPath, err := filepath.Abs(path)
		checkerr(err)
		pwdPath, err := filepath.Abs("./")
		fullPkg := c.Module + strings.Replace(fullPath, pwdPath, "", -1)
		fmt.Println("package is: ", list[0].Package, "; path is :", path, "; full package is :", fullPkg)
		for _, itf := range list {
			itf2proto(itf)
		}
	}
}

func main() {
	flag.StringVar(&cfg, "cfg", "./auto-code-cfg.json", "a config file to describ smnrpc-autocode should do what.")
	flag.BoolVar(&doc, "doc", false, "show the doc about cfg file.")
	flag.Parse()
	if doc {
		printDoc()
		return
	}
	autocode()
}
