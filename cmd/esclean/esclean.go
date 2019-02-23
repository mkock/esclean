package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mkock/esclean/engine"
	"github.com/mkock/esclean/engine/loaders"
)

// Exit codes.
const (
	ExitMissArgs int = iota + 1
	ExitDirErr
	ExitFileErr
	ExitParserErr
)

func main() {
	if len(os.Args) != 2 || (!strings.HasSuffix(os.Args[1], ".js") && !strings.HasSuffix(os.Args[1], ".ts")) {
		fmt.Println("Missing: name of index.js or index.ts file")
		os.Exit(ExitMissArgs)
	}
	fix := os.Args[1]

	// Resolve fix to an absolute path.
	if !filepath.IsAbs(fix) {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Unable to determine current working directory")
			os.Exit(ExitDirErr)
		}
		fix = filepath.Join(cwd, fix)
	}

	// Check if the index file exists.
	_, err := os.Stat(fix)
	if err != nil {
		fmt.Printf("No such file: %q\n", fix)
		os.Exit(ExitFileErr)
	}

	// Parse the project and output the report results.
	fil := loaders.NewFileLoader()
	ng := engine.New(fix, fil)
	rep, err := ng.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(ExitParserErr)
	}
	fmt.Println(rep.String())
}
