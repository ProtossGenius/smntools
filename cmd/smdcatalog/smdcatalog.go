package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

const (
	// ConstCatalog .
	ConstCatalog = "# Catalog"
	// ConstReadmdFileName = "README.md" .
	ConstReadmdFileName = "README.md"
	// ConstSubCatalog = "# SubCatalog" .
	ConstSubCatalog = "# SubCatalog"
)

func readReadme(path string) []string {
	data, err := smn_file.FileReadAll(path)
	check(err)

	lines := strings.Split(string(data), "\n")
	res := make([]string, 0, len(lines))
	haveCatalog := false

	for _, line := range lines {
		res = append(res, line)

		if line == ConstCatalog {
			haveCatalog = true

			break
		}
	}

	if !haveCatalog {
		res = append(res, ConstCatalog)
	}

	return res
}

func getTitle(fpath string) string {
	datas, err := smn_file.FileReadAll(fpath)
	check(err)

	for _, line := range strings.Split(string(datas), "\n") {
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(line[1:])
		}
	}

	return fpath
}

func asLink(name, path string) string {
	return fmt.Sprintf("* [%s](%s)", name, path)
}

func calcCatalog(path string, dirs []os.FileInfo) []string {
	const ps = smn_file.PathSep

	catalogs := make([]string, 0, len(dirs))
	subCatalogs := make([]string, 0, len(dirs))

	addLastPage := func() {
		if smn_file.IsFileExist(path + ps + ".." + ps + ConstReadmdFileName) {
			catalogs = append(catalogs, "---")
			catalogs = append(catalogs, fmt.Sprintf("[%s](%s)", "<<< upper page", "../"+ConstReadmdFileName))
			catalogs = append(catalogs, "---")
		}
	}

	addLastPage()

	for _, info := range dirs {
		if info.IsDir() {
			readmePath := path + smn_file.PathSep + info.Name() + smn_file.PathSep + ConstReadmdFileName
			if !smn_file.IsFileExist(readmePath) {
				continue
			}

			subCatalogs = append(subCatalogs,
				asLink("\\<dir>"+info.Name()+" -> "+getTitle(readmePath), "./"+info.Name()+"/"+ConstReadmdFileName))
		} else if strings.HasSuffix(info.Name(), ".md") && info.Name() != ConstReadmdFileName {
			catalogs = append(catalogs,
				asLink("\\<file>"+info.Name()+" -> "+getTitle(path+smn_file.PathSep+info.Name()), "./"+info.Name()))
		}
	}

	catalogs = append(catalogs, "", ConstSubCatalog, "")
	catalogs = append(catalogs, subCatalogs...)

	addLastPage()

	return catalogs
}

func main() {
	dir := flag.String("dir", ".", "the dir path.")
	flag.Parse()

	err := smn_file.ListDirs(*dir, func(path string) {
		readmePath := path + smn_file.PathSep + ConstReadmdFileName
		if !smn_file.IsFileExist(readmePath) {
			return
		}

		readme := readReadme(readmePath)

		dirs, err := ioutil.ReadDir(path)
		check(err)

		catalog := calcCatalog(path, dirs)
		readme = append(readme, catalog...)
		f, err := smn_file.CreateNewFile(readmePath)
		check(err)
		defer f.Close()
		_, err = f.WriteString(strings.Join(readme, "\n"))
		check(err)
	})
	check(err)
}
