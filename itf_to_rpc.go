package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_pglang"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_str"
	"github.com/ProtossGenius/SureMoonNet/smn/analysis/smn_rpc_itf"
	"github.com/ProtossGenius/SureMoonNet/smn/code_file_build"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var GOPATH string
var targetPaths []string

/** file as:
package xxxx
import(...)

*/

func goi64toi(ot, v string) (string, bool) {
	isArr, typ := smn_str.ProtoUseDeal(ot)
	if !strings.Contains(ot, typ) {
		if !isArr {
			if typ[0] == 'i' {
				return fmt.Sprintf("int(%s)", v), false
			} else {
				return fmt.Sprintf("uint(%s)", v), false
			}
		} else {
			if typ[0] == 'i' {
				return fmt.Sprintf("smn_rpc.Int64ArrToIntArr(%s)", v), true
			} else {
				return fmt.Sprintf("smn_rpc.UInt64ArrToUIntArr(%s)", v), true
			}
		}
	} else {
		return v, false
	}
}

func goitoi64(ot, v string) (string, bool) {
	isArr, typ := smn_str.ProtoUseDeal(ot)
	if !strings.Contains(ot, typ) {
		if !isArr {
			if typ[0] == 'i' {
				return fmt.Sprintf("int64(%s)", v), false
			} else {
				return fmt.Sprintf("uint64(%s)", v), false
			}
		} else {
			if typ[0] == 'i' {
				return fmt.Sprintf("smn_rpc.IntArrToInt64Arr(%s)", v), true
			} else {
				return fmt.Sprintf("smn_rpc.UIntArrToUInt64Arr(%s)", v), true
			}
		}
	} else {
		return v, false
	}
}

func writeSvrRpcFile(path string, list []*smn_pglang.ItfDef) {
	for _, itf := range list {
		file, err := smn_file.CreateNewFile(path + itf.Name + ".go")
		check(err)
		gof := code_file_build.NewGoFile("svr_rpc_"+itf.Package, file, "Product by SureMoonNet", "Author: ProtossGenius", "Auto-code should not change.")
		gof.AddImports(code_file_build.LocalImptTarget(GOPATH, targetPaths...))
		gof.Imports(itf.Package, "github.com/golang/protobuf/proto")
		{ // rpc struct
			b := gof.AddBlock("type SvrRpc%s struct", itf.Name)
			b.WriteLine("itf %s.%s", itf.Package, itf.Name)
			b.WriteLine("dicts []smn_dict.EDict")
			b.Imports("smn_dict")
		}
		{ // new func
			b := gof.AddBlock("func NewSvrRpc%s(itf %s.%s) *SvrRpc%s", itf.Name, itf.Package, itf.Name, itf.Name)
			b.WriteLine("list := make([]smn_dict.EDict, 0)")
			for _, f := range itf.Functions {
				b.WriteLine("list = append(list, smn_dict.EDict_rip_%s_%s_%s_Prm)", itf.Package, itf.Name, f.Name)
			}
			b.WriteLine("return &SvrRpc%s{itf:itf, dicts:list}", itf.Name)
		}
		{ // used message dict
			b := gof.AddBlock("func (this *SvrRpc%s)getEDictList() []smn_dict.EDict", itf.Name)
			b.WriteLine("return this.dicts")
		}
		{ // struct get net-package
			b := gof.AddBlock("func (this *SvrRpc%s)OnMessage(c *smn_base.Call, conn net.Conn) (_d smn_dict.EDict, _p proto.Message, _e error)", itf.Name)
			b.Imports("smn_base")
			b.Imports("smn_pbr")
			b.Imports("net")
			{ // rb = recover func
				b.WriteLine("defer func() {")
				ib := b.AddBlock("if err := recover(); err != nil {")
				ib.IndentationAdd(1)
				ib.WriteLine("_p = nil")
				ib.Imports("fmt")
				ib.WriteLine("_e = fmt.Errorf(\"%%v\", err)")
				b.WriteLine("}()")
			}
			b.WriteLine("_m := smn_pbr.GetMsgByDict(c.Msg,smn_dict.EDict(c.Dict))")
			sb := b.AddBlock("switch smn_dict.EDict(c.Dict)") //sb -> switch block
			for _, f := range itf.Functions {
				cb := sb.AddBlock("case smn_dict.EDict_rip_%s_%s_%s_Prm:", itf.Package, itf.Name, f.Name)
				cb.Imports("rip_" + itf.Package)
				cb.WriteLine("_d = smn_dict.EDict_rip_%s_%s_%s_Ret", itf.Package, itf.Name, f.Name)
				cb.WriteLine("_msg := _m.(*rip_%s.%s_%s_Prm)", itf.Package, itf.Name, f.Name)
				rets := ""
				for i := 0; i < len(f.Returns); i++ {
					if i != 0 {
						rets += ", "
					}
					rets += fmt.Sprintf("p%d", i)
				}
				if rets != "" {
					rets += " :="
				}
				cb.WriteToNewLine("%s this.itf.%s(", rets, f.Name)
				for i, r := range f.Params {
					if i != 0 {
						cb.Write(", ")
					}
					if strings.TrimSpace(r.Type) != "net.Conn" {
						pv, usmn := goi64toi(r.Type, "_msg."+smn_str.InitialsUpper(r.Var))
						if usmn {
							cb.Imports("smn_rpc")
						}
						cb.Write(pv)
					} else {
						cb.Write("conn")
					}
				}
				cb.Write(")\n")
				cb.WriteToNewLine("return _d, &rip_%s.%s_%s_Ret{", itf.Package, itf.Name, f.Name)
				for i, r := range f.Returns {
					if i != 0 {
						cb.Write(", ")
					}
					pv, usmn := goitoi64(r.Type, fmt.Sprintf("p%d", i))
					if usmn {
						cb.Imports("smn_rpc")
					}
					cb.Write("%s:%s", smn_str.InitialsUpper(r.Var), pv)
				}
				cb.WriteLine("}, nil")
			}
			b.WriteLine("return -1, nil, nil")
		}

		gof.Output()
	}
}

