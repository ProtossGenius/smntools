package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	Exts        = ".log"
	Dirs        = "."
	Interval    = 10 // mill-sec
	IgnoreError = true
	ShowPath    = false
)

var (
	FileMap = map[string]int64{}
	Lock    sync.Mutex
	ExtList = []string{}
)

func find(key string) (int64, bool) {
	Lock.Lock()
	defer Lock.Unlock()
	val, ok := FileMap[key]

	return val, ok
}

func set(key string, val int64) {
	Lock.Lock()
	defer Lock.Unlock()
	FileMap[key] = val
}

func println(a ...interface{}) {
	Lock.Lock()
	defer Lock.Unlock()
	fmt.Println(a...)
}

func print(a ...interface{}) {
	Lock.Lock()
	defer Lock.Unlock()
	fmt.Print(a...)
}

func haveError(err error) bool {
	if err == nil {
		return false
	}

	if IgnoreError {
		println("[error]", err)

		return true
	}

	panic(err)
}

func printTail(lastSize int64, path string) {
	f, err := os.Open(path)
	if haveError(err) {
		println()
	}
	defer f.Close()

	_, err = f.Seek(lastSize, io.SeekStart)
	if haveError(err) {
		println("[error] seek", path)
	}

	buffer := make([]byte, 1000)

	for {
		n, err := f.Read(buffer)

		if ShowPath {
			print(path, " --> ")
		}

		print(string(buffer[:n]))

		if errors.Is(err, io.EOF) {
			break
		}

		if haveError(err) {
			println("[error] when read ", path)

			break
		}
	}
}

func ReadTail(path string) {
	ticker := time.NewTicker(time.Duration(Interval) * time.Millisecond)

	for {
		<-ticker.C

		fi, err := os.Stat(path)

		if haveError(err) {
			println("[error] path in ", path)

			continue
		}

		size, _ := find(path)
		if size == -1 {
			set(path, fi.Size())

			continue
		}

		if size < fi.Size() {
			printTail(size, path)
			set(path, fi.Size())
		}
	}
}

func extcheck(name string) bool {
	for _, ext := range ExtList {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}

	return false
}

func AnalysisDir(path string) {
	list, err := ioutil.ReadDir(path)
	if haveError(err) {
		fmt.Println("[error] when AnalysisDir: ", path)

		return
	}

	for _, f := range list {
		if f.IsDir() {
			continue
		}

		if !extcheck(f.Name()) {
			continue
		}

		fPath := path + "/" + f.Name()
		_, exist := find(fPath)

		if exist {
			continue
		}

		set(fPath, -1)

		go ReadTail(fPath)
	}
}

func main() {
	flag.StringVar(&Exts, "exts", Exts, "exts, split with |")
	flag.StringVar(&Dirs, "dirs", Dirs, "dirs, split with |")
	flag.IntVar(&Interval, "interval", Interval, "interval mill-sec")
	flag.BoolVar(&IgnoreError, "ignoreerror", IgnoreError, "if ignore error .")
	flag.BoolVar(&ShowPath, "showpath", ShowPath, "is show path")
	flag.Parse()

	ExtList = strings.Split(Exts, "|")
	dirList := strings.Split(Dirs, "|")
	ticker := time.NewTicker(time.Duration(Interval) * time.Millisecond)

	for {
		<-ticker.C

		for _, dir := range dirList {
			AnalysisDir(dir)
		}
	}
}
