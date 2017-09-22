package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/loov/gold"
)

var (
	html        = flag.Bool("html", false, "output as html")
	printZero   = flag.Bool("zero", false, "print nodes with no-error")
	printSource = flag.Bool("source", false, "show source information")
	levelLimit  = flag.Int("max-level", -1, "level limit")
)

func main() {
	flag.Parse()
	paths := flag.Args()
	if len(paths) == 0 {
		fmt.Fprintf(os.Stderr, "No data files specified.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	tree := gold.NewTree()

	for _, path := range paths {
		label, path := extractLabelSource(path)

		data, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%q: %v\n", path, err)
			continue
		}

		output := gold.Output{}
		if err := json.Unmarshal(data, &output); err != nil {
			fmt.Fprintf(os.Stderr, "%q: %v\n", path, err)
			continue
		}
		if label != "" {
			output.Source = label
		}
		if output.Source == "" {
			output.Source = pathToSource(path)
		}
		output.AddInto(tree)
	}

	tree.UpdateError()

	renderer := gold.NewRenderer()
	renderer.LevelLimit = *levelLimit
	renderer.ShowZero = *printZero
	renderer.ShowSource = *printSource

	if *html {
		renderer.HTML(os.Stdout, tree)
	} else {
		renderer.Console(os.Stdout, tree)
	}
}

func extractLabelSource(arg string) (label, source string) {
	p := strings.IndexByte(arg, '=')
	if p < 0 {
		return "", arg
	}
	return arg[:p], arg[p+1:]
}

func pathToSource(arg string) string {
	base := filepath.Base(arg)
	return base[:len(base)-len(filepath.Ext(base))]
}
