package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
)

/*
 *   command:
 *   pull push
 */
var (
	Comment string
)

const (
	PULL = "pull"
	PUSH = "push"
	SP   = "sp"
	RSP  = "rsp"
	MF   = "mf"

	CfgDir = "~/.smtools/smgit-cfg"
)

type FlagFunc func()

type FlagMap map[string]FlagFunc

func (fmap FlagMap) Register(key string, f FlagFunc) {
	fmap[key] = f
}

func (fmap FlagMap) Parse(key string) {
	if f, ok := fmap[key]; ok {
		f()
	}
}
func ec(name string, arg ...string) {
	c := exec.Command(name, arg...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	check(err)
}
func ffPull() {
	//git stash
	fmt.Println("doing stash")
	ec("git", "stash", "save", fmt.Sprintf(`'save when %s'`, time.Now().Format("2016-01-02 15:04:05")))
	//git pull
	fmt.Println("doing pull")
	ec("git", "pull")
	fmt.Println("doing pop --index")
	//git pop
	ec("git", "stash", "pop", "--index")
}

func ffPush() {
	if Comment == "" {
		panic(fmt.Errorf("no comment message"))
	}
	//make install
	fmt.Println("make install")
	ec("make", "install")
	//make test
	fmt.Println("make test")
	ec("make", "test")
	//make clean
	fmt.Println("make clean")
	ec("make", "clean")
	//git add .
	fmt.Println("git add -A")
	ec("git", "add", "-A")
	ffPull()
	//git commit -m
	fmt.Println("git commit -m ", fmt.Sprintf(`"%s"`, Comment))
	ec("git", "commit", "-m", fmt.Sprintf(`"%s"`, Comment))
	fmt.Println("git push")
	ec("git", "push")
}

//ffMf init makefile.
func ffMf() {
	if smn_file.IsFileExist("./Makefile") || smn_file.IsFileExist("./makefile") {
		return
	}

	f, err := smn_file.CreateNewFile("./Makefile")
	check(err)

	defer f.Close()

	write := func(str string) {
		_, err := f.WriteString(str)
		check(err)
	}

	write(`##Tail
debug:

qrun:

test:

install:

clean:

`)
}

func ffSp() {
	//create directory
	if !smn_file.IsFileExist(CfgDir) {
		err := os.MkdirAll(CfgDir, os.ModePerm)
		check(err)
	}
}

func ffRsp() {
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	//flagMap init.
	var flagMap = FlagMap{
		PULL: ffPull,
		PUSH: ffPush,
		SP:   ffSp,
		RSP:  ffRsp,
		MF:   ffMf,
	}

	flag.StringVar(&Comment, "m", "", "comment message for push")
	flag.Parse()

	doFlag := false
	args := flag.Args()
	fmt.Println(args)
	argMap := make(map[string]bool, len(args))

	if Comment != "" {
		argMap[PUSH] = true
	}

	for _, arg := range args {
		argMap[arg] = true
	}

	for arg := range argMap {
		flagMap.Parse(arg)

		doFlag = true
	}

	if !doFlag {
		fmt.Println(`smgit pull : pull from remote
------------equals--------------
		git stash save ''
		git pull 
		git stash pop
################################
		smgit -m [push] : push to remote, push not must
------------equals--------------
		make test
		make clean
		git add .
		git comment -m "..."
		git push
################################ wait for add
		sp  : startup pull
		rsp : remove statup pull		
$ git status`)
		ec("git", "status")
	}

	fmt.Println("################### smgit FINISH")
}
