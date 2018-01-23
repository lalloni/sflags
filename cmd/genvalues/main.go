// This generator is for internal usage only.
//
// It generates values described in values.json.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"reflect"
	"text/template"
	"unicode"
	"unicode/utf8"
)

const (
	tmpl = `package sflags

// This file is autogenerated by "go generate .". Do not modify.

import (
{{range .Imports}}
"{{.}}"{{end}}
)

{{$mapKeyTypes := .MapKeysTypes}}

// MapAllowedKinds stores list of kinds allowed for map keys.
var MapAllowedKinds = []reflect.Kind{ \nn
{{range $mapKeyTypes}}
	reflect.{{. | Title}},{{end}}
}

func parseGenerated(value interface{}) Value {
	switch value.(type) {
	{{range .Values}}{{ if eq (.|InterfereType) (.Type) }}\nn
	case *{{.Type}}:
		return new{{.|Name}}Value(value.(*{{.Type}}))
	{{ end }}{{ end }}\nn
	{{range .Values}}{{ if not .NoSlice }}\nn
	case *[]{{.Type}}:
		return new{{.|Plural}}Value(value.(*[]{{.Type}}))
	{{end}}{{end}}\nn
	default:
		return nil
	}
}

func parseGeneratedPtrs(value interface{}) Value {
	switch value.(type) {
	{{range .Values}}{{ if ne (.|InterfereType) (.Type) }}\nn
	case *{{.Type}}:
		return new{{.|Name}}Value(value.(*{{.Type}}))
	{{end}}{{end}}\nn
	default:
		return nil
	}
}

func parseGeneratedMap(value interface{}) Value {
	switch value.(type) {
	{{range .Values}}{{ if not .NoMap }}\nn
	{{ $value := . }}{{range $mapKeyTypes}}\nn
	case *map[{{.}}]{{$value.Type}}:
		return new{{MapValueName $value . | Title}}(value.(*map[{{.}}]{{$value.Type}}))
	{{end}}{{end}}{{end}}\nn
	default:
		return nil
	}
}

{{range .Values}}
{{if not .NoValueParser}}
// -- {{.Type}} Value
type {{.|ValueName}} struct {
	value *{{.Type}}
}

var _ Value = (*{{.|ValueName}})(nil)
var _ Getter = (*{{.|ValueName}})(nil)

func new{{.|Name}}Value(p *{{.Type}}) *{{.|ValueName}} {
	return &{{.|ValueName}}{value: p}
}

func (v *{{.|ValueName}}) Set(s string) error {
	{{if .Parser }}\nn
	parsed, err := {{.Parser}}
	if err == nil {
		{{if .Convert}}\nn
		*v.value = ({{.Type}})(parsed)
		{{else}}\nn
		*v.value = parsed
		{{end}}\nn
		return nil
	}
	return err
	{{ else }}\nn
	*v.value = s
	return nil
	{{end}}\nn
}

func (v *{{.|ValueName}}) Get() interface{} {
 	if v != nil && v.value != nil {
{{/* flag package create zero Value and compares it to actual Value */}}\nn
 		return *v.value
 	}
	return nil
}

func (v *{{.|ValueName}}) String() string {
	if v != nil && v.value != nil {
{{/* flag package create zero Value and compares it to actual Value */}}\nn
		return {{.|Format}}
	}
	return ""
}

func (v *{{.|ValueName}}) Type() string { return "{{.|Type}}" }

{{ if not .NoSlice }}
// -- {{.Type}}Slice Value

type {{.|SliceValueName}} struct{
	value   *[]{{.Type}}
	changed bool
}

var _ RepeatableFlag = (*{{.|SliceValueName}})(nil)
var _ Value = (*{{.|SliceValueName}})(nil)
var _ Getter = (*{{.|SliceValueName}})(nil)


func new{{.|Name}}SliceValue(slice *[]{{.Type}}) *{{.|SliceValueName}}  {
	return &{{.|SliceValueName}}{
		value: slice,
	}
}

func (v *{{.|SliceValueName}}) Set(raw string) error {
	ss := strings.Split(raw, ",")
	{{if .Parser }}
	out := make([]{{.Type}}, len(ss))
	for i, s := range ss {
		parsed, err := {{.Parser}}
		if err != nil {
			return err
		}
		{{if .Convert}}\nn
		out[i] = ({{.Type}})(parsed)
		{{else}}\nn
		out[i] = parsed
		{{end}}\nn
	}
	{{ else }}out := ss{{end}}
	if !v.changed {
		*v.value = out
	} else {
		*v.value = append(*v.value, out...)
	}
	v.changed = true
	return nil
}

func (v *{{.|SliceValueName}}) Get() interface{} {
 	if v != nil && v.value != nil {
{{/* flag package create zero Value and compares it to actual Value */}}\nn
 		return *v.value
 	}
	return ([]{{.Type}})(nil)
}

func (v *{{.|SliceValueName}}) String() string {
	if v == nil || v.value == nil {
{{/* flag package create zero Value and compares it to actual Value */}}\nn
		return "[]"
	}
	out := make([]string, 0, len(*v.value))
	for _, elem := range *v.value {
		out = append(out, new{{.|Name}}Value(&elem).String())
	}
	return "[" + strings.Join(out, ",") + "]"
}

func (v *{{.|SliceValueName}}) Type() string { return "{{.|Type}}Slice" }

func (v *{{.|SliceValueName}}) IsCumulative() bool {
	return true
}

{{end}}

{{ if not .NoMap }}
{{ $value := . }}
{{range $mapKeyTypes}}
// -- {{ MapValueName $value . }}
type {{ MapValueName $value . }} struct {
	value *map[{{.}}]{{$value.Type}}
}

var _ RepeatableFlag = (*{{MapValueName $value .}})(nil)
var _ Value = (*{{MapValueName $value .}})(nil)
var _ Getter = (*{{MapValueName $value .}})(nil)


func new{{MapValueName $value . | Title}}(m *map[{{.}}]{{$value.Type}}) *{{MapValueName $value .}}  {
	return &{{MapValueName $value .}}{
		value: m,
	}
}

func (v *{{MapValueName $value .}}) Set(s string) error {
	ss := strings.Split(s, ":")
    if len(ss) < 2 {
        return errors.New("invalid map flag syntax, use -map=key1:val1")
    }

	{{ $kindVal := KindValue . }}

	s = ss[0]

	{{if $kindVal.Parser }}\nn
	parsedKey, err := {{$kindVal.Parser}}
	if err != nil {
        return err
	}

	{{if $kindVal.Convert}}\nn
	key := ({{$kindVal.Type}})(parsedKey)
	{{else}}\nn
	key := parsedKey
	{{end}}\nn

	{{ else }}\nn
	key := s
	{{end}}\nn


	s = ss[1]
 
	{{if $value.Parser }}\nn
	parsedVal, err := {{$value.Parser}}
	if err != nil {
        return err
	}

	{{if $value.Convert}}\nn
	val := ({{$value.Type}})(parsedVal)
	{{else}}\nn
	val := parsedVal
	{{end}}\nn

	{{ else }}\nn
	val := s
	{{end}}\nn

	(*v.value)[key] = val

	return nil
}

func (v *{{MapValueName $value .}}) Get() interface{} {
 	if v != nil && v.value != nil {
{{/* flag package create zero Value and compares it to actual Value */}}\nn
 		return *v.value
 	}
	return nil
}

func (v *{{MapValueName $value .}}) String() string {
	if v != nil && v.value != nil && len(*v.value) > 0 {
{{/* flag package create zero Value and compares it to actual Value */}}\nn
		return fmt.Sprintf("%v", *v.value)
	}
	return ""
}

func (v *{{MapValueName $value .}}) Type() string { return "map[{{.}}]{{$value.Type}}" }

func (v *{{MapValueName $value .}}) IsCumulative() bool {
	return true
}
{{end}}
{{end}}

{{end}}


{{end}}
`
	testTmpl = `package sflags

// This file is autogenerated by "go generate .". Do not modify.

import (
	"github.com/stretchr/testify/assert"
	"testing"
{{range .Imports}}\nn
"{{.}}"
{{end}}\nn
)

{{$mapKeyTypes := .MapKeysTypes}}

{{range .Values}}

func Test{{.|Name}}Value_Zero(t *testing.T) {
	nilValue := new({{.|ValueName}})
	assert.Equal(t, "", nilValue.String())
	assert.Nil(t, nilValue.Get())
	nilObj := (*{{.|ValueName}})(nil)
	assert.Equal(t, "", nilObj.String())
	assert.Nil(t, nilObj.Get())
}


{{ if .Tests }}{{ $value := . }}
func Test{{.|Name}}Value(t *testing.T) {
	{{range .Tests}}\nn
		t.Run("{{.}}", func(t *testing.T){
		{{ if ne ($value|InterfereType) ($value.Type) }}\nn
		a := new({{$value|InterfereType}})
		v := new{{$value|Name}}Value(&a)
		assert.Equal(t, parseGeneratedPtrs(&a), v)
		{{ else }}\nn
		a := new({{$value.Type}})
		v := new{{$value|Name}}Value(a)
		assert.Equal(t, parseGenerated(a), v)
		{{ end }}\nn
		err := v.Set("{{.In}}")
		{{if .Err}}\nn
		assert.EqualError(t, err, "{{.Err}}")
		{{ else }}\nn
		assert.Nil(t, err)
		{{end}}\nn
		assert.Equal(t, "{{.Out}}", v.String())
		{{ if ne ($value|InterfereType) ($value.Type) }}\nn
		assert.Equal(t, a, v.Get())
		{{else}}\nn
		assert.Equal(t, *a, v.Get())
		{{end}}\nn
		assert.Equal(t, "{{$value|Type}}", v.Type())
	})
	{{end}}
}{{end}}


{{ if not .NoSlice }}
func Test{{.|Name}}SliceValue_Zero(t *testing.T) {
	nilValue := new({{.|SliceValueName}})
	assert.Equal(t, "[]", nilValue.String())
	assert.Nil(t, nilValue.Get())
	nilObj := (*{{.|SliceValueName}})(nil)
	assert.Equal(t, "[]", nilObj.String())
	assert.Nil(t, nilObj.Get())
}{{end}}


{{ if not .NoMap }}
{{ $value := . }}
{{range $mapKeyTypes}}
func Test{{MapValueName $value . | Title}}_Zero(t *testing.T) {
	nilValue := new({{MapValueName $value .}})
	assert.Equal(t, "", nilValue.String())
	assert.Nil(t, nilValue.Get())
	nilObj := (*{{MapValueName $value . }})(nil)
	assert.Equal(t, "", nilObj.String())
	assert.Nil(t, nilObj.Get())
}
{{end}}
{{end}}


{{ if .SliceTests }}{{ $value := . }}
func Test{{.|Name}}SliceValue(t *testing.T) {
	{{range .SliceTests}}{{ $test := . }}\nn
	t.Run("{{.}}", func(t *testing.T){
		var err error
		a := new([]{{$value.Type}})
		v := new{{$value|Name}}SliceValue(a)
		assert.Equal(t, parseGenerated(a), v)
		assert.True(t, v.IsCumulative())
		{{range .In}}\nn
		err = v.Set("{{.}}")
		{{if $test.Err}}\nn
		assert.EqualError(t, err, "{{$test.Err}}")
		{{ else }}\nn
		assert.Nil(t, err)
		{{end}}\nn
		{{end}}\nn
		assert.Equal(t, "{{.Out}}", v.String())
		assert.Equal(t, *a, v.Get())
		assert.Equal(t, "{{$value|Type}}Slice", v.Type())
	})
	{{end}}
}{{end}}


{{end}}


	`
)

