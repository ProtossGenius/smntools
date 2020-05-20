package main

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_data"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_err"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/SureMoonNet/smn/analysis/smn_rpc_itf"
	"github.com/ProtossGenius/SureMoonNet/smn/proto_tool/goitf2lang"
	"github.com/ProtossGenius/SureMoonNet/smn/proto_tool/itf2proto"
	"github.com/ProtossGenius/SureMoonNet/smn/proto_tool/itf2rpc"
	"github.com/ProtossGenius/SureMoonNet/smn/proto_tool/proto_compile"
)

/*
rem:
	1. in target language, protobuf's compile path can't change. if target-language  not go,
	interface's package like go package, can't change.
	2. interface's proto package name always startWith rip_ (rpc-interface-proto).

	work do ..

	(itf -> lang-itf) //call smn_goitf2lang

	itf -> proto.
	itf -> rpc-code

	compile proto

*/

//AutoCodeCfg json.
type AutoCodeCfg struct {
	ItfPath   string            `json:"itf_path"   node:"go-interface path"`
	Target    map[string]string `json:"target"     node:"target to output path."`
	ProtoPath string            `json:"proto_path" node:"proto file path."`
	Module    string            `json:"module"     node:"project's go-package."`
	Src       string            `json:"src"        node:"code path, such as java is ./src/ "`
	//language suit. not must .
}

func checkerr(err error) {
	smn_err.DftOnErr(err)
}

func readCfg(cfg string) *AutoCodeCfg {
	data, err := smn_file.FileReadAll(cfg)
	checkerr(err)

	cfgStruct := &AutoCodeCfg{}
	err = smn_data.GetDataFromStr(string(data), &cfgStruct)
	checkerr(err)
	if cfgStruct.Src == "" {
		cfgStruct.Src = "./"
	}
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

func autocode(cfg string) {
	c := readCfg(cfg)
	itfs, err := smn_rpc_itf.GetItfListFromDir(c.ItfPath)
	checkerr(err)

	langMap := make(map[string]bool)

	for target := range c.Target {
		lang := strings.Split(target, "_")[0]
		langMap[lang] = true
	}

	for path, list := range itfs {
		//go interface to proto.
		err := itf2proto.WriteProto(c.ProtoPath, list)
		checkerr(err)

		for lang := range langMap {
			//go interface to lang interface.
			goitf2lang.WriteInterface(lang, c.Src, list[0].Package, list)
		}

		fullPath, err := filepath.Abs(path)
		checkerr(err)
		pwdPath, err := filepath.Abs("./")
		checkerr(err)
		//get fullPkg
		fullPkg := c.Module + strings.Replace(fullPath, pwdPath, "", -1)
		fmt.Println("package is: ", list[0].Package, "; path is :", path, "; full package is :", fullPkg)
		//write RPC code
		for target, oPath := range c.Target {
			for _, itf := range list {
				err = itf2rpc.Write(target, oPath, c.Module, fullPkg, itf)
				checkerr(err)
			}
		}
	}

	errList := []string{}
	//proto compile
	for lang := range langMap {
		err := proto_compile.Compile(c.ProtoPath, c.Src+"/./pb/", c.Module, lang)

		if err != nil {
			errList = append(errList, fmt.Sprintf("\tWhen compile lang [%s], error is %s", lang, err.Error()))
		}
	}

	if len(errList) != 0 {
		checkerr(errors.New("Error When compile proto, error List: \n" + strings.Join(errList, "\n")))
	}
}

func main() {
	// flag var
	var (
		cfg string
		doc bool
	)

	flag.StringVar(&cfg, "cfg", "./auto-code-cfg.json", "a config file to describ smnrpc-autocode should do what.")
	flag.BoolVar(&doc, "doc", false, "show the doc about cfg file.")
	flag.Parse()

	if doc {
		printDoc()
		return
	}

	autocode(cfg)
}
