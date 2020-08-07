package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_exec"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/smntools/smnt/codedeal"
	"github.com/ProtossGenius/smntools/smnt/smcfg"
	jsoniter "github.com/json-iterator/go"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

//SMakeLink .
type SMakeLink struct {
	Type      string   `json:"type"`
	Path      string   `json:"path"`
	Name      string   `json:"name"`
	CmdTarget []string `json:"cmd_target"`
	CmdLocal  []string `json:"cmd_local"`
}

/*SMakeUnit target and relys..*/
type SMakeUnit struct {
	Target string
	Src    string
	Rely   sort.StringSlice
}

//CXXEnd C & Cpp file's end.
var CXXEnd = []string{".c", ".cpp", ".cxx", ".cc"}
var CC = "g++"
var FLAGS = "-Wall -c"
var SLS = "sls.json" // symbol links config file's name
var wg sync.WaitGroup
var mutex sync.Mutex
var DealedPath = map[string]bool{}

var (
	SMakeCfgDir     = smcfg.GetUserHome() + "/.smake/"
	SMakeCfgGitDir  = SMakeCfgDir + "git/"
	SMakeCfgWgetDir = SMakeCfgDir + "wget/"
)

func isNeedDown(path string) bool {
	mutex.Lock()
	defer mutex.Unlock()

	if DealedPath[path] {
		return false
	}

	DealedPath[path] = true

	return true
}

func letPathShort(path string) string {
	for strings.Contains(path, "//") || strings.Contains(path, "/./") {
		path = strings.ReplaceAll(path, "//", "/")
		path = strings.ReplaceAll(path, "/./", "/")
	}

	return path
}

func println(a ...interface{}) {
	mutex.Lock()
	defer mutex.Unlock()

	fmt.Println(a...)
}

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

	return strings.TrimSpace(line)
}

//SplitPath split path to dir and fileName.
func SplitPath(path string) (dir, fileName string) {
	path = strings.ReplaceAll(path, "\\", "/")
	list := strings.Split(path, "/")
	fileName = list[len(list)-1]
	dir = strings.Join(list[:len(list)-1], "/")

	return
}

//getHeaderRely header may include header too .
func getHeaderRely(path string, remMap map[string]bool) []string {
	relys := []string{}
	dir, _ := SplitPath(path)
	data, err := smn_file.FileReadAll(path)
	check(err)

	code, err := codedeal.DeleteComment(string(data))
	check(err)

	for _, line := range strings.Split(code, "\n") {
		inc := getInc(line)
		if inc == "" {
			continue
		}

		incPath := dir + "/" + inc
		if remMap[incPath] {
			continue
		}

		if smn_file.IsFileExist(incPath) {
			remMap[incPath] = true

			relys = append(relys, inc)
			subIncs := getHeaderRely(incPath, remMap)
			subDir, _ := SplitPath(incPath)
			prex := "./" + subDir[len(dir):] + "/"

			if subDir[len(dir):] == "/" || subDir == dir {
				prex = "./"
			}

			for idx := range subIncs {
				subIncs[idx] = letPathShort(prex + subIncs[idx])
			}

			relys = append(relys, subIncs...)
		}
	}

	return relys
}

