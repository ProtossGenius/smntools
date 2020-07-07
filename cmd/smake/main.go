package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_exec"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/smntools/smnt/codedeal"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

/*SMakeUnit target and relys..*/
type SMakeUnit struct {
	Target string
	Src    string
	Rely   []string
}

//CXXEnd C & Cpp file's end.
var CXXEnd = []string{".c", ".cpp", ".cxx", ".cc"}
var CC = "g++"
var FLAGS = "-Wall -c"

//asTarget .
func asTarget(name string) string {
	for _, end := range CXXEnd {
		if !strings.HasSuffix(name, end) {
			continue
		}

		idx := strings.LastIndex(name, end)

		return name[:idx] + ".o"
	}

	return ""
}

//getInc get include name.
func getInc(line string) string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "#") {
		return ""
	}

	line = strings.TrimSpace(line[1:])

	if !strings.HasPrefix(line, "include") {
		return ""
	}

	line = line[len("include"):]
	line = strings.ReplaceAll(line, "\"", "")
	line = strings.ReplaceAll(line, "<", "")
	line = strings.ReplaceAll(line, ">", "")

	if !strings.HasSuffix(line, ".h") {
		return ""
	}

	return line[:strings.LastIndex(line, ".h")] + ".o"
}

//SplitPath split path to dir and fileName.
func SplitPath(path string) (dir, fileName string) {
	path = strings.ReplaceAll(path, "\\", "/")
	list := strings.Split(path, "/")
	fileName = list[len(list)-1]
	dir = strings.Join(list[:len(list)-1], "/")

	return
}

//analysisRely get .cxx file's include rely.
func analysisRely(path string) *SMakeUnit {
	fmt.Println("[INFO] analysising ", path, "...")
	dir, name := SplitPath(path)
	res := &SMakeUnit{Src: name, Target: asTarget(name)}
	data, err := smn_file.FileReadAll(path)
	check(err)

	code, err := codedeal.DeleteComment(string(data))
	check(err)

	for _, line := range strings.Split(code, "\n") {
		inc := getInc(line)

		if smn_file.IsFileExist(dir + "/" + inc) {
			res.Rely = append(res.Rely, inc)
		}
	}

	return res
}

//getUserDef get user define .
func getUserDef(path string) (head, tail string) {
	var mf string
	if smn_file.IsFileExist(path + "/Makefile") {
		mf = "Makefile"
	}

	if smn_file.IsFileExist(path + "/makefile") {
		mf = "makefile"
	}

	if mf == "" {
		return "", ""
	}

	data, err := smn_file.FileReadAll(path + "/" + mf)
	check(err)

	list := strings.Split(string(data), "\n")

	for idx, line := range list {
		if strings.HasPrefix(line, "##Head") {
			head = strings.Join(list[:idx+1], "\n")
		}

		if strings.HasPrefix(line, "##Tail") {
			return head, strings.Join(list[idx:], "\n")
		}
	}

	return head, ""
}

func join(arr []string, j1, j2 string) string {
	size := len(arr)
	ress := []string{}
	maxInOneLine := 10

	for i := 0; i < size; i += maxInOneLine {
		end := i + maxInOneLine
		if end > size {
			end = size
		}

		ress = append(ress, strings.Join(arr[i:end], j1))
	}

	return strings.Join(ress, j2)
}

//WriteToMakeFile write.
func WriteToMakeFile(path string, tList []*SMakeUnit) {
	udHead, udTail := getUserDef(path)

	err := smn_file.RemoveFileIfExist(path + "/makefile")
	check(err)

	f, err := smn_file.CreateNewFile(path + "/Makefile")
	check(err)

	write := func(str string, args ...interface{}) {
		_, err := f.WriteString(fmt.Sprintf(str+"\n", args...))
		check(err)
	}

	write(udHead)

	targetList := make([]string, 0, len(tList))
	//write build one
	for _, unit := range tList {
		write(unit.Target+": %s %s", unit.Src, join(unit.Rely, " ", " \\\n"))
		write("\t%s %s %s", CC, FLAGS, unit.Src)
		targetList = append(targetList, unit.Target)
	}

	fileList, err := ioutil.ReadDir(path)
	check(err)

	//write build all
	write("sm_build_all: %s", join(targetList, " ", "\\\n"))

	for _, info := range fileList {
		if info.IsDir() && !strings.HasPrefix(info.Name(), ".") {
			write("\tcd %s && make sm_build_all", info.Name())
		}
	}
	//write clean_o
	write("sm_clean_o:\n\trm -rf ./*.o")

	for _, info := range fileList {
		if info.IsDir() && !strings.HasPrefix(info.Name(), ".") {
			write("\tcd %s && make sm_clean_o", info.Name())
		}
	}
	//write

	write(udTail)
}

func dealDir(path string) {
	tList := []*SMakeUnit{}
	fileList, err := ioutil.ReadDir(path)

	check(err)

	for _, fi := range fileList {
		if fi.IsDir() {
			continue
		}

		target := asTarget(fi.Name())
		if target == "" {
			continue
		}

		tList = append(tList, analysisRely(path+"/"+fi.Name()))
	}

	WriteToMakeFile(path, tList)
}

func main() {
	flag.StringVar(&CC, "cc", CC, "c compiler.")
	flag.StringVar(&FLAGS, "flags", FLAGS, "c++ compile flags.")
	flag.Parse()

	if smn_file.IsFileExist("./build_pre.sh") {
		check(smn_exec.EasyDirExec(".", "sh", "./build_pre.sh"))
	}

	dealDir(".")

	_, err := smn_file.DeepTraversalDir(".", func(path string, info os.FileInfo) smn_file.FileDoFuncResult {
		if !info.IsDir() {
			return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
		}

		if strings.HasPrefix(info.Name(), ".") {
			return smn_file.FILE_DO_FUNC_RESULT_NO_DEAL
		}

		dealDir(path)
		return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
	})

	check(err)
}