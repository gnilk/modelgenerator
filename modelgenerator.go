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
	"strconv"
	"strings"

	"modelgenerator/common"
	"modelgenerator/generators/cpp"
	golang "modelgenerator/generators/golang"
	"modelgenerator/generators/typescript"
)

const Name = "ModelGenerator"
const Version = "2.1"

//
// Load's an XML document an preprocess (load and merge any include directive)
//
func loadDocument(options *common.Options, Filename string) (common.XMLDoc, error) {
	var doc common.XMLDoc

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

func preprocessDocument(options *common.Options, doc *common.XMLDoc) {
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
		include.Document = incDoc
		mergeDocuments(doc, &incDoc)
	}
}

func mergeDocuments(dst *common.XMLDoc, src *common.XMLDoc) *common.XMLDoc {
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

	return dst
}

//
// Generate domain model for selected language
//
func generateLanguageModel(options *common.Options, doc common.XMLDoc) {

	codeGenerator := options.Language.GetModelGenerator()
	code := codeGenerator.GenerateCode(doc, options)

	if options.OutputName != "-" {
		byteCode := []byte(code)
		ioutil.WriteFile(options.OutputName, byteCode, 0644)
	} else {
		log.Printf("%s\n", code)
	}
}

//
// generate persistence layer for selected language
//
func generatePersistence(options *common.Options, doc common.XMLDoc) {
	crudGenerator := options.Language.GetCrudGenerator()
	if crudGenerator != nil {
		if options.Verbose > 0 {
			log.Printf("Generating persistence code, saving to '%s'", options.OutputDBName)
			log.Printf("  DB Control: %v\n", doc.DBControl)
		}
		//var persistenceCode = generatePersistenceCode(doc, options.PersistenceClass, options.Filename, options.SplitInFiles, options.Converters, options.Verbose, options.OutputName)
		var persistenceCode = crudGenerator.GenerateCode(doc, options)
		persistenceByteCode := []byte(persistenceCode)
		ioutil.WriteFile(options.OutputDBName, persistenceByteCode, 0644)
	} else {
		log.Printf("No Crud generator for language\n")
	}

	//
	// Create DB Create/Alter script - this is dumped to STDOUT
	//
	dbGenerator := options.Language.GetDBCreateGenerator()
	if dbGenerator != nil {
		var dbCreateCode = dbGenerator.GenerateCode(doc, options)
		if options.Verbose > 0 {
			log.Printf("dbCreateCode:\n")
		}
		fmt.Printf("%s\n", dbCreateCode)
	} else {
		log.Printf("No DB Script Generator\n")
	}
}

func printHelp() {
	fmt.Printf("%s %s - XML Data Model to Language structure converter\n", Name, Version)
	fmt.Println("Usage: modelgenerator [-sv] [-p <class>] [-f <num>] [-o <file/dir>] <inputfile>")
	fmt.Println("General Options")
	fmt.Println("  -f : From Version, generates any class/field matching >= specified version (0 means as virgin)")
	fmt.Println("  -p : Generate persistence (use optional 'class' to specifiy which class for persistence, or '-' for all - default)")
	fmt.Println("  -s : split each type in separate file")
	fmt.Println("  -l : specify output language (go/cpp/ts)")
	fmt.Println("Domain Model Options")
	fmt.Println("  -c : generate convertes (to/from XML/JSON)")
	fmt.Println("  -g : disable getters/setters")
	fmt.Println("  -o : specify output model file or '-' for stdout (default) ")
	fmt.Println("DB Layer Options")
	fmt.Println("  -P : Table name prefix (default is 'nagini_se_')")
	fmt.Println("  -d : Generate drop statements before create (default = false)")
	fmt.Println("  -O : specify output database go file or dir (if split in multiplefiles is true), default is 'db.go'")
	fmt.Println("  -v : increase verbose output (default 0 - none)")
	fmt.Println("  -h : this page")
	fmt.Println("inputfile : XML Data Model definition file")
	fmt.Println("")
}

func getLanguage(name string) common.Language {
	switch strings.ToLower(name) {
	case "go":
		fallthrough
	case "golang":
		log.Printf("Creating generators for GO\n")
		return golang.CreateGoLanguage()
	case "cpp":
		fallthrough
	case "c++":
		log.Printf("Creating generators for C++\n")
		return cpp.CreateCppLanguage()
	case "typescript":
		fallthrough
	case "ts":
		log.Printf("Creating generators for TS (TypeScript)\n")
		return typescript.CreateTSLanguage()
	}
	log.Fatalf("No support for language: %s\n", name)
	return nil
}

func main() {
	options := common.Options{
		SplitInFiles:          false,
		Converters:            false,
		Verbose:               0,
		DBTablePrefix:         "nagini_se_",
		Filename:              "",
		OutputName:            "-",
		OutputDBName:          "db.go",
		PersistenceClass:      "-",
		AllPersistenceClasses: nil,
		UseLanguage:           "go",
		GettersAndSetters:     true,
		DoPersistence:         false,
		IsUpgrade:             false,
		FromVersion:           0, // Always assume from version 0
		GenerateDropStatement: false,
	}

	if len(os.Args) > 1 {

		for i := 0; i < len(os.Args); i++ {
			arg := os.Args[i]
			//log.Printf("Arg: %s\n", arg)
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
				case 'l':
					i++
					options.UseLanguage = os.Args[i]
					break
				case 'f':
					i++
					options.FromVersion, _ = strconv.Atoi(os.Args[i])
					if options.FromVersion > 0 {
						options.IsUpgrade = true
					}
					break
				case 'g':
					options.GettersAndSetters = false
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

		log.Printf("%s %s\n", Name, Version)

		log.Printf("Processing file: %s\n", options.Filename)
		log.Printf("With root directory: %s\n", filepath.Dir(intputFilePath))
		log.Printf("Generating from version: %d\n", options.FromVersion)
		log.Printf("Output language: %s\n", options.UseLanguage)
	}

	options.Language = getLanguage(options.UseLanguage)
	if options.Language == nil {
		log.Fatalf("No implementation for '%s'\n", options.UseLanguage)
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
		log.Println("File read ok, generating data model code...")
	}

	generateLanguageModel(&options, doc)

	if options.DoPersistence {
		generatePersistence(&options, doc)
	}
}
