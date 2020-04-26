package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
)

var tpl = `package smntac_asppl

import (
	"crypto/rsa"
	"crypto/x509"
)

func init() {
%s
}

var pubKey *rsa.PublicKey
var priKey *rsa.PrivateKey
var pkLen = %d 
var OutLen = pkLen / 8
var ReadLen = OutLen - 11

func PubKey() *rsa.PublicKey {
	return nil
}

func PriKey() *rsa.PrivateKey {
	return nil
}`

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func toByteArr(name string, data []byte) string {
	lTpl := "\t%s := []byte{%s}"
	nums := []string{}
	dLen := len(data)

	for i := 0; i < dLen; i++ {
		if i%30 == 0 {
			nums = append(nums, "\n\t\t")
		}

		nums = append(nums, fmt.Sprintf("%d, ", data[i]))
	}

	return fmt.Sprintf(lTpl, name, strings.Join(nums, ""))
}

func main() {
	var initStr string

	if smn_file.IsFileExist("./auto_code/smntac_asppl/key.go") {
		fmt.Println("file ", " ./auto_code/smntac_asppl/key.go ", "exist,  exit.")
		return
	}

	pkLen := 1280
	err := os.MkdirAll("./auto_code/smntac_asppl", os.ModePerm)

	check(err)
	f, err := smn_file.CreateNewFile("./auto_code/smntac_asppl/key.go")
	check(err)

	defer f.Close()

	priKey, err := rsa.GenerateKey(rand.Reader, 1280)
	check(err)

	priBytes := x509.MarshalPKCS1PrivateKey(priKey)
	pubBytes, err := x509.MarshalPKIXPublicKey(&priKey.PublicKey)
	check(err)

	sts := []string{"\tvar err error",
		toByteArr("pubBytes", pubBytes),
		toByteArr("priBytes", priBytes),
		"\tpubItf, err := x509.ParsePKIXPublicKey(pubBytes)",
		"\tif err != nil { panic(err) }",
		"\tpubKey = pubItf.(*rsa.PublicKey)",
		"\tpriKey, err = x509.ParsePKCS1PrivateKey(priBytes)",
		"\tif err != nil { panic(err) }",
	}
	initStr = strings.Join(sts, "\n")
	_, err = f.WriteString(fmt.Sprintf(tpl, initStr, pkLen))

	check(err)
}
