package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"

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

var (
	ErrCfgPathExist     string = "Error Config directory exist."
	ErrNoCheckTarget           = "Error not found check target."
	ErrNothingCanRemove        = "Error nothing can remove"
	ErrNothingCanUpdate        = "Error nothing can update"
)

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

func SmCfgInstall(args []string) error {
	//rely install
	err := dirCmd(cfgPath+install, "sh", "rely.sh")
	if err != nil {
		return err
	}
	return dirCmd(cfgPath+install, "sh", "install.sh")
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
