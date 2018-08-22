package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

var (
	parsedFiles = make(map[string]struct{})
)

func parseXSDFile(fname string) ([]xsdSchema, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return parse(f, fname)
}

// makeCharsetReader returns special readers as needed for xml encodings, or
// nil.
func makeCharsetReader(charset string, input io.Reader) (io.Reader, error) {
	if charset == "Windows-1252" {
		return charmap.Windows1252.NewDecoder().Reader(input), nil
	}
	return nil, fmt.Errorf("Unknown charset: %s", charset)
}

func parse(r io.Reader, fname string) ([]xsdSchema, error) {
	var schema xsdSchema

	d := xml.NewDecoder(r)
	// handle special character sets
	d.CharsetReader = makeCharsetReader
	if err := d.Decode(&schema); err != nil {
		return nil, err
	}

	schemas := []xsdSchema{schema}
	dir, file := filepath.Split(fname)
	parsedFiles[file] = struct{}{}
	for _, imp := range schema.Imports {
		if _, ok := parsedFiles[imp.Location]; ok {
			continue
		}
		s, err := parseXSDFile(filepath.Join(dir, imp.Location))
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, s...)
	}

	return schemas, nil
}

// xsdSchema is the root of our Go representation of an XSD schema.
type xsdSchema struct {
	XMLName xml.Name

	Tns                string           `xml:"xmlns tns,attr"`
	Xs                 string           `xml:"xmlns xs,attr"`
	Version            string           `xml:"version,attr"`
	TargetNamespace    string           `xml:"targetNamespace,attr"`
	ElementFormDefault string           `xml:"elementFormDefault,attr"`
	Includes           []xsdInclude     `xml:"include"`
	Ns                 string           `xml:"xmlns,attr"`
	Imports            []xsdImport      `xml:"import"`
	Elements           []xsdElement     `xml:"element"`
	ComplexTypes       []xsdComplexType `xml:"complexType"`
	SimpleTypes        []xsdSimpleType  `xml:"simpleType"`
}

type xsdInclude struct {
	SchemaLocation string `xml:"schemaLocation,attr"`
}

// ns parses the namespace from a value in the expected format
// http://host/namespace/v1
func (s xsdSchema) ns() string {
	split := strings.Split(s.Ns, "/")
	if len(split) > 2 {
		return split[len(split)-2]
	}
	return ""
}

type xsdImport struct {
	Location string `xml:"schemaLocation,attr"`
}

type xsdElement struct {
	XMLName     xml.Name        `xml:"element"`
	Name        string          `xml:"name,attr"`
	Nillable    bool            `xml:"nillable,attr"`
	Ref         string          `xml:"ref,attr"`
	Type        string          `xml:"type,attr"`
	Default     string          `xml:"default,attr"`
	Min         string          `xml:"minOccurs,attr"`
	Max         string          `xml:"maxOccurs,attr"`
	Annotation  string          `xml:"annotation>documentation"`
	Sequence    []xsdElement    `xml:"sequence>choice"`
	Choice      []xsdElement    `xml:"choice>element"`
	ComplexType *xsdComplexType `xml:"complexType"` // inline complex type
	SimpleType  *xsdSimpleType  `xml:"simpleType"`  // inline simple type
	Groups      []xsdGroup      `xml:"group"`
}

func (e xsdElement) isList() bool {
	return e.Max == "unbounded"
}

func (e xsdElement) inlineType() bool {
	return e.Type == ""
}

// xsdGroup element is used to define a group of elements to be used in complex type definitions.
type xsdGroup struct {
	Name     string       `xml:"name,attr"`
	Ref      string       `xml:"ref,attr"`
	Sequence []xsdElement `xml:"sequence>choice"`
	Choice   []xsdElement `xml:"choice>element"`
	All      []xsdElement `xml:"all>element"`
}

type xsdComplexType struct {
	Name           string             `xml:"name,attr"`
	Abstract       string             `xml:"abstract,attr"`
	Annotation     string             `xml:"annotation>documentation"`
	Sequence       []xsdElement       `xml:"sequence>choice>element"`
	Choice         []xsdElement       `xml:"choice"`
	All            []xsdElement       `xml:"all>element"`
	Attributes     []xsdAttribute     `xml:"attribute"`
	ComplexContent *xsdComplexContent `xml:"complexContent"`
	SimpleContent  *xsdSimpleContent  `xml:"simpleContent"`
}

type xsdComplexContent struct {
	Extension   *xsdExtension   `xml:"extension"`
	Restriction *xsdRestriction `xml:"restriction"`
}

type xsdSimpleContent struct {
	Extension   *xsdExtension   `xml:"extension"`
	Restriction *xsdRestriction `xml:"restriction"`
}

type xsdExtension struct {
	Base       string         `xml:"base,attr"`
	Attributes []xsdAttribute `xml:"attribute"`
	Sequence   []xsdElement   `xml:"sequence>element"`
}

type xsdAttribute struct {
	Name       string `xml:"name,attr"`
	Type       string `xml:"type,attr"`
	Use        string `xml:"use,attr"`
	Annotation string `xml:"annotation>documentation"`
}

type xsdSimpleType struct {
	Name        string         `xml:"name,attr"`
	Annotation  string         `xml:"annotation>documentation"`
	Restriction xsdRestriction `xml:"restriction"`
}

type xsdRestriction struct {
	Base        string           `xml:"base,attr"`
	Pattern     xsdPattern       `xml:"pattern"`
	Enumeration []xsdEnumeration `xml:"enumeration"`
}

type xsdPattern struct {
	Value string `xml:"value,attr"`
}

type xsdEnumeration struct {
	Value string `xml:"value,attr"`
}
