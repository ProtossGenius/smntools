package codedeal

import (
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_analysis"
	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
)

func DeleteComment(src string) (string, error) {
	sm := lex_pgl.NewLexAnalysiser()

	go func() {
		for _, char := range src {
			err := sm.Read(&lex_pgl.PglaInput{Char: char})
			if err != nil {
				sm.ErrEnd(err.Error())
				break
			}
		}

		sm.End()
	}()

	rc := sm.GetResultChan()
	strArr := make([]string, 0, len(rc))

	for {
		lp := <-rc
		if lp.ProductType() == smn_analysis.ResultEnd {
			break
		}

		if lp.ProductType() == smn_analysis.ResultError {
			errP := lp.(*smn_analysis.ProductError)
			return "", errP.ToError()
		}

		if lp.ProductType() < 0 {
			continue
		}

		if lp.ProductType() != int(lex_pgl.PGLA_PRODUCT_COMMENT) {
			lexP := lex_pgl.ToLexProduct(lp)
			strArr = append(strArr, lexP.Value)
		}
	}

	return strings.Join(strArr, ""), nil
}