func writeClientRpcFile(path string, list []*smn_pglang.ItfDef) {
	for _, itf := range list {
		file, err := smn_file.CreateNewFile(path + itf.Name + ".go")
		check(err)
		gof := code_file_build.NewGoFile("clt_rpc_"+itf.Package, file, "Product by SureMoonNet", "Author: ProtossGenius", "Auto-code should not change.")
		gof.AddImports(code_file_build.LocalImptTarget(GOPATH, targetPaths...))
		gof.Imports(itf.Package, "github.com/golang/protobuf/proto")
		gof.Imports("rip_" + itf.Package)
		tryImport := func(typ string) {
			_, typ = smn_str.ProtoUseDeal(typ)
			if typ == "net.Conn" {
				return
			}
			lst := strings.Split(typ, ".")
			if len(lst) != 1 {
				gof.Imports(lst[0])
			}
		}

		{ // rpc struct
			b := gof.AddBlock("type CltRpc%s struct", itf.Name)
			b.WriteLine("%s.%s", itf.Package, itf.Name)
			b.WriteLine("conn smn_rpc.MessageAdapterItf")
			b.WriteLine("lock sync.Mutex")
			b.Imports("smn_dict")
			b.Imports("smn_rpc")
			b.Imports("sync")
		}
		{ // new func
			b := gof.AddBlock("func NewCltRpc%s(conn smn_rpc.MessageAdapterItf) *CltRpc%s", itf.Name, itf.Name)
			b.Imports("smn_rpc")
			b.WriteLine("return &CltRpc%s{conn:conn}", itf.Name)
		}
		{ // interface achieve
			for _, f := range itf.Functions {
				prmList := ""
				resList := ""
				rpcPrms := ""
				rpcRes := ""
				connFunc := ""
				haveConn := false
				for i, prm := range f.Params {
					tryImport(prm.Type)
					isConn := strings.TrimSpace(prm.Type) == "net.Conn"
					if isConn {
						haveConn = true
					}
					if i != 0 {
						prmList += ", "
						if !isConn {
							rpcPrms += ", "
						}
					}
					if !isConn {
						prmList += fmt.Sprintf("%s %s", prm.Var, prm.Type)
					} else {
						prmList += fmt.Sprintf("%s %s", prm.Var, "smn_rpc.ConnFunc")
						connFunc = prm.Var
						gof.Import("smn_rpc")
					}
					if !isConn {
						pv, usmn := goitoi64(prm.Type, prm.Var)
						rpcPrms += fmt.Sprintf("%s:%s", smn_str.InitialsUpper(prm.Var), pv)
						if usmn {
							gof.Imports("smn_rpc")
						}
					}

				}
				for i, rp := range f.Returns {
					tryImport(rp.Type)
					if i != 0 {
						resList += ", "
						rpcRes += ", "
					}
					resList += rp.Type
					pv, usmn := goi64toi(rp.Type, "_res."+smn_str.InitialsUpper(rp.Var))
					rpcRes += pv
					if usmn {
						gof.Imports("smn_rpc")
					}
				}
				b := gof.AddBlock("func (this *CltRpc%s)%s(%s) (%s)", itf.Name, f.Name, prmList, resList)
				b.WriteLine("this.lock.Lock()")
				b.WriteLine("defer this.lock.Unlock()")
				b.WriteLine("_msg := &rip_%s.%s_%s_Prm{%s}", itf.Package, itf.Name, f.Name, rpcPrms)
				b.WriteLine("this.conn.WriteCall(int32(smn_dict.EDict_rip_%s_%s_%s_Prm), _msg)", itf.Package, itf.Name, f.Name)
				if haveConn {
					b.WriteLine("%s(this.conn.GetConn())", connFunc)
				}
				b.WriteLine("_rm, _err := this.conn.ReadRet()")
				b.WriteLine("if _err != nil{\n\tpanic(_err)\n}")
				b.WriteLine("if _rm.Err{\n\tpanic(string(_rm.Msg))\n}")
				b.WriteLine("_res := &rip_%s.%s_%s_Ret{}", itf.Package, itf.Name, f.Name)
				b.WriteLine("_err = proto.Unmarshal(_rm.Msg, _res)")
				b.WriteLine("if _err != nil{\n\tpanic(_err)\n}")
				b.WriteLine("return %s", rpcRes)
			}
		}

		gof.Output()
	}
}

