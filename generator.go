package parcel

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"io"
	"os"
	"strings"
	"time"
)

var _ Composer = &Generator{}

// GeneratorConfig controls how the code generation happens
type GeneratorConfig struct {
	// Package determines the name of the package
	Package string
	// InlcudeDocs determines whether to include documentation
	InlcudeDocs bool
}

// Generator generates an embedable resource
type Generator struct {
	// FileSystem represents the underlying file system
	FileSystem FileSystem
	// Config controls how the code generation happens
	Config *GeneratorConfig
}

// Compose generates an embedable resource for given directory
func (g *Generator) Compose(bundle *Bundle) error {
	template := &bytes.Buffer{}

	if g.Config.InlcudeDocs {
		fmt.Fprintln(template, "// Package", g.Config.Package, "contains embedded resources")
		fmt.Fprintln(template, "// Auto-generated at", time.Now().Format(time.UnixDate))
	}

	fmt.Fprintln(template, "package", g.Config.Package)
	fmt.Fprintln(template)
	fmt.Fprintf(template, "import \"github.com/phogolabs/parcel\"")
	fmt.Fprintln(template)
	fmt.Fprintln(template)
	fmt.Fprintln(template, "func init() {")
	fmt.Fprintln(template, "\tparcel.AddResource([]byte{")

	template.Write(g.prepare(bundle.Body))

	fmt.Fprintln(template, "\t})")
	fmt.Fprintln(template, "}")

	return g.write(bundle.Name, template.Bytes())
}

func (g *Generator) prepare(data []byte) []byte {
	prepared := &bytes.Buffer{}
	body := bytes.NewBuffer(data)
	reader := bufio.NewReader(body)
	buffer := &bytes.Buffer{}

	for {
		bit, rErr := reader.ReadByte()
		if rErr == io.EOF {
			line := strings.TrimSpace(buffer.String())
			fmt.Fprintln(prepared, line)
			return prepared.Bytes()
		}

		if buffer.Len() == 0 {
			fmt.Fprint(buffer, "\t\t")
		}

		fmt.Fprintf(buffer, "%d, ", int(bit))

		if buffer.Len() >= 60 {
			line := strings.TrimSpace(buffer.String())
			fmt.Fprintln(prepared, line)
			buffer.Reset()
			continue
		}
	}
}

func (g *Generator) write(name string, data []byte) error {
	var err error

	if data, err = format.Source(data); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.go", name)

	file, err := g.FileSystem.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer func() {
		if ioErr := file.Close(); err == nil {
			err = ioErr
		}
	}()

	_, err = file.Write(data)
	return err
}