// MapAllowedKinds stores list of kinds allowed for map keys.
var mapAllowedKinds = []reflect.Kind{
	reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
	reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
}

type test struct {
	In  string
	Out string
	Err string
}

func (t *test) String() string {
	return fmt.Sprintf("in: %s", t.In)
}

type sliceTest struct {
	In  []string
	Out string
	Err string
}

func (t *sliceTest) String() string {
	return fmt.Sprintf("in: %v", t.In)
}

type value struct {
	Name          string      `json:"name"`
	Kind          string      `json:"kind"`
	NoValueParser bool        `json:"no_value_parser"`
	Convert       bool        `json:"convert"`
	Type          string      `json:"type"`
	Parser        string      `json:"parser"`
	Format        string      `json:"format"`
	Plural        string      `json:"plural"`
	Help          string      `json:"help"`
	Import        []string    `json:"import"`
	Tests         []test      `json:"tests"`
	SliceTests    []sliceTest `json:"slice_tests"`
	NoSlice       bool        `json:"no_slice"`
	NoMap         bool        `json:"no_map"`
}

func fatalIfError(err error) {
	if err != nil {
		panic(err)
	}
}

// removeNon removes \nn\n from string
func removeNon(src string) string {
	return strings.Replace(src, "\\nn\n", "", -1)
}

