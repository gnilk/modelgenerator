package main

//
// Implements a XML to GO datamodel transformation
// Creates some kind of POCO/POJO GO objects from the XML definition
//

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

//
// Options for generator, takem from command line
//
type Options struct {
	SplitInFiles          bool
	DBTablePrefix         string
	Converters            bool
	Verbose               int
	Filename              string
	OutputName            string
	OutputDBName          string
	AllPersistenceClasses []string
	PersistenceClass      string
	DoPersistence         bool
	IsUpgrade             bool
	GenerateDropStatement bool
	FromVersion           int    // Always assume from version 0
	DocumentRootDirectory string // This is set by code to the root directory of the first document, relative for all includes
	CurrentDoc            *XMLDoc
}

//
// Reflection type name prefix
//
var xmlTypePrefix = "main."

//
// XML Import structures
//

type XMLDBControl struct {
	Host     string `xml:"host"`
	DBName   string `xml:"dbname"`
	User     string `xml:"user"`
	Password string `xml:"password"`
}

// XMLDataTypeField declares a variable in an definition (see XmlDefine)
type XMLDataTypeField struct {
	Name            string `xml:"name,attr"`
	Default         string `xml:"default,attr"`
	Value           int    `xml:"value,attr"`
	DBSize          int    `xml:"dbsize,attr"`
	FieldSize       int    `xml:"fieldsize,attr"`
	Type            string `xml:"type,attr"`
	IsPointer       bool   `xml:"ispointer,attr"`
	IsList          bool   `xml:"islist,attr"`
	FromVersion     int    `xml:"fromversion,attr"`
	SkipPersistance bool   `xml:"nopersist,attr"`
	DBAutoID        bool   `xml:"dbautoid,attr"`
}

// XMLDefine declares an object (type/struct)
type XMLDefine struct {
	Type            string             `xml:"type,attr"`
	Name            string             `xml:"name,attr"`
	Inherits        string             `xml:"inherits,attr"`
	DBSchema        string             `xml:"dbschema,attr"`
	SkipPersistance bool               `xml:"nopersist,attr"`
	Fields          []XMLDataTypeField `xml:"field"`
	Guids           []XMLDataTypeField `xml:"guid"`
	Strings         []XMLDataTypeField `xml:"string"`
	Bools           []XMLDataTypeField `xml:"bool"`
	Times           []XMLDataTypeField `xml:"time"`
	Ints            []XMLDataTypeField `xml:"int"`
	Lists           []XMLDataTypeField `xml:"list"`
	Objects         []XMLDataTypeField `xml:"object"`
	Enums           []XMLDataTypeField `xml:"enum"`

	// private stuff
	Methods []AccessMethod
}

// XMLImport holds import directives
type XMLImport struct {
	DisablePersistence bool   `xml:"no_persistence,attr"`
	Package            string `xml:",innerxml"`
}

// XMLTypeMapping Holds type mappings definitions
type XMLTypeMapping struct {
	FromType  string `xml:"from,attr"`
	ToType    string `xml:"to,attr"`
	FieldSize int    `xml:"fieldsize,attr"`
}

type XMLInclude struct {
	Filename string `xml:",innerxml"`
	document XMLDoc
}

// XMLDoc holds the document root
type XMLDoc struct {
	Namespace      string           `xml:"namespace,attr"`
	DBSchema       string           `xml:"dbschema,attr"`
	Includes       []XMLInclude     `xml:"include"`
	Imports        []XMLImport      `xml:"imports>package"`
	Defines        []XMLDefine      `xml:"define"`
	DBTypeMappings []XMLTypeMapping `xml:"dbtypemappings>map"`
	GOTypeMappings []XMLTypeMapping `xml:"gotypemappings>map"`
	DBControl      XMLDBControl     `xml:"dbcontrol"`
}

//
// Internal structures
//

// XMLToGoTypeTranslation defines the translation between XML types and GO types
var XMLToGoTypeTranslation = map[string]string{
	"XmlString": "string",
	"XmlTime":   "time",
	"XmlGuid":   "guid",
	"XmlObject": "object",
	"XmlList":   "list",
	"XmlInt":    "int",
}

