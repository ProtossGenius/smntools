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
	update      string
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
	ErrCfgPathExist     string = "Error Config directory exist."
	ErrLoopRely                = "Error loop rely : [%s] and [%s] "
	ErrNoCheckTarget           = "Error not found check target."
	ErrNothingCanRemove        = "Error nothing can remove"
	ErrNothingCanUpdate        = "Error nothing can update"
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
func install_sh(path string) string {
	osv := issue()
	osInstall := path + "/" + osv + ".install.sh"
	if smn_file.IsFileExist(osInstall) {
		return osInstall
	}
	return "install.sh"
}

func do_install(cfgName string) error {
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
	return dirCmd(dirPath, "sh", install_sh(dirPath))
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
	return dirCmd(cfgPath+check, "sh", "check.sh")
}

func SmCfgRemove(args []string) error {
	if remove == "" {
		return errors.New(ErrNothingCanRemove)
	}
	return dirCmd(cfgPath+remove, "sh", "remove.sh")
}

func SmCfgUpdate(args []string) error {
	if update == "" {
		return errors.New(ErrNothingCanUpdate)
	}
	return dirCmd(cfgPath+update, "sh", "update.sh")
}

func SmCfgCollect(args []string) error {
	if update == "" {
		return errors.New(ErrNothingCanUpdate)
	}
	return dirCmd(cfgPath+update, "sh", "collect.sh")
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
	smn_flag.RegisterString("update", &update, "do update", SmCfgUpdate)
	smn_flag.RegisterString("check", &check, "do check, is exist success", SmCfgCheck)
	smn_flag.RegisterString("collect", &collect, "collect local config to update remote.", SmCfgCollect)
	smn_flag.RegisterBool("pull", &pull, "update the smcfg response", SmCfgPull)
	flag.Parse()
	smn_flag.Parse(flag.Args(), onErr)
}
