package main

import (
	"fmt"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	. "github.com/dave/jennifer/jen"
)

func main() {
	if len(os.Args) < 2 {
		panic(fmt.Errorf("not enough arguments"))
	}

	sourceType := os.Args[1]
	sourceTypePkg, sourceTypeName := splitSourceType(sourceType)

	pkg := loadPkg(sourceTypePkg)

	obj := pkg.Types.Scope().Lookup(sourceTypeName)
	if obj == nil {
		panic(fmt.Errorf("type name not found"))
	}

	if _, ok := obj.(*types.TypeName); !ok {
		panic(fmt.Errorf("%v is not a named type", obj))
	}

	structType, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		panic(fmt.Errorf("type %v is not a struct", obj))
	}

	err := generate(sourceTypeName, structType, pkg.String())
	if err != nil {
		panic(err)
	}

}

var (
	structColPattern = regexp.MustCompile(`col:"([^"]+)"`)
	tableNamePattern = regexp.MustCompile(`table_name:"([^"]+)"`)
	selectColPattern = regexp.MustCompile(`sel:"([^"]+)"`)
)


var (
	fields 			[]string
	fieldCols 		[]string
	selectCols 		[]string
	selectFields 	[]string
	cols 			[]string
	tableName 		string
)

func generate(sourceTypeName string, structType *types.Struct, pkgName string) error {

	// 1. Get the package of the file with go:generate comment
	goPackage := os.Getenv("GOPACKAGE")

	// 2. Start a new file in this package
	f := NewFile(goPackage)

	c := Qual(pkgName, "User")

	// 3. Add a package comment, so IDEs detect files as generated
	f.PackageComment("Code generated by generator, DO NOT EDIT.")


	// 4. Iterate over struct fieldCols
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		tagValue := structType.Tag(i)
		match := structColPattern.FindStringSubmatch(tagValue)
		if s := tableNamePattern.FindStringSubmatch(tagValue); s != nil {
			tableName = s[1]
			continue
		}
		if s := selectColPattern.FindStringSubmatch(tagValue); s != nil {
			selectFields = append(selectFields, fmt.Sprintf("%s", field.Name()))
			selectCols = append(selectCols, s[1])
		}
		if match == nil {
			continue
		}
		col := match[1]
		fields = append(fields, fmt.Sprintf("d.%s",field.Name()))
		fieldCols = append(fieldCols, col)
		cols = append(cols, fmt.Sprintf("$%d", len(fields)))
	}


	CreateType(f, c, sourceTypeName)
	GetTypes(f, c, sourceTypeName)

	// 6. Build the target file name
	goFile := os.Getenv("GOFILE")
	ext := filepath.Ext(goFile)
	baseFilename := goFile[0 : len(goFile)-len(ext)]
	targetFilename := baseFilename + "_" + strings.ToLower(sourceTypeName) + "_gen.go"

	// 7. Write generated file
	return f.Save(targetFilename)
}

func CreateType(f *File, c *Statement, sourceTypeName string) {
	var codes []Code
	code1 := If(Id("_, err")).
		Op(":=").
		Id(fmt.Sprintf("c.db.Exec(\"INSERT INTO %s (%s) VALUES (%s)\", %s)", tableName,
			strings.Join(fieldCols, ","), strings.Join(cols, ","), strings.Join(fields, ","))).
		Id("; err").Op("==").Id(" nil {").
		Return(Id("nil, err").
			Id("}"))

	codes = append(codes, code1)
	codes = append(codes, Return(Id("&d, nil")))


	// create func
	receiverT := "store"
	f.Func().Params(
		Id("c").Id(receiverT),
	).Id(fmt.Sprintf("Create%s", sourceTypeName)).Params(Id(fmt.Sprintf("d domain.%s", sourceTypeName))).
		Id("(*").List(c, Error()).Id(")").Block(
		codes...,
	)
}

func GetTypes(f *File,c *Statement, sourceTypeName string) {
	var codes []Code
	for i, _ := range selectFields {
		selectFields[i] = fmt.Sprintf("&user.%s", selectFields[i])
	}

	codes = append(codes,
		Var().Id("res").Id(fmt.Sprintf("[]%s", c.GoString())),

		Id("rows").Op(",").Id("err").Op(":=").Id("c.db.Exec").
			Id(fmt.Sprintf("(\"SELECT %s FROM %s\")", strings.Join(selectCols, ","), tableName)),

		If(Id("err").Op("!=").Id("nil")).Block(Return(Id("nil, err"))),

		For(Id("rows.Next")).Block(
			Var().Id("user").Id(c.GoString()),

			If(Id("err").Op(":=").Id(fmt.Sprintf("rows.Scan(%s)", strings.Join(selectFields, ",")))).
				Op(";").Id("err != nil").Block(Return(Id("nil, err"))),

			Id("res").Op("=").Append(Id("res"), Id("user")),
			),
		)

	codes = append(codes, Return(Id("&res, nil")))

	receiverT := "store"
	f.Func().Params(
		Id("c").Id(receiverT),
		).Id(fmt.Sprintf("Get%sList", sourceTypeName)).Params(Id("d").Id(c.GoString())).
		Id("(*[]").List(c, Error()).
		Id(")").Block(
			codes...,
			)
}

func loadPkg(path string) *packages.Package {
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedImports}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		panic(fmt.Errorf("loading packages for inspection: %v", err))
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	return pkgs[0]
}

func splitSourceType(sourceType string) (string, string) {
	idx := strings.LastIndexByte(sourceType, '.')
	if idx == -1 {
		panic(fmt.Errorf(`expected qualified type as "pkg/path.MyType"`))
	}
	sourceTypePackage := sourceType[0:idx]
	sourceTypeName := sourceType[idx+1:]
	return sourceTypePackage, sourceTypeName
}