//
// Meta info for structure access methods (getter/setters)
//

type AccessMethod struct {
	getter    bool
	setter    bool
	isList    bool
	isPointer bool
	noPersist bool
	autoID    bool
	Name      string
	Type      string
}

func loadDocument(options *Options, Filename string) (XMLDoc, error) {
	var doc XMLDoc

	xmlData, err := ioutil.ReadFile(Filename)
	if err != nil {
		log.Println("Error while opening file: ", err)
		return doc, err
	}

	err = xml.Unmarshal(xmlData, &doc)
	if err != nil {
		log.Println("Error while unmarshalling XML:", err)
		return doc, err
	}
	preprocessDocument(options, &doc)
	return doc, nil
}

func preprocessDocument(options *Options, doc *XMLDoc) {
	for _, include := range doc.Includes {
		incPathName := filepath.Join(options.DocumentRootDirectory, include.Filename)

		if options.Verbose > 0 {
			log.Printf("Including: %s\n", include.Filename)
			log.Printf("Pathname: %s\n", incPathName)
		}
		//options.DocumentRootDirectory = filepath.Dir(intputFilePath)

		incDoc, err := loadDocument(options, incPathName)
		if err != nil {
			log.Fatalf("Unable to include file: %s (%s)\n", include.Filename, incPathName)
			return
		}
		include.document = incDoc
		mergeDocuments(doc, &incDoc)
	}
}

func mergeDocuments(dst *XMLDoc, src *XMLDoc) *XMLDoc {
	//
	// Note: We don't merge includes!!!
	//
	dst.Imports = append(dst.Imports, src.Imports...)
	dst.Defines = append(dst.Defines, src.Defines...)
	dst.DBTypeMappings = append(dst.DBTypeMappings, src.DBTypeMappings...)
	dst.GOTypeMappings = append(dst.GOTypeMappings, src.GOTypeMappings...)
	if src.DBControl.DBName != "" {
		dst.DBControl.DBName = src.DBControl.DBName
	}
	if src.DBControl.Host != "" {
		dst.DBControl.Host = src.DBControl.Host
	}
	if src.DBControl.Password != "" {
		dst.DBControl.Password = src.DBControl.Password
	}
	if src.DBControl.User != "" {
		dst.DBControl.User = src.DBControl.User
	}
	//	dst.DBControl = src.DBControl

	return dst
}

func translateXMLType(xmlType interface{}) (string, string) {
	//reflect.TypeOf(xmlType).Name
	//return strings.TrimPrefix(reflect.TypeOf(xmlType).String(), xmlTypePrefix)

	var xmlTypeName = reflect.TypeOf(xmlType).Name()
	var goTypeName = XMLToGoTypeTranslation[xmlTypeName]

	return goTypeName, xmlTypeName

}

func (field *XMLDataTypeField) dump() {
	fmt.Printf("  Name: %s\n", field.Name)
	fmt.Printf("  Default: %s\n", field.Default)
}

func (define *XMLDefine) generateCode(options *Options, converters bool) string {

	code := ""
	//fmt.Printf("Generating code for type='%s', named='%s'\n", define.Type, define.Name)
	switch define.Type {
	case "class":
		code += define.generateClassCode(options)
		if converters {
			code += define.generateClassConverters()
		}
		break
	case "enum":
		code += define.generateEnumCode()
		break
	default:
		fmt.Printf("[XMLDefine::generateCode] Error, can't generate code for type '%s'\n", define.Type)
		break
	}

	return code
}

