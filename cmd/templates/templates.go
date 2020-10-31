// This is a modification of https://github.com/cyberdelia/templates which
// includes directory names in template names.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

var (
	templateType = flag.String("t", "text", "Type of template (text/html)")
	source       = flag.String("s", path.Join(".", "templates"), "Location of templates")
	output       = flag.String("o", "", "Output file")
)

func main() {
	flag.Parse()

	if *templateType != "html" && *templateType != "text" {
		log.Fatalf("unexpected template type given: %s", *templateType)
	}

	buf := new(bytes.Buffer)
	fmt.Fprint(buf, fmt.Sprintf(`package templates

	import "%s/template"

  	var templates = map[string]string{`, *templateType))

	if err := filepath.Walk(*source, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if *templateType == "text" {

			// Ignore non-templates files
			if filepath.Ext(path) != ".tmpl" {
				return nil
			}
		} else {
			// Ignore non-templates files
			if filepath.Ext(path) != ".html" {
				return nil
			}
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		fmt.Fprintf(buf, "\"%s\": `%s`,\n", path, b)

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(buf, `}

	// Parse parses declared templates.
	func Parse(t *template.Template) (*template.Template, error) {
  		for name, s := range templates {
  			var tmpl *template.Template
  			if t == nil {
  				t = template.New(name)
  			}
  			if name == t.Name() {
  				tmpl = t
  			} else {
  				tmpl = t.New(name)
  			}
	  		if _, err := tmpl.Parse(s); err != nil {
  				return nil, err
  			}
  		}
  		return t, nil
  	}`)

	clean, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	file := os.Stdout
	if *output != "" {
		file, err = os.Create(*output)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Fprintln(file, string(clean))
}
