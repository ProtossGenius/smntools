package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_exec"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

const declaration = `//the file product by build.go  ProtossGenius whose email is guyvejianglou@outlook.com
//you should never change this file.
`

// RuneList rune-array for sort.
type RuneList []rune

// Len is the number of elements in the collection.
func (r RuneList) Len() int {
	return len(r)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (r RuneList) Less(i int, j int) bool {
	return r[i] < r[j]
}

// Swap swaps the elements with indexes i and j.
func (r RuneList) Swap(i int, j int) {
	r[i], r[j] = r[j], r[i]
}

var LexCfgVarsFile *os.File

//SymbolVarCfg write symbol-var-cfg to file.
func SymbolVarCfg() {
	fmt.Println("[start]read symbol list from file and write to code")
	defer fmt.Println("[end]read symbol list from file and write to code")

	charMap := map[rune]bool{}
	datas, err := smn_file.FileReadAll("./datas/analysis/lex_pgl/symbol.cfg")
	check(err)
	writecv(`package lex_pgl

var SymbolList = map[string]bool{`)

	smbList := strings.Split(string(datas), "\n")
	for i := range smbList {
		smbList[i] = strings.TrimSpace(smbList[i])
		smbList[i] = strings.Replace(smbList[i], "\\", "\\\\", -1)
		line := smbList[i]

		if line == "" {
			continue
		}

		for _, char := range line {
			charMap[char] = true
		}

		writecvf("\"%s\":true,", line)
	}

	charList := make(RuneList, 0, len(charMap))
	for char := range charMap {
		charList = append(charList, char)
	}

	sort.Sort(charList)
	writecv(`}

var SymbolCharSet = map[rune]bool{`)

	for _, char := range charList {
		if char == '\\' {
			writecvf(`'\\':true,`)
		} else {
			writecvf("'%c':true,", char)
		}
	}

	ccMap := map[string]bool{"": true}

	writecv(`}

var SymbolCanContinue = map[string]bool{`)

	for _, c1 := range smbList {
		for _, c2 := range smbList {
			if ccMap[c2] || c1 == c2 {
				continue
			}

			if strings.HasPrefix(c1, c2) {
				writecvf("\"%s\":true, ", c2)

				ccMap[c2] = true
			}
		}
	}

	writecv(`}

//some maybe define in another type, but not as symbol. like comment's "//" and "/*"
var SymbolUnuse = map[string]bool{"//":true, "/*":true}
`)
}

//NumberVarCfg write to file.
func NumberVarCfg() {
	fmt.Println("[start]read Number Charset and write to code ")
	defer fmt.Println("[end]read Number Charset and write to code ")

	datas, err := smn_file.FileReadAll("./datas/analysis/lex_pgl/number.cfg")
	check(err)
	writecv(`
//number charSet
var NumberCharSet = map[rune]bool{`)

	for _, char := range string(datas) {
		if unicode.IsSpace(char) {
			continue
		}

		writecvf(`'%c':true, `, char)
	}

	writecv(`}
`)
}
func writecv(str string) {
	_, err := LexCfgVarsFile.WriteString(str)
	check(err)
}

func writecvf(format string, a ...interface{}) {
	writecv(fmt.Sprintf(format, a...))
}

//LexTypesCfg write to file.
func LexTypesCfg() {
	fmt.Println("[start] read lex types config and write totcode")
	defer fmt.Println("[end] read lex types config and write totcode")
	writecv(`type PglaProduct int

const (
	PGLA_PRODUCT_ PglaProduct = iota
	`)

	datas, err := smn_file.FileReadAll("./datas/analysis/lex_pgl/lextypes.cfg")
	check(err)

	constSet := map[string]bool{"": true}
	constList := []string{}
	checkList := []string{} //IsXXX(LexProdcut) bool

	for _, line := range strings.Split(string(datas), "\n") {
		line = strings.Split(line, "#")[0]
		line = strings.TrimSpace(line)
		line = strings.ToUpper(line)

		if constSet[line] {
			continue
		}

		constList = append(constList, "PGLA_PRODUCT_"+line)
		checkList = append(checkList, strings.ToUpper(line[:1])+strings.ToLower(line[1:]))
		constSet[line] = true

		writecvf("PGLA_PRODUCT_%s\n", line)
	}

	writecv(`
)
`)
	writecv(`var PglaNameMap = map[PglaProduct]string{
`)
	writecv("-1 : \"EMD\",\n")

	for _, cst := range constList {
		writecvf("%s:\"%s\",\n", cst, cst)
	}

	writecv(`}
`)

	for _, chk := range checkList {
		writecvf(`//Is%s chack if lex is %s.
func Is%s(lex *LexProduct) bool{
	return lex.ProductType() == int(PGLA_PRODUCT_%s)
}
`, chk, chk, chk, strings.ToUpper(chk))
	}
}

func main() {
	var err error
	LexCfgVarsFile, err = smn_file.CreateNewFile("./analysis/lex_pgl/cfg_vars.go")
	check(err)

	defer LexCfgVarsFile.Close()
	fmt.Println("$$$$$$$$$$$$$$$$$$$ start build project $$$$$$$$$$$$$$$$$$$$$")
	//read symbol list from file and write to code
	writecv(declaration)
	SymbolVarCfg()
	NumberVarCfg()
	LexTypesCfg()
	fmt.Println("SUCCESS")
	check(smn_exec.EasyDirExec("./", "gofmt", "-w", "./analysis/lex_pgl/cfg_vars.go"))
}