func (define *XMLDefine) generateClassCode(options *Options) string {
	code := ""

	code += fmt.Sprintf("//\n")
	code += fmt.Sprintf("// %s is generated\n", define.Name)
	code += fmt.Sprintf("//\n")

	code += fmt.Sprintf("type %s struct {\n", define.Name)

	if define.Inherits != "" {
		code += fmt.Sprintf("  %s\n\n", define.Inherits)
	}

	code += define.generateFields(options, define.Fields)

	// code += define.generateFieldCode(define.Guids, "uuid.UUID")
	// code += define.generateFieldCode(define.Strings, "string")
	// code += define.generateFieldCode(define.Ints, "int")
	// code += define.generateFieldCode(define.Bools, "bool")
	// // TODO: Fix this!!
	// code += define.generateFieldCode(define.Times, "time.Time")
	// code += define.generateFieldListCode(define.Lists)
	// code += define.generateFieldEnumCode(define.Enums)
	// code += define.generateFieldObjectCode(define.Objects)
	code += fmt.Sprintf("}\n")
	code += fmt.Sprintf("\n")

	for _, method := range define.Methods {
		ptrAttrib := "*"
		if method.isPointer != true {
			ptrAttrib = ""
		}

		if method.getter == true {
			if method.isList != true {
				code += fmt.Sprintf("func (this *%s) Get%s() %s%s {\n", define.Name, method.Name, ptrAttrib, method.Type)
				code += fmt.Sprintf("  return this.%s\n", method.Name)
				code += fmt.Sprintf("}\n")
				code += fmt.Sprintf("\n")
			} else {
				code += fmt.Sprintf("func (this *%s) Get%sAsRef() []%s%s {\n", define.Name, method.Name, ptrAttrib, method.Type)
				code += fmt.Sprintf("  return this.%s[:len(this.%s)]\n", method.Name, method.Name)
				code += fmt.Sprintf("}\n")
				code += fmt.Sprintf("\n")

				code += fmt.Sprintf("func (this *%s) Get%sAsCopy() []%s%s {\n", define.Name, method.Name, ptrAttrib, method.Type)
				code += fmt.Sprintf("  newSlice := make([]%s, len(this.%s))\n", method.Type, method.Name)
				code += fmt.Sprintf("  copy(newSlice, this.%s)\n", method.Name)
				code += fmt.Sprintf("  return newSlice\n")
				code += fmt.Sprintf("}\n")
				code += fmt.Sprintf("\n")
			}
		}

		if method.setter == true {
			if method.isList != true {
				code += fmt.Sprintf("func (this *%s) Set%s(value %s%s) {\n", define.Name, method.Name, ptrAttrib, method.Type)
				code += fmt.Sprintf("  this.%s = value\n", method.Name)
				code += fmt.Sprintf("}\n")
				code += fmt.Sprintf("\n")
			} else {
				code += fmt.Sprintf("func (this *%s) Set%s(value []%s%s) {\n", define.Name, method.Name, ptrAttrib, method.Type)
				code += fmt.Sprintf("  this.%s = make([]%s, len(value))\n", method.Name, method.Type)
				code += fmt.Sprintf("  copy(this.%s, value)\n", method.Name)
				code += fmt.Sprintf("}\n")
				code += fmt.Sprintf("\n")
			}
		}
	}

	return code
}

func (define *XMLDefine) generateClassConverters() string {
	code := ""

	code += define.generateToJSONCode()
	code += fmt.Sprintf("\n")
	code += define.generateToXMLCode()
	code += fmt.Sprintf("\n")

	code += define.generateFromJSONCode()
	code += fmt.Sprintf("\n")
	code += define.generateFromXMLCode()
	code += fmt.Sprintf("\n")
	return code
}

func (define *XMLDefine) generateToJSONCode() string {
	code := ""
	code += fmt.Sprintf("// ToJSON creates a JSON representation of the data for the type\n")
	code += fmt.Sprintf("func (this *%s) ToJSON() string {\n", define.Name)

	code += fmt.Sprintf("  b, err := json.MarshalIndent(this, \"\", \"    \")\n")
	code += fmt.Sprintf("  if err != nil {\n")
	code += fmt.Sprintf("    return \"\"\n")
	code += fmt.Sprintf("  }\n")
	code += fmt.Sprintf("  return bytes.NewBuffer(b).String()\n")
	code += fmt.Sprintf("}\n")

	// code += fmt.Sprintf("  b := new(bytes.Buffer)\n")
	// code += fmt.Sprintf("  encoder := json.NewEncoder(b)\n")
	// code += fmt.Sprintf("  encoder.SetIndent(\"\", \"    \")\n")
	// code += fmt.Sprintf("  err := encoder.Encode(this)\n")
	// code += fmt.Sprintf("  if err != nil {\n")
	// code += fmt.Sprintf("    return \"\"\n")
	// code += fmt.Sprintf("  }\n")
	// code += fmt.Sprintf("  return b.String()\n")
	return code
}

