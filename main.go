package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/urfave/negroni"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	chromahtml "github.com/alecthomas/chroma/formatters/html"
)

var (
	markdown goldmark.Markdown
	file     *string
	path     *string
	chroma   *string
	addr     *string
	outp     *string
)

var usage = `mkweb, a simple static site generator
Usage %[1]s: -file [file.md] to convert a file
      %[1]s: -path [template/] to serve a folder in dev

Options:
`

func init() {
	chroma = flag.String("chroma", "monokai", "Chroma code highlighter theme")
	file = flag.String("file", "", "CommonMark file to convert")
	path = flag.String("path", "", "Path with CommonMark files")
	addr = flag.String("addr", "localhost:3000", "HTTP service address")
	outp = flag.String("outp", "", "convert path arg instead to output folder")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	markdown = goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
			highlighting.NewHighlighting(
				highlighting.WithStyle(*chroma),
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
			),
			extension.DefinitionList,
			extension.Footnote,
			extension.GFM,
		),
		goldmark.WithExtensions(mathjax.MathJax),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
}

func renderFile(fn string, w io.Writer) {
	var buf bytes.Buffer

	source, err := ioutil.ReadFile(filepath.Join(*path, fn))
	if err != nil {
		log.Panicln(err)
	}

	context := parser.NewContext()
	if err := markdown.Convert([]byte(source), &buf, parser.WithContext(context)); err != nil {
		panic(err)
	}
	metaData := meta.Get(context)

	tmplFilename := metaData["Template"]
	if tmplFilename == nil {
		log.Panicln("Template value is nil")
	}

	tmpl, err := ioutil.ReadFile(fmt.Sprintf("%s%s", *path, tmplFilename.(string)))
	if err != nil {
		log.Panicln(fmt.Errorf("Could not read template file referenced in markdown: %v", err))
	}

	metaData["Body"] = buf.String()
	metaData["Dev"] = true

	t := template.Must(template.New("page").Parse(string(tmpl)))
	if err := t.Execute(w, metaData); err != nil {
		log.Panicln(err)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(*file) > 0 {
		renderFile(*file, os.Stdout)
		return
	} else if len(*outp) > 0 && len(*path) > 0 {
		err := filepath.Walk(*path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				if strings.HasSuffix(path, "public") {
					return filepath.SkipDir
				}
				fmt.Printf("walking %s\n", path)
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".md" {
				fmt.Printf("converting %s\n", path)
				fn := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
				f, err := os.Create(filepath.Join(*outp, fn+".html"))
				if err != nil {
					fmt.Printf("could not create output file %s: %v\n", fn, err)
					return err
				}
				renderFile(filepath.Base(path), f)
			}
			return nil
		})
		if err != nil {
			os.Exit(1)
		}
		return
	}

	go initWatcher()

	router := mux.NewRouter()

	router.HandleFunc("/ws", serveWs)
	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		renderFile("index.md", w)
	})
	router.HandleFunc("/{fn:[a-z]+.html}", func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		fn := strings.TrimSuffix(vars["fn"], filepath.Ext(vars["fn"])) + ".md"

		w.Header().Set("Cache-Control", "no-cache")
		renderFile(fn, w)
	})

	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.NewStatic(http.Dir(filepath.Join(*path, "public"))),
	)
	n.UseHandler(router)

	log.Printf("Starting mkweb in dev mode on http://%s", *addr)
	log.Fatalln(http.ListenAndServe(*addr, n))
}
