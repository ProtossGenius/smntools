package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_err"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_flag"
	"github.com/ProtossGenius/smntools/smnt/smcfg"
)

var (
	force bool
)

var (
	onErr = smn_err.NewErrDeal()
)
var (
	cfgPath  = smcfg.GetCfgPath()
	homePath = smcfg.GetUserHome()
)

//loopRely   configName ==> is this config need to install
var loopRely = map[string]bool{}

const (
	ErrCfgPathExist      string = "Error Config directory exist."
	ErrLoopRely                 = "Error loop rely : [%s] and [%s] "
	ErrNoCheckTarget            = "Error not found check target."
	ErrNothingCanRemove         = "Error nothing can remove"
	ErrNothingCanUpdate         = "Error nothing can update"
	ErrNothingCanCollect        = "Error nothing can collect"
	ErrNotSupportSystem         = "Error not support system Windows"
)

func issue() string {
	if runtime.GOOS == "darwin" {
		return "darwin"
	}
	if runtime.GOOS == "windows" {
		panic(ErrNotSupportSystem)
	}
	if smn_file.IsFileExist("/etc/redhat-release") {
		return "centos"
	}

	data, err := smn_file.FileReadAll("/etc/issue")
	onErr.OnErr(err)

	osv := strings.ToLower(string(data))

	return strings.Split(osv, " ")[0]
}
func dirCmd(dir, e string, args ...string) error {
	cmd := exec.Command(e, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetFromGit get config from git.
func GetFromGit(target string) error {
	//init config path
	if target == "." {
		target = "https://github.com/ProtossGenius/smcfgs.git"
	}

	if smn_file.IsFileExist(cfgPath) {
		fmt.Println("config path exist : ", cfgPath)
		if !force {
			return errors.New(ErrCfgPathExist)
		}
		err := os.RemoveAll(cfgPath)
		if err != nil {
			return err
		}
	}
	err := os.MkdirAll(cfgPath, os.ModePerm)
	if err != nil {
		return err
	}
	fmt.Println("cloning path ...", target)
	//clone from git_path
	return dirCmd(homePath, "git", "clone", target, ".smcfg")
}

func findFile(basePath, shellName string) string {
	osv := issue()
	forOs := osv + "." + shellName + ".sh"
	fullPath := basePath + "/" + forOs
	if smn_file.IsFileExist(fullPath) {
		return forOs
	}

	forU := basePath + "/ubuntu." + shellName + ".sh"
	if osv != "centos" && smn_file.IsFileExist(forU) {
		return forU
	}

	return shellName + ".sh"
}

func doInstall(cfgName string) error {
	fmt.Println("installing ", cfgName, "......")
	loopRely[cfgName] = true
	defer func() { loopRely[cfgName] = false }()
	dirPath := cfgPath + cfgName
	if !smn_file.IsFileExist(dirPath) {
		return fmt.Errorf("Err no such config %s", cfgName)
	}
	//check
	err := dirCmd(dirPath, "sh", findFile(dirPath, "check"))
	if err == nil {
		return nil
	}
	//rely install
	relyList := dirPath + "/rely.list"
	if smn_file.IsFileExist(relyList) {
		data, err := smn_file.FileReadAll(relyList)
		if err != nil {
			return err
		}

		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || line[0] == '#' || line[0] == '/' {
				continue
			}

			// watch what real need to rely.
			if spls := strings.Split(line, "|"); len(spls) > 1 {
				osn := issue()
				line = spls[0]
				ignore := true
				for _, tOs := range strings.Split(spls[1], ",") {
					if tOs == osn {
						ignore = false
					}
				}

				if ignore {
					continue
				}
			}

			if !smn_file.IsFileExist(cfgPath + line) {
				return fmt.Errorf("No such config %s ", line)
			}

			if loopRely[line] {
				return fmt.Errorf(ErrLoopRely, cfgName, line)
			}
			err = doInstall(line)
			if err != nil {
				return err
			}
		}
	}

	return dirCmd(dirPath, "sh", findFile(dirPath, "install"))
}

func SmCfgInstall(target string) error {
	return doInstall(target)
}

func SmUpdate(target string) error {
	fmt.Println("updateing ", target, "......")
	err := dirCmd(cfgPath+target, "sh", "check.sh")
	if err != nil {
		return doInstall(target)
	}

	return dirCmd(cfgPath+target, "sh", findFile(cfgPath+target, "update"))
}

func SmCfgNormal(act string) func(target string) error {
	return func(target string) error {
		return dirCmd(cfgPath+target, "sh", findFile(cfgPath+target, act))
	}
}

func createShell(path string, shellName string) error {
	f, err := smn_file.CreateNewFile(path + "/" + shellName + ".sh")
	defer f.Close()

	return err
}

func SmCfgCreate(target string) error {
	path := cfgPath + target
	if smn_file.IsFileExist(path) {
		return fmt.Errorf("target exist: %s", target)
	}

	if err := os.MkdirAll(path, os.ModeDir); err != nil {
		return err
	}

	shellList := []string{"install", "check", "remove", "update", "status"}

	for _, shellName := range shellList {
		if err := createShell(path, shellName); err != nil {
			return err
		}
	}

	return nil
}

func registe(name, usage string) {
	smn_flag.RegisterString(name, usage, SmCfgNormal(name))
}

func main() {
	// check if the version more than .smcfg need.
	if smn_file.IsFileExist(smcfg.GetCfgPath() + "/less.ver") {
		datas, err := smn_file.FileReadAll(smcfg.GetCfgPath() + "/less.ver")
		if err == nil {
			lessVer := new(smn_flag.Version).FromString(string(datas))
			if Version.Less(lessVer) {
				panic(fmt.Sprintf("need new smcfg, need version: %s, current version %s", lessVer.ToString(), Version.ToString()))
			}
		}
	}

	smn_flag.RegisterVersion("smcfg", Version, "product by ProtossGenius/smntools")
	flag.BoolVar(&force, "f", force, "force excute. ")
	smn_flag.RegisterString("get",
		fmt.Sprintf("git path and install it to path[%s],   -f means delete old CfgPath ", smcfg.GetCfgPath()),
		GetFromGit)
	smn_flag.RegisterString("install", "do install", SmCfgInstall)
	smn_flag.RegisterString("update", "do update", SmUpdate)
	smn_flag.RegisterString("create", "create new config_target", SmCfgCreate)
	registe("remove", "do remvoe")
	registe("status", "show status")
	registe("check", "do check, is exist success")
	registe("collect", "collect local config to update remote.")

	flag.Parse()
	smn_flag.Parse(flag.Args(), onErr)
}