func (define *XMLDefine) generateToXMLCode() string {
	code := ""
	code += fmt.Sprintf("// ToXML creates an XML representation of the data for the type\n")
	code += fmt.Sprintf("func (this *%s) ToXML() string {\n", define.Name)

	code += fmt.Sprintf("  b, err := xml.MarshalIndent(this, \"\", \"    \")\n")
	code += fmt.Sprintf("  if err != nil {\n")
	code += fmt.Sprintf("    return \"\"\n")
	code += fmt.Sprintf("  }\n")
	code += fmt.Sprintf("  return bytes.NewBuffer(b).String()\n")
	code += fmt.Sprintf("}\n")

	// code += fmt.Sprintf("  b := new(bytes.Buffer)\n")
	// code += fmt.Sprintf("  encoder := xml.NewEncoder(b)\n")
	// code += fmt.Sprintf("  encoder.Indent(\"\", \"    \")\n")
	// code += fmt.Sprintf("  err := encoder.Encode(this)\n")
	// code += fmt.Sprintf("  if err != nil {\n")
	// code += fmt.Sprintf("    return \"\"\n")
	// code += fmt.Sprintf("  }\n")
	// code += fmt.Sprintf("  return b.String()\n")
	return code
}

func (define *XMLDefine) generateFromXMLCode() string {
	code := ""

	code += fmt.Sprintf("// %sFromXML converts an XML representation to the type\n", define.Name)
	code += fmt.Sprintf("func %sFromXML(xmldata string) (*%s, error) {\n", define.Name, define.Name)
	code += fmt.Sprintf("  var value %s\n", define.Name)
	code += fmt.Sprintf("  err := xml.Unmarshal([]byte(xmldata), &value)\n")
	code += fmt.Sprintf("  if err != nil {\n")
	code += fmt.Sprintf("	  return nil, err\n")
	code += fmt.Sprintf("  }\n")
	code += fmt.Sprintf("  return &value, nil\n")
	code += fmt.Sprintf("}\n")

	return code

	// -- Original code (this is what it should be)
	// func UserFromXML(xmldata string) (*User, error) {
	// 	var user User
	// 	err := xml.Unmarshal([]byte(xmldata), &user)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return &user, nil
	// }

}

func (define *XMLDefine) generateFromJSONCode() string {
	code := ""

	code += fmt.Sprintf("// %sFromJSON converts a JSON representation to the data type\n", define.Name)
	code += fmt.Sprintf("func %sFromJSON(jsondata string) (*%s, error) {\n", define.Name, define.Name)
	code += fmt.Sprintf("  var value %s\n", define.Name)
	code += fmt.Sprintf("  err := json.Unmarshal([]byte(jsondata), &value)\n")
	code += fmt.Sprintf("  if err != nil {\n")
	code += fmt.Sprintf("	  return nil, err\n")
	code += fmt.Sprintf("  }\n")
	code += fmt.Sprintf("  return &value, nil\n")
	code += fmt.Sprintf("}\n")
	return code

	// -- original code
	// func UserFromJSON(jsondata string) (*User, error) {
	// 	var user User
	// 	err := json.Unmarshal([]byte(jsondata), &user)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return &user, nil
	// }
}

func (define *XMLDefine) methodFromField(field XMLDataTypeField, typeString string, isList bool) {
	method := AccessMethod{
		getter:    true,
		setter:    true,
		isList:    isList,
		Name:      field.Name,
		Type:      typeString,
		isPointer: field.IsPointer,
		noPersist: field.SkipPersistance,
		autoID:    field.DBAutoID,
	}

	define.Methods = append(define.Methods, method)
}

