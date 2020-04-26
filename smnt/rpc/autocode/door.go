package autocode

import (
	"errors"
	"fmt"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_pglang"
)

//SM just a falg for quick import package.
func SM() {}

//WriteFunc write RPC-code from interface define.
type WriteFunc func(itf *smn_pglang.ItfDef, dirPath string) error

//RPCLangMap a map save WriteFunc.
var RPCLangMap = map[string]WriteFunc{
	"cpp_s": CppSvrRPC,
}

const (
	//ErrNotSupport not support such target.
	ErrNotSupport = "ErrNotSupport"
)

/*WriteLangRPC write RPC-code from interface define.
*  target as [lang]_[c/s] such as cpp_s
 */
func WriteLangRPC(target string, itf *smn_pglang.ItfDef, dirPath string) error {
	f, exist := RPCLangMap[target]
	if !exist {
		fmt.Println("Error in autocode.WriteLangRPC, can't found target:", target)
		return errors.New(ErrNotSupport)
	}
	return f(itf, dirPath)
}
