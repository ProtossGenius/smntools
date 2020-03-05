package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"unicode"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_data"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_str"
	"github.com/ProtossGenius/smntools/smnt/smcfg"
)

var (
	SmCfgPath string
	GOPATH    string
	GOROOT    string
)

var LinkMap = map[string]string{}
var PkgToImport = map[string]bool{}
var FailList = []string

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func readCfg() {
	cfgFilePath := SmCfgPath + "/smgoget/links.map"
	if !smn_file.IsFileExist(cfgFilePath) {
		check(os.MkdirAll(SmCfgPath+"/smgoget/", os.ModePerm))
		f, err := smn_file.CreateNewFile(cfgFilePath)
		check(err)
		_, err = f.WriteString(`{"golang.org/x":"github.com/golang"}`)
		check(err)
		err = f.Close()
		check(err)
		fmt.Println(`not found config file, now create cfg file, file path is `, cfgFilePath, `; it's defaut value is `, `
		{"golang.org/x":"github.com/golang"}
		`)
	}
	cfg, err := smn_file.FileReadAll(cfgFilePath)
	check(err)
	err = smn_data.GetDataFromStr(string(cfg), &LinkMap)
	check(err)
}

func IsPkgExist(pkg string) bool {
	if pkg == "C" {
		return true
	}
	//in GOROOT
	if smn_file.IsFileExist(GOROOT + "src/" + pkg) {
		return true
	}
	//in GOPATH
	pkgList := strings.Split(pkg, "/")
	firstChar := []rune(pkg)[0]
	//if not from https
	if !unicode.IsDigit(firstChar) && !unicode.IsLetter(firstChar) && !strings.Contains(pkgList[0], ".") {
		return true
	}
	if len(pkgList) < 3 {
		return true
	}
	if len(pkgList) > 3 {
		return smn_file.IsFileExist(GOPATH + "src/" + pkg)
	}
	ownerPath := fmt.Sprintf("%ssrc/%s/%s/", GOPATH, pkgList[0], pkgList[1])
	responsePath := fmt.Sprintf("%s/%s", ownerPath, pkgList[2])
	return smn_file.IsFileExist(responsePath) && smn_file.IsFileExist(responsePath+".git")
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
			if isImport && !IsPkgExist(pkg) {
				PkgToImport[pkg] = true
			}
		}
		return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
	})
}

func dirCmd(dir, cmdStr string, args ...string) error {
	cmd := exec.Command(cmdStr, args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()

}

func doGoGet() {
	var pkg string
	if len(PkgToImport) == 0 {
		return
	}
	for pkg = range PkgToImport {
		break
	}
	pkg = strings.Replace(pkg, "\\", "/", -1)
	pkgList := strings.Split(pkg, "/")
	if len(pkgList) < 3 {
		panic(fmt.Errorf("please check if pkg error: %s", pkg))
	}
	fmt.Println("getting ", pkg, "......")
	//=== pkgList[0] = host, pkgList[1] owner, pkgList[2] response
	ownerPath := fmt.Sprintf("%ssrc/%s/%s/", GOPATH, pkgList[0], pkgList[1])
	responsePath := fmt.Sprintf("%s/%s/", ownerPath, pkgList[2])
	targetPath := fmt.Sprintf("%ssrc/%s", GOPATH, pkg)
	//get or update code.
	if smn_file.IsFileExist(responsePath) && smn_file.IsFileExist(responsePath+".git") {
		check(dirCmd(responsePath, "git", "pull"))
	} else {
		if !smn_file.IsFileExist(ownerPath) {
			check(os.MkdirAll(ownerPath, os.ModePerm))
		}
		if smn_file.IsFileExist(responsePath) {
			os.Remove(responsePath)
		}
		hPath := fmt.Sprintf(`https://%s/%s/%s.git`, pkgList[0], pkgList[1], pkgList[2])
		searchPath := fmt.Sprintf("%s/%s", pkgList[0], pkgList[1])
		if rep, ok := LinkMap[searchPath]; ok {
			hPath = strings.Replace(hPath, searchPath, rep, 1)
		}
		fmt.Println("useing http path:", hPath)
		check(dirCmd(ownerPath, "git", "clone", hPath))
	}
	delete(PkgToImport, pkg)
	//analysis project and import
	analysis(responsePath)
	//install target.
	fmt.Println("try install target.... target path is :", targetPath)
	err := dirCmd(targetPath, "go", "install") //because can't import is not a big thing. so don't panic error
	if err != nil {
		fmt.Println("when try install target path error happened. target path is :", targetPath, " error is ", err)
	}
	doGoGet()
}

func dealGoPath() {
	GOPATH = os.Getenv("GOPATH")
	GOROOT = os.Getenv("GOROOT")

	if GOPATH == "" || GOROOT == "" {
		panic(fmt.Errorf("please install go or confg GOPATH(%s)  GOROOT(%s) ", GOPATH, GOROOT))
	}
	//gopath clean, gopath maybe have lots of dir
	switch runtime.GOOS {
	case "windows":
		GOPATH = strings.TrimSpace(strings.Split(GOPATH, ";")[0])
		//let windows's '\' seq to '/'
		GOPATH = smn_str.PathFmt(GOPATH)
		//another split type..
	default:
		GOPATH = strings.TrimSpace(strings.Split(GOPATH, ":")[0])
	}
	GOROOT = smn_str.PathFmt(GOROOT)
	if !strings.HasSuffix(GOPATH, "/") {
		GOPATH += "/"
	}
	if !strings.HasSuffix(GOROOT, "/") {
		GOROOT += "/"
	}
}

func main() {
	var err error
	SmCfgPath, err = smcfg.GetCfgPath()
	check(err)
	flag.StringVar(&SmCfgPath, "sp", SmCfgPath, `smcfg path, config file's path is "*sp/smgoget/links.map" `)
	flag.Parse()
	readCfg()
	dealGoPath()
	args := flag.Args()
	PkgToImport = make(map[string]bool, len(args))
	for _, arg := range args {
		PkgToImport[arg] = true
	}
	doGoGet()
}