// TypeMapping maps a field's type to potential mappings from the specific table
func (field *XMLDataTypeField) TypeMapping(typeMappings []XMLTypeMapping) string {
	for _, mapping := range typeMappings {
		if mapping.FromType == field.Type {
			// If this is not specified, assume the type mapping does not require it
			if mapping.FieldSize == 0 {
				return mapping.ToType
			}
			fs := field.FieldSize // Assume field has this defined
			if fs == 0 {
				fs = mapping.FieldSize
			}
			return fmt.Sprintf(mapping.ToType, fs)

		}
	}
	return field.Type
}

func (field *XMLDataTypeField) goFieldCode(options *Options) string {
	code := ""

	typePrefix := ""
	if field.IsList {
		typePrefix = typePrefix + "[]"
	}
	if field.IsPointer {
		typePrefix = typePrefix + "*"
	}
	code += fmt.Sprintf("  %s %s%s\n", field.Name, typePrefix, field.TypeMapping(options.CurrentDoc.GOTypeMappings))

	return code
}

func (define *XMLDefine) generateFields(options *Options, list []XMLDataTypeField) string {
	code := ""

	for _, field := range list {
		code += field.goFieldCode(options)
		define.methodFromField(field, field.TypeMapping(options.CurrentDoc.GOTypeMappings), field.IsList)
	}

	return code
}

// generateEnumCode creates Go Code for an ENUM const declaration
func (define *XMLDefine) generateEnumCode() string {
	code := ""
	code += fmt.Sprintf("type %s int64\n", define.Name)
	code += fmt.Sprintf("const (\n")
	code += fmt.Sprintf("  _ = iota\n")
	for _, Int := range define.Ints {
		code += fmt.Sprintf("  %s %s = %d\n", Int.Name, define.Name, Int.Value)
	}
	code += fmt.Sprintf(")\n")
	code += fmt.Sprintf("\n")

	code += fmt.Sprintf("var map%sToName=map[%s]string {\n", define.Name, define.Name)
	for _, Int := range define.Ints {
		code += fmt.Sprintf("  %d:\"%s\",\n", Int.Value, Int.Name)
	}
	// todo generate map here
	code += fmt.Sprintf("}\n")
	code += fmt.Sprintf("\n")

	code += fmt.Sprintf("func (this *%s)String() string {\n", define.Name)
	code += fmt.Sprintf("  return map%sToName[*this]\n", define.Name)
	code += fmt.Sprintf("}\n")
	code += fmt.Sprintf("\n")

	return code
}

func (define *XMLDefine) dump() {
	fmt.Printf("\nName: %s\n", define.Name)
	fmt.Printf("  Type: %s\n", define.Type)
	fmt.Printf("  Inherits: %s\n", define.Inherits)

	fmt.Printf("  -- Fields --\n")

	// process variables per type - is there a better way to do this????
	for _, GUID := range define.Guids {
		var s, x = translateXMLType(GUID)
		fmt.Printf("  Type: %s (%s)\n", s, x)
		GUID.dump()
	}
	for _, Str := range define.Strings {
		var s, x = translateXMLType(Str)
		fmt.Printf("  Type: %s (%s)\n", s, x)
		Str.dump()
	}
	for _, Time := range define.Times {
		var s, x = translateXMLType(Time)
		fmt.Printf("  Type: %s (%s)\n", s, x)
		Time.dump()
	}

	for _, Int := range define.Ints {
		var s, x = translateXMLType(Int)
		fmt.Printf("  Type: %s (%s)\n", s, x)
		Int.dump()
	}

	for _, List := range define.Lists {
		var s, x = translateXMLType(List)
		fmt.Printf("  Type: %s (%s)\n", s, x)
		List.dump()
	}

	for _, Object := range define.Objects {
		var s, x = translateXMLType(Object)
		fmt.Printf("  Type: %s (%s)\n", s, x)
		Object.dump()
	}

}

