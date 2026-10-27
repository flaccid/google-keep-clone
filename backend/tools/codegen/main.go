// Command codegen reapplies source-controlled compatibility overlays after Goa generation.
package main

import (
	_ "embed"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
)

//go:embed templates/raw_download.go.txt
var rawDownload []byte

//go:embed templates/openapi.go.txt
var openAPI []byte

func main() {
	const path = "gen/http/media/server/encode_decode.go"
	source, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, source, 0)
	if err != nil {
		panic(err)
	}
	found := false
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "EncodeDownloadResponse" {
			continue
		}
		start, end := fset.Position(fn.Pos()).Offset, fset.Position(fn.End()).Offset
		updated := append(append(append([]byte(nil), source[:start]...), rawDownload...), source[end:]...)
		updated, err = format.Source(updated)
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(path, updated, 0644); err != nil {
			panic(err)
		}
		found = true
		break
	}
	if !found {
		panic("Goa media response encoder not found; review compatibility overlay")
	}
	if err := os.WriteFile("gen/http/openapi.go", openAPI, 0644); err != nil {
		panic(err)
	}
}