//analysisRely get .cxx file's include rely.
func analysisRely(path string) *SMakeUnit {
	println("[INFO] analysising ", path, "...")
	_, name := SplitPath(path)
	res := &SMakeUnit{Src: name, Target: asTarget(name)}
	remMap := map[string]bool{}
	res.Rely = getHeaderRely(path, remMap)

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
			head = strings.Join(list[:idx+1], "\n") + "\n"
		}

		if strings.HasPrefix(line, "##Tail") {
			return head, "\n" + strings.Join(list[idx:], "\n")
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

func dropLastBL(str string) string {
	size := len(str)

	for i := size - 1; i >= 0; i-- {
		if str[i] != '\n' {
			return str[:i+1] + "\n"
		}
	}

	return str
}

//WriteToMakeFile write.
func WriteToMakeFile(path string, tList []*SMakeUnit) {
	udHead, udTail := getUserDef(path)
	udTail = dropLastBL(udTail)

	err := smn_file.RemoveFileIfExist(path + "/makefile")
	check(err)

	f, err := smn_file.CreateNewFile(path + "/Makefile")
	check(err)

	write := func(str string, args ...interface{}) {
		_, err := f.WriteString(fmt.Sprintf(str+"\n", args...))
		check(err)
	}

	write(udHead)

	targetList := make(sort.StringSlice, 0, len(tList))
	//write build one
	for _, unit := range tList {
		sort.Sort(unit.Rely)
		write(unit.Target+": %s %s", unit.Src, join(unit.Rely, " ", " \\\n"))
		write("\t%s %s %s", CC, FLAGS, unit.Src)
		targetList = append(targetList, unit.Target)
	}

	fileList, err := ioutil.ReadDir(path)
	check(err)

	cleanSubs := make(sort.StringSlice, 0, len(fileList))
	buildAllSubs := make(sort.StringSlice, 0, len(fileList))
	buildSubs := make(sort.StringSlice, 0, len(fileList))
	dirIdx := 0
	//sub dir.
	for _, info := range fileList {
		if strings.HasPrefix(info.Name(), ".") {
			continue
		}

		if info.IsDir() || smn_file.IsFileExist(path+"/"+info.Name()+"/Makefile") {
			//subdir's build
			buildAllSubs = append(buildAllSubs, fmt.Sprintf("\t+make -C %s sm_build_all", info.Name()))

			if info.IsDir() {
				buildSubs = append(buildSubs, fmt.Sprintf("\t+make -C %s sm_build", info.Name()))
				//subdir's clean.
				cleanSubs = append(cleanSubs, fmt.Sprintf("\t+make -C %s sm_clean_o", info.Name()))
			}
			dirIdx++
		}
	}

	sort.Sort(targetList)
	sort.Sort(cleanSubs)
	sort.Sort(buildAllSubs)
	sort.Sort(buildSubs)
	//write build
	write("sm_build: %s", join(targetList, " ", "\\\n"))
	write(strings.Join(buildSubs, "\n"))
	write("")
	//write build all
	write("sm_build_all: %s", join(targetList, " ", "\\\n"))
	write(strings.Join(buildAllSubs, "\n"))
	write("")
	//write clean_o
	write("sm_clean_o:\n\trm -rf ./*.o")
	write(strings.Join(cleanSubs, "\n"))
	//write Tail
	write(udTail)
}

//MakeSLink .
func MakeSLink(path string, cfg *SMakeLink) {
	defer wg.Done()

	var err error

	hasError := func(action string) bool {
		if err == nil {
			return false
		}

		println("[ERROR] in MakeSLink, path is [", path, "],  error is [", err, "] reason is :", action)

		return true
	}

	if cfg.Name == "" {
		_, cfg.Name = SplitPath(cfg.Path)
	}

	switch strings.ToLower(cfg.Type) {
	case "local", "l":
		{
			println("[INFO] make local symlink", cfg.Path, " ", path+"/"+cfg.Name)
			_ = os.Remove(path + "/" + cfg.Name)
			err = os.Symlink(cfg.Path, path+"/"+cfg.Name)
			if hasError("create Symlink") {
				return
			}
		}
	case "git", "g":
		{
			println("[INFO] make symlink", SMakeCfgGitDir+cfg.Path, path+"/"+cfg.Name)
			_ = os.Remove(path + "/" + cfg.Name)
			err = os.Symlink(SMakeCfgGitDir+cfg.Path, path+"/"+cfg.Name)

			if !isNeedDown(cfg.Path) {
				break
			}
			const gitDownSize = 3
			//split git-path   github.com/user/r-name/subdir1/s2/...
			list := strings.Split(cfg.Path, "/")
			if len(list) < gitDownSize {
				println("[ERROR] config path error: ", cfg)
				return
			}
			localPath := SMakeCfgGitDir + strings.Join(list[:3], "/")
			println("downloading ", cfg.Path, "...")
			if !smn_file.IsFileExist(localPath) {
				err = smn_exec.EasyDirExec(path, "git", "clone", "https://"+strings.Join(list[:3], "/"), localPath)
			} else {
				err = smn_exec.EasyDirExec(localPath, "git", "checkout", ".")
				if hasError("git checkout .") {
					return
				}

				err = smn_exec.EasyDirExec(localPath, "git", "clean", "-xfd")
				if hasError("git clean -xfd") {
					return
				}

				err = smn_exec.EasyDirExec(localPath, "git", "pull")
			}

			println("download", cfg.Path, "finish.")
			if hasError("git") {
				return
			}

			cfg.Path = SMakeCfgGitDir + cfg.Path
		}

	default:
		println("[ERROR] unkonw type: ", cfg.Type)
	}

	for _, cmd := range cfg.CmdTarget {
		err = runCmd(cmd, cfg.Path)

		if hasError("run target cmd") {
			return
		}
	}

	for _, cmd := range cfg.CmdLocal {
		err = runCmd(cmd, path)

		if hasError("run local cmd") {
			return
		}
	}
}

func runCmd(cmd string, path string) (err error) {
	if cmd == "" {
		return
	}

	var arr []string
	arr, err = codedeal.CmdAnalysis(cmd)

	if err != nil {
		return err
	}

	if len(arr) == 0 {
		return nil
	}

	return smn_exec.EasyDirExec(path, arr[0], arr[1:]...)
}

//ThirdPartDir import 3rd_part as SymbolLinks.
func ThirdPartDir(path string) {
	defer wg.Done()

	data, err := smn_file.FileReadAll(path + "/" + SLS)
	check(err)

	cfgs := &[]*SMakeLink{}
	err = jsoniter.Unmarshal(data, cfgs)
	check(err)

	for _, cfg := range *cfgs {
		wg.Add(1)

		go MakeSLink(path, cfg)
	}
}

//InitSMakeCfgDir if dir not exist create it.
func InitSMakeCfgDir() {
	checkDir := func(path string) {
		if !smn_file.IsFileExist(path) {
			check(os.MkdirAll(path, os.ModePerm))
		}
	}

	checkDir(SMakeCfgDir)
	checkDir(SMakeCfgGitDir)
	checkDir(SMakeCfgWgetDir)
}

func dealDir(path string) {
	defer wg.Done()

	if smn_file.IsFileExist(path + "/" + SLS) {
		wg.Add(1)

		go ThirdPartDir(path)
	}

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
	InitSMakeCfgDir()

	var dirAction = dealDir

	wg.Add(1)

	go dirAction(".")

	_, err := smn_file.DeepTraversalDir(".", func(path string, info os.FileInfo) smn_file.FileDoFuncResult {
		if !info.IsDir() {
			return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
		}

		if strings.HasPrefix(info.Name(), ".") {
			return smn_file.FILE_DO_FUNC_RESULT_NO_DEAL
		}
		wg.Add(1)
		go dirAction(path)
		return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
	})

	check(err)

	wg.Wait()
}