func dumpXMLDefinitions(doc XMLDoc) {
	fmt.Printf("Dumping results\n")
	fmt.Printf("Namespace: %s\n", doc.Namespace)

	fmt.Printf("\nDefines\n")
	fmt.Printf("-------\n")
	for i := 0; i < len(doc.Defines); i++ {
		doc.Defines[i].dump()
	}
}

func dumpImports(doc XMLDoc) {
	log.Printf("Imports: %d\n", len(doc.Imports))
	for _, imp := range doc.Imports {
		log.Printf("  %s\n", imp.Package)
	}
}

//
// this generates the data model code...
//

func testReturnSliceRef(data []int) []int {
	return data[:]
}
func testReturnSliceCopy(data []int) []int {
	newSlice := make([]int, len(data))
	copy(newSlice, data)
	return newSlice
}

func testSlices() {
	data := []int{1, 2, 3, 4, 5}
	fmt.Println("Slices with references")
	fmt.Println("before")
	fmt.Println(data)
	data2 := testReturnSliceRef(data)
	fmt.Println("after")
	fmt.Println(data2)

	data[0] = 9
	fmt.Println("data")
	fmt.Println(data)
	fmt.Println("data2 - should be equal to first slice")
	fmt.Println(data2)

	data[0] = 1

	fmt.Println("Slices with copy")
	fmt.Println("before")
	fmt.Println(data)
	data3 := testReturnSliceCopy(data)
	fmt.Println("after")
	fmt.Println(data3)

	data[0] = 9
	fmt.Println("data")
	fmt.Println(data)
	fmt.Println("data3 - should be unchanged at first element")
	fmt.Println(data3)
}

func printHelp() {
	fmt.Println("modelgenerator v2 - XML Data Model to GO structure converter")
	fmt.Println("Usage: modelgenerator [-sv] [-p <class>] [-f <num>] [-o <file/dir>] <inputfile>")
	fmt.Println("Options")
	fmt.Println("  -f : From Version, generates any class/field matching >= specified version (0 means as virgin)")
	fmt.Println("  -p : Generate persistence (use optional 'class' to specifiy which class for persistence, or '-' for all - default)")
	fmt.Println("  -d : Generate drop statements before create (default = false)")
	fmt.Println("  -s : split each type in separate file")
	fmt.Println("  -c : generate convertes (to/from XML/JSON)")
	fmt.Println("  -o : specify output model go file or dir (if split in multiplefiles is true), default is stdout")
	fmt.Println("  -O : specify output database go file or dir (if split in multiplefiles is true), default is 'db.go'")
	fmt.Println("  -v : increase verbose output (default 0 - none)")
	fmt.Println("  -h : this page")
	fmt.Println("inputfile : XML Data Model definition file")
	fmt.Println("")
}

func init() {

	// const (
	// 	fromversion_default = 0,
	// 	fromversion_usage = "Changes DB Creation from assuming ALTER instead of CREATE, fields parsed based on >= 'fromversion' attribute",

	// 	persistence_class_default = "-",
	// 	persistence_class_usage = "Specifies explict class to generate persistence code for (default: '-', all)",

	// 	outputname_default = "",
	// 	outputname_usage = "Specifies output file or directory (if split in multiple is true), default is stdout",

	// )

	// flag.IntVar(&options.FromVersion, "-f", fromversion_default, fromversion_usage)
	// flag.IntVar(&options.FromVersion, "--from_version", fromversion_default, fromversion_usage)
	// flag.StringVar(&options.PersistenceClass, "-p", persistence_class_default, persistence_class_usage)
	// flag.StringVar(&options.PersistenceClass, "--persistence_class", persistence_class_default, persistence_class_usage)

	// flag.StringVar(&options.OutputName, "-o", outputname_default, outputname_usage)
	// flag.StringVar(&options.OutputName, "--output", outputname_default, outputname_usage)

}

