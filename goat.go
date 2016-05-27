package main

import (
	"errors"
	"flag"
	"fmt"
	htemplate "html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	ttemplate "text/template"

	"github.com/delputnam/parser"
)

var (
	// ErrUnknownParser is an error indicating that a parser for the requested
	// inputType is not registered.
	errUnknownOutputType = errors.New("goat: invalid output type specified, must be 'text' or 'html'")
)

type goat struct {
	templFilename string
	inFilename    string
	inFileType    string
	outFilename   string
	outFileType   string
	in            *os.File
	out           *os.File
	template      string
	input         string
	output        string
}

func main() {

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	var g goat

	flag.StringVar(&g.templFilename, "template", "", "the template file to use")
	flag.StringVar(&g.inFilename, "in", "", "the input file to use, defaults to Stdin if not set")
	flag.StringVar(&g.inFileType, "informat", "", "override the input file format. (This is otherwise determined by the file extension and is required if the input comes from Stdin.)")
	flag.StringVar(&g.outFilename, "out", "", "the output filename to use, defaults to Stdout if not set")
	flag.StringVar(&g.outFileType, "outformat", "text", "the output format (text|html), defaults to text.")
	flag.Parse()

	if g.templFilename == "" {
		log.Print("Error: No template file was specified.")
		flag.Usage()
		os.Exit(1)
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

	// open input file
	if g.inFilename == "" {
		// use Stdin if --in is not specified
		g.in = os.Stdin

		// must specify --format flag if using Stdin
		if g.inFileType == "" {
			log.Print("Error: must specify input type when data comes from Stdin.")
			flag.Usage()
			os.Exit(1)
		}
	} else {
		// open file specified by --in flag
		var err error // ugh, because of https://github.com/golang/go/issues/6842
		g.in, err = os.Open(g.inFilename)
		if err != nil {
			log.Fatal(err)
		}
		defer g.in.Close()

		// if --format flag was not specified, attempt to get format from
		// file extenstion
		if g.inFileType == "" {
			g.inFileType = filepath.Ext(g.inFilename)
			g.inFileType = strings.TrimPrefix(g.inFileType, ".")
		}
	}
	// all registered parser formats are lowercase
	g.inFileType = strings.ToLower(g.inFileType)

	// read all data from the input file
	bytes, err := ioutil.ReadAll(g.in)
	if err != nil {
		log.Fatal(err)
	}
	g.input = string(bytes)

	// get output filename
	if g.outFilename == "" {
		// user Stdout if no output file is specified
		g.out = os.Stdout
	} else {
		// create/truncate the output file
		g.out, err = os.Create(g.outFilename)
		if err != nil {
			log.Fatal(err)
		}
		defer g.out.Close()
	}

	// get output from parser
	p := parser.NewParser()
	data, err := p.Parse(g.inFileType, g.input)
	if err != nil {
		log.Fatal(err)
	}

	// execute the template
	switch g.outFileType {
	case "text":
		t := ttemplate.Must(ttemplate.New("goat").Parse(g.template))
		err = t.Execute(g.out, data)
		if err != nil {
			log.Fatal(err)
		}
		break
	case "html":
		t := htemplate.Must(htemplate.New("goat").Parse(g.template))
		err = t.Execute(g.out, data)
		if err != nil {
			log.Fatal(err)
		}
		break
	default:
		log.Fatal(errUnknownOutputType)
	}
}
