package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_err"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_flag"
	"github.com/ProtossGenius/smntools/smnt/smcfg"
)

var (
	git_path    string
	force       bool
	install     string
	remove      string
	check       string
	collect     string
	pull        bool
	update_all  bool
	install_all bool
	check_all   bool
	remove_all  bool
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

var (
	ErrCfgPathExist      string = "Error Config directory exist."
	ErrLoopRely                 = "Error loop rely : [%s] and [%s] "
	ErrNoCheckTarget            = "Error not found check target."
	ErrNothingCanRemove         = "Error nothing can remove"
	ErrNothingCanUpdate         = "Error nothing can update"
	ErrNothingCanCollect        = "Error nothing can collect"
)

func issue() string {
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

func GetFromGit(args []string) error {
	//init config path
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
	fmt.Println("cloning path ...", git_path)
	//clone from git_path
	return dirCmd(homePath, "git", "clone", git_path, ".smcfg")
}

func findFile(basePath, shellName string) string {
	osv := issue()
	forOs := osv + "." + shellName + ".sh"
	fullPath := basePath + "/" + forOs
	fmt.Println(fullPath)
	if smn_file.IsFileExist(fullPath) {
		return forOs
	}

	return shellName + ".sh"
}

func do_install(cfgName string) error {
	fmt.Println("installing ", cfgName, "......")
	loopRely[cfgName] = true
	defer func() { loopRely[cfgName] = false }()
	dirPath := cfgPath + cfgName
	if !smn_file.IsFileExist(dirPath) {
		return fmt.Errorf("Err no such config %s", cfgName)
	}
	//check
	err := dirCmd(dirPath, "sh", "check.sh")
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

			if !smn_file.IsFileExist(cfgPath + line) {
				return fmt.Errorf("No such config %s ", line)
			}

			if loopRely[line] {
				return fmt.Errorf(ErrLoopRely, cfgName, line)
			}
			err = do_install(line)
			if err != nil {
				return err
			}
		}
	}

	return dirCmd(dirPath, "sh", findFile(dirPath, "install"))
}

func SmCfgInstall(args []string) error {
	return do_install(install)
}

func SmCfgCheck(args []string) error {
	if check == "" {
		if install == "" {
			return errors.New(ErrNoCheckTarget)
		}
		check = install
	}

	return dirCmd(cfgPath+check, "sh", findFile(cfgPath+check, "check"))
}

func SmCfgRemove(args []string) error {
	if remove == "" {
		return errors.New(ErrNothingCanRemove)
	}

	return dirCmd(cfgPath+remove, "sh", findFile(cfgPath+remove, "remove"))
}

func SmCfgCollect(args []string) error {
	if collect == "" {
		return errors.New(ErrNothingCanCollect)
	}

	return dirCmd(cfgPath+collect, "sh", findFile(cfgPath+collect, "collect"))
}

func SmCfgPull(args []string) error {
	return dirCmd(cfgPath, "git", "pull")
}
func main() {
	flag.BoolVar(&force, "f", force, "force excute. ")
	smn_flag.RegisterString("get", &git_path,
		fmt.Sprintf("git path and install it to path[%s],   -f means delete old CfgPath ", smcfg.GetCfgPath()),
		GetFromGit)
	smn_flag.RegisterString("install", &install, "do install", SmCfgInstall)
	smn_flag.RegisterString("remove", &remove, "do remvoe", SmCfgRemove)
	smn_flag.RegisterString("update", &install, "do update", SmCfgInstall)
	smn_flag.RegisterString("check", &check, "do check, is exist success", SmCfgCheck)
	smn_flag.RegisterString("collect", &collect, "collect local config to update remote.", SmCfgCollect)
	smn_flag.RegisterBool("pull", &pull, "update the smcfg response", SmCfgPull)
	flag.Parse()
	smn_flag.Parse(flag.Args(), onErr)
}
