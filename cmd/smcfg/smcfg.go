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
	git_path string
	force    bool
)

var (
	onErr = smn_err.NewErrDeal()
)

var (
	ErrCfgPathExist string
)

func GetFromGit(sf *smn_flag.SmnFlag, args []string) error {
	cfgPath := smcfg.GetCfgPath()
	//init config path
	if smn_file.IsFileExist(cfgPath) {
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
	cmd := exec.Command("git", "clone", git_path, ".vimrc")
	cmd.Dir = smcfg.GetUserHome()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()

}

func main() {
	smFlag := smn_flag.NewSmnFlag()
	flag.StringVar(&git_path, "get", git_path, fmt.Sprintf("git path and install it to path[%s],   -f means delete old CfgPath ", smcfg.GetCfgPath()))
	flag.BoolVar(&force, "f", force, "force excute. ")
	smFlag.RegisterString("get", &git_path, GetFromGit)
	flag.Parse()
	smFlag.Parse(flag.Args(), onErr)
}
