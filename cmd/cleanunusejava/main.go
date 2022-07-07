package main

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// JFile java file info.
type JFile struct {
	Package   string
	ClassName string
	Path      string
}

func clean(line string, ignore ...string) string {
	for _, ig := range ignore {
		line = strings.ReplaceAll(line, ig, "")
	}

	return strings.TrimSpace(line)
}

func anaysisJava(path string, lines []string) (jFile JFile, imports []string) {
	jFile = JFile{Path: path}
	imports = make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, "package") {
			jFile.Package = clean(line, "package", ";")
		}

		if strings.HasPrefix(line, "imprts") {
			imports = append(imports, clean(line, "imports", ";"))
		}

		if strings.HasPrefix(line, "class") {
			jFile.ClassName = clean(line, "class", "{")
		}
	}

	return
}

func main() {
	fmt.Println("$$$$$$$$$$$ analysis java files $$$$$$$$$")
	javas := make([]JFile, 0, 1000000)
	use := make(map[string]bool)
	smn_file.DeepTraversalDir(".", func(path string, info fs.FileInfo) smn_file.FileDoFuncResult {
		if info.IsDir() {
			return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
		}

		if !strings.HasSuffix(path, ".java") {
			return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
		}

		data, err := smn_file.FileReadAll(path)
		check(err)
		classPath, imports := anaysisJava(path, strings.Split(string(data), "\n"))
		for _, imp := range imports {
			use[imp] = true
		}
		javas = append(javas, classPath)

		return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
	})

	fmt.Println("########################### analysis java files done")
	fmt.Println("$$$$$$$$$$$  analysis to delete files")
	toDel := make([]string, 0, len(javas))

	for _, jfile := range javas {
		if use[jfile.Package+"."+jfile.ClassName] || use[jfile.Package] || use[jfile.Package+".*"] {
			continue
		}

		toDel = append(toDel, jfile.Path)
	}

	fmt.Println("##################done, ", len(toDel), "files to delete.")
	fmt.Println("$$$$$$$$$$$$$$$$$ start delete files $$$$$$$$$$$$$")
	for _, jfile := range toDel {
		fmt.Println("delete file ", jfile)
		check(smn_file.RemoveFileIfExist(jfile))
	}

	fmt.Println("########################### done!!!")
}