func main() {
	options := Options{
		SplitInFiles:          false,
		Converters:            false,
		Verbose:               0,
		DBTablePrefix:         "nagini_se_",
		Filename:              "",
		OutputName:            "",
		OutputDBName:          "db.go",
		PersistenceClass:      "-",
		AllPersistenceClasses: nil,
		DoPersistence:         false,
		IsUpgrade:             false,
		FromVersion:           0, // Always assume from version 0
		GenerateDropStatement: false,
	}

	if len(os.Args) > 1 {

		for i := 0; i < len(os.Args); i++ {
			arg := os.Args[i]
			///fmt.Printf("Arg: %s\n", arg)
			if arg[0] == '-' {
				switch arg[1] {
				case 's':
					options.SplitInFiles = true
					break
				case 'c':
					options.Converters = true
					break
				case 'd':
					options.GenerateDropStatement = true
					break
				case 'f':
					i++
					options.FromVersion, _ = strconv.Atoi(os.Args[i])
					if options.FromVersion > 0 {
						options.IsUpgrade = true
					}
					break
				case 'p':
					options.DoPersistence = true
					if os.Args[i+1][0] != '-' {
						i++
						options.AllPersistenceClasses = strings.Split(os.Args[i], ",")
						options.PersistenceClass = options.AllPersistenceClasses[0] //os.Args[i]
					} else {
						i++
						options.PersistenceClass = os.Args[i]
					}
					break
				case 'P':
					i++
					options.DBTablePrefix = os.Args[i]
					break
				case 'v':
					options.Verbose++
					break
				case 'o':
					i++
					options.OutputName = os.Args[i]
					break
				case 'O':
					i++
					options.OutputDBName = os.Args[i]
					break
				case 'h':
					printHelp()
					return
				default:
					log.Printf("Error: Unknown argument %s\n", arg)
					printHelp()
				}
			} else {
				options.Filename = arg
			}
		}
	}

	if options.Filename == "" {
		printHelp()
		return
	}

	intputFilePath, err := filepath.Abs(options.Filename)
	options.DocumentRootDirectory = filepath.Dir(intputFilePath)

	if options.Verbose > 0 {
		log.Printf("Processing file: %s\n", options.Filename)
		log.Printf("With root directory: %s\n", filepath.Dir(intputFilePath))
		log.Printf("Generating from version: %d\n", options.FromVersion)
	}
	//fmt.Printf("Processing file: %s\n", filename)
	//fmt.Println("Unmarshal XML")

	doc, err := loadDocument(&options, options.Filename)

	if err != nil {
		log.Fatalln("Failed to load document: " + options.Filename)
		return
	}

	options.CurrentDoc = &doc // set this so we have access

	if options.Verbose > 0 {
		log.Printf("DB Typemappoings: %d\n", len(doc.DBTypeMappings))
		log.Printf("GO Typemappoings: %d\n", len(doc.GOTypeMappings))
		dumpImports(doc)

		log.Println("File read ok, generating data model code...")
	}
	//
	// this generates the data model code
	//
	var code = generateCode(&options, doc, options.Filename, options.SplitInFiles, options.Converters, options.Verbose, options.OutputName)
	if options.OutputName != "" {
		byteCode := []byte(code)
		ioutil.WriteFile(options.OutputName, byteCode, 0644)
	} else {
		log.Printf("%s\n", code)
	}

	if options.DoPersistence {
		//
		// this generates the persistence code
		//
		if options.Verbose > 0 {
			log.Printf("Generating persistence code, saving to '%s'", options.OutputDBName)
			log.Printf("  DB Control: %v\n", doc.DBControl)
		}
		//var persistenceCode = generatePersistenceCode(doc, options.PersistenceClass, options.Filename, options.SplitInFiles, options.Converters, options.Verbose, options.OutputName)
		var persistenceCode = generatePersistenceCode(doc, &options)
		persistenceByteCode := []byte(persistenceCode)
		ioutil.WriteFile(options.OutputDBName, persistenceByteCode, 0644)

		//
		//
		//
		var dbCreateCode = generateDBCreateCode(doc, &options)
		if options.Verbose > 0 {
			log.Printf("dbCreateCode:\n")
		}
		fmt.Printf("%s\n", dbCreateCode)
	}
}