func main() {
	i := flag.String("i", "./src/rpc_itf/", "rpc interface dir.")
	o := flag.String("o", "./src/rpc_nitf/", "rpc interface's net accepter, from proto.Message call interface.")
	s := flag.Bool("s", true, "is product server code")
	c := flag.Bool("c", true, "is product client code")
	pkgh := flag.String("pkgh", "github.com/ProtossGenius/SureMoonNet", "package head. muti-head split with ',' ")
	flag.StringVar(&GOPATH, "gopath", "$GOPATH", "gopath")
	flag.Parse()
	harr := strings.Split(*pkgh, ",")
	for _, str := range harr {
		p := GOPATH + "/" + str
		targetPaths = append(targetPaths, p)

	}
	itfs, err := smn_rpc_itf.GetItfListFromDir(*i)
	check(err)
	for _, list := range itfs {
		if len(list) == 0 {
			continue
		}
		pkg := list[0].Package
		if *s {
			op := *o + "/svrrpc/svr_rpc_" + pkg + "/"
			err := os.MkdirAll(op, os.ModePerm)
			check(err)
			writeSvrRpcFile(op, list)
		}
		if *c {
			op := *o + "/cltrpc/clt_rpc_" + pkg + "/"
			err := os.MkdirAll(op, os.ModePerm)
			check(err)
			writeClientRpcFile(op, list)
		}
	}
}