func main() {
	r, err := os.Open("values.json")
	fatalIfError(err)
	defer r.Close()

	values := []value{}
	err = json.NewDecoder(r).Decode(&values)
	fatalIfError(err)

	valueName := func(v *value) string {
		if v.Name != "" {
			return strings.Title(v.Name)
		}
		return strings.Title(v.Type)
	}
	imports := []string{}
	for _, value := range values {
		imports = append(imports, value.Import...)
	}

	baseT := template.New("genvalues").Funcs(template.FuncMap{
		"Lower": strings.ToLower,
		"Title": strings.Title,
		"Format": func(v *value) string {
			if v.Format != "" {
				return v.Format
			}
			return "fmt.Sprintf(\"%v\", *v.value)"
		},
		"ValueName": func(v *value) string {
			if v.Name == v.Type {
				return v.Type // that's package type
			}
			name := valueName(v)
			return camelToLower(name) + "Value"
		},
		"SliceValueName": func(v *value) string {
			name := valueName(v)
			return camelToLower(name) + "SliceValue"
		},
		"MapValueName": func(v *value, kind string) string {
			name := valueName(v)

			return kind + name + "MapValue"
		},
		"KindValue": func(kind string) value {
			for _, value := range values {
				if value.Type == kind {
					return value
				}
			}

			return value{}
		},
		"Name": valueName,
		"Plural": func(v *value) string {
			if v.Plural != "" {
				return v.Plural
			}
			return valueName(v) + "Slice"
		},
		"Type": func(v *value) string {
			name := valueName(v)
			return camelToLower(name)
		},
		"InterfereType": func(v *value) string {
			if v.Type[0:1] == "*" {
				return v.Type[1:]
			}
			return v.Type
		},
		"SliceType": func(v *value) string {
			name := valueName(v)
			return camelToLower(name)
		},
	})

	{
		t, err := baseT.Parse(removeNon(tmpl))
		fatalIfError(err)

		w, err := os.Create("values_generated.go")
		fatalIfError(err)
		defer w.Close()

		err = t.Execute(w, struct {
			Values       []value
			Imports      []string
			MapKeysTypes []string
		}{
			Values:       values,
			Imports:      imports,
			MapKeysTypes: stringifyKinds(mapAllowedKinds),
		})
		fatalIfError(err)

		gofmt("values_generated.go")
	}

	{
		t, err := baseT.Parse(removeNon(testTmpl))
		fatalIfError(err)

		w, err := os.Create("values_generated_test.go")
		fatalIfError(err)
		defer w.Close()

		err = t.Execute(w, struct {
			Values       []value
			Imports      []string
			MapKeysTypes []string
		}{
			Values:       values,
			Imports:      imports,
			MapKeysTypes: stringifyKinds(mapAllowedKinds),
		})
		fatalIfError(err)

		gofmt("values_generated_test.go")
	}

}

