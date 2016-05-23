package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/delputnam/parser"
)

var usage = `
USAGE: goat -template TemplateFileName.tpl [-in InputFilename.txt] [-format (input file format)] [-out OutputFilename.txt]

Note: The format of the input file is determined by the extension of the input filename, but can be overridden by using the '-format' option. This option is required if the input comes from Stdin.
`

type goat struct {
	templFilename string
	inFilename    string
	inFileType    string
	outFilename   string
	in            *os.File
	out           *os.File
	template      string
	input         string
	output        string
}

func main() {

	var g goat

	flag.StringVar(&g.templFilename, "template", "", "the template file to use")
	flag.StringVar(&g.inFilename, "in", "", "the input file to use, defaults to Stdin if not set")
	flag.StringVar(&g.inFileType, "format", "", "Override the input file format. (This is otherwise determined by the file extension.)")
	flag.StringVar(&g.outFilename, "out", "", "the output file to use, defaults to Stdout if not set")

	flag.Parse()

	if g.templFilename == "" {
		log.Fatal("Error: No template file was specified." + usage)
	} else {
		f, err := os.Open(g.templFilename)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		bytes, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal(err)
		}
		g.template = string(bytes)
	}

	if g.inFilename == "" {
		g.in = os.Stdin

		if g.inFileType == "" {
			log.Fatal("Error: must specify input type when data comes from Stdin.")
		}
	} else {
		var err error // ugh, because of https://github.com/golang/go/issues/6842
		g.in, err = os.Open(g.inFilename)
		if err != nil {
			log.Fatal(err)
		}
		defer g.in.Close()
		if g.inFileType == "" {
			g.inFileType = filepath.Ext(g.inFilename)
		}
	}

	bytes, err := ioutil.ReadAll(g.in)
	if err != nil {
		log.Fatal(err)
	}
	g.input = string(bytes)

	if g.outFilename == "" {
		g.out = os.Stdout
	} else {
		g.out, err = os.Open(g.outFilename)
		if err != nil {
			log.Fatal(err)
		}
		defer g.out.Close()
	}

	fmt.Printf("g: %v", g)

	// get output from parser
	p := parser.NewParser()
	data, err := p.Parse(g.inFileType, g.input)
	if err != nil {
		log.Fatal(err)
	}

	t := template.Must(template.New("goat").Parse(g.template))
	err = t.Execute(g.out, data)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
	}
}
