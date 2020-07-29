package codedeal

import (
	"github.com/ProtossGenius/SureMoonNet/basis/smn_analysis"
	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
)

//CmdAnalysis .
func CmdAnalysis(cmd string) ([]string, error) {
	sm := lex_pgl.NewLexAnalysiser()

	go func() {
		for _, char := range cmd {
			err := sm.Read(&lex_pgl.PglaInput{Char: char})
			if err != nil {
				sm.ErrEnd(err.Error())
				break
			}
		}

		sm.End()
	}()

	rc := sm.GetResultChan()
	strArr := make([]string, 1)
	shouldAppend := false
	first := true

	for {
		lp := <-rc
		if lp.ProductType() == smn_analysis.ResultEnd {
			break
		}

		if lp.ProductType() == smn_analysis.ResultError {
			errP := lp.(*smn_analysis.ProductError)
			return nil, errP.ToError()
		}

		if lp.ProductType() < 0 {
			continue
		}

		lexP := lex_pgl.ToLexProduct(lp)

		if lexP.ProductType() == int(lex_pgl.PGLA_PRODUCT_SPACE) {
			if !first {
				shouldAppend = true
			}

			continue
		}

		first = false

		if shouldAppend {
			strArr = append(strArr, lexP.Value)
		} else {
			strArr[len(strArr)-1] += lexP.Value
		}

		shouldAppend = false
	}

	return strArr, nil
}