func stringifyKinds(kinds []reflect.Kind) []string {
	var l []string

	for _, kind := range kinds {
		l = append(l, kind.String())
	}

	return l
}

func gofmt(path string) {
	cmd := exec.Command("goimports", "-w", path)
	b, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("gofmt error: %s\n%s", err, b)
	}
}

// transform s from CamelCase to mixedCase
func camelToLower(s string) string {
	splitted := split(s)
	splitted[0] = strings.ToLower(splitted[0])
	return strings.Join(splitted, "")
}

// This part was taken from the https://github.com/fatih/camelcase package
// This part is licensed under MIT license
// Copyright (c) 2015 Fatih Arslan
//
// Split splits the camelcase word and returns a list of words. It also
// supports digits. Both lower camel case and upper camel case are supported.
// For more info please check: http://en.wikipedia.org/wiki/CamelCase
//
// Examples
//
//   "" =>                     [""]
//   "lowercase" =>            ["lowercase"]
//   "Class" =>                ["Class"]
//   "MyClass" =>              ["My", "Class"]
//   "MyC" =>                  ["My", "C"]
//   "HTML" =>                 ["HTML"]
//   "PDFLoader" =>            ["PDF", "Loader"]
//   "AString" =>              ["A", "String"]
//   "SimpleXMLParser" =>      ["Simple", "XML", "Parser"]
//   "vimRPCPlugin" =>         ["vim", "RPC", "Plugin"]
//   "GL11Version" =>          ["GL", "11", "Version"]
//   "99Bottles" =>            ["99", "Bottles"]
//   "May5" =>                 ["May", "5"]
//   "BFG9000" =>              ["BFG", "9000"]
//   "BöseÜberraschung" =>     ["Böse", "Überraschung"]
//   "Two  spaces" =>          ["Two", "  ", "spaces"]
//   "BadUTF8\xe2\xe2\xa1" =>  ["BadUTF8\xe2\xe2\xa1"]
//
// Splitting rules
//
//  1) If string is not valid UTF-8, return it without splitting as
//     single item array.
//  2) Assign all unicode characters into one of 4 sets: lower case
//     letters, upper case letters, numbers, and all other characters.
//  3) Iterate through characters of string, introducing splits
//     between adjacent characters that belong to different sets.
//  4) Iterate through array of split strings, and if a given string
//     is upper case:
//       if subsequent string is lower case:
//         move last character of upper case string to beginning of
//         lower case string
func split(src string) (entries []string) {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	// split into fields based on class of unicode character
	for _, r := range src {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}
	return
}
