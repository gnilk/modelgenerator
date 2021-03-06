package golang

//
// Generates persistence base code for the POGO class
// Simple CRUD functions - you have to extend your self
//

import (
	"fmt"
	"io/ioutil"
	"log"
	"modelgenerator/common"
	"strings"
)

// This is ugly but I don't want to rewrite fetchQueryFromString to be type-qualified in the function name.
// In case we are generating multiple classes for one domain it is required that the fetch function is different as GO don't support polymorphic functions
// This is set to 'true' after the first round of classes has been done
var createRetrieveFuncPostfix = false

// func generatePersistenceCode(doc XMLDoc, className string, source string, splitInFiles bool, converters bool, verbose int, outputDir string) string {

func (generator *CrudGenerator) GenerateCode(doc common.XMLDoc, options *common.Options) string {
	code := ""

	/*
		if options.Converters {
			generator.addImport(&doc, "bytes")         //append(doc.Imports, "bytes")
			generator.addImport(&doc, "encoding/json") //append(doc.Imports, "encoding/json")
			generator.addImport(&doc, "encoding/xml")  //append(doc.Imports, "encoding/xml")

			// doc.Imports = append(doc.Imports, "bytes")
			// doc.Imports = append(doc.Imports, "encoding/json")
			// doc.Imports = append(doc.Imports, "encoding/xml")
		}
	*/
	if options.SplitInFiles != true {
		code += generator.generatePersistenceHeader(doc, options)
		// generate code for all defines
		for i := 0; i < len(doc.Defines); i++ {
			if doc.Defines[i].SkipPersistance == true {
				fmt.Printf("Skipping: %s\n", doc.Defines[i].Name)
				continue
			}
			if strings.Compare(options.PersistenceClass, "-") != 0 {
				for j := 0; j < len(options.AllPersistenceClasses); j++ {
					code += generatePersistenceCodeForDefine(&doc.Defines[i], options, options.AllPersistenceClasses[j], options.Converters)
				}
			} else {
				code += generatePersistenceCodeForDefine(&doc.Defines[i], options, options.PersistenceClass, options.Converters)
			}
		}
	} else {
		log.Panicln("SPLIT IN FILES NOT SUPPORTED!!!!")

		// generate code for all defines
		for _, define := range doc.Defines {
			code := ""
			code += generator.generatePersistenceHeader(doc, options)
			code += generatePersistenceCodeForDefine(&define, options, options.PersistenceClass, options.Converters)
			outputDir := options.OutputDBName
			if string(outputDir[len(outputDir)-1:]) != "/" {
				outputDir += "/"
			}
			fileName := outputDir + define.Name + ".go"

			if options.Verbose > 0 {
				fmt.Printf("Writing code for %s to %s\n", define.Name, fileName)
			}
			ioutil.WriteFile(fileName, []byte(code), 0644)
		}
	}
	return code
}

// func (generator *CrudGenerator) addImport(pkgName string) {
// 	generator.Imports = append(generator.Imports, common.XMLImport{
// 		DisablePersistence: false,
// 		Package:            pkgName,
// 	})
// }

func (generator *CrudGenerator) generatePersistenceHeader(doc common.XMLDoc, options *common.Options) string {

	code := ""
	code += fmt.Sprintf("package %s\n", doc.Namespace)
	code += fmt.Sprintf("\n")
	if len(doc.Imports) > 0 {
		code += fmt.Sprintf("import (\n")
		// TODO: need some more attributes here...
		// for _, Import := range doc.Imports {
		// 	importstatements := strings.Split(Import, " ")
		// 	if len(importstatements) == 1 {
		// 		code += fmt.Sprintf("  \"%s\"\n", Import)
		// 	} else {
		// 		code += fmt.Sprintf("  %s \"%s\"\n", importstatements[0], importstatements[1])
		// 	}

		// }
		// Add some static DB imports which we require
		code += fmt.Sprintf("  \"database/sql\"\n")
		code += fmt.Sprintf("  \"fmt\"\n")
		code += fmt.Sprintf("  \"log\"\n")
		code += fmt.Sprintf("  \"errors\"\n")
		//		code += fmt.Sprintf("  uuid \"github.com/satori/go.uuid\"\n")
		code += fmt.Sprintf("  // Need initialization\n")               // Ok, so I hardcoded this...
		code += fmt.Sprintf("  _ \"github.com/go-sql-driver/mysql\"\n") // Ok, so I hardcoded this...
		code += fmt.Sprintf(")\n")
	}

	code += fmt.Sprintf("//\n")
	code += fmt.Sprintf("// this code is generated by the modelgenerator\n")
	code += fmt.Sprintf("// data model source = %s\n", options.Filename)
	code += fmt.Sprintf("//\n")
	code += fmt.Sprintf("\n")

	code += fmt.Sprintf("var globalDataBase *sql.DB\n")

	schemaName := doc.DBSchema
	if len(schemaName) < 1 {
		schemaName = doc.Namespace
	}

	code += fmt.Sprintf("// Constants for DB connectivity\n")
	code += fmt.Sprintf("const (\n")
	//	code += fmt.Sprintf("   DB_NAME        = \"gnilk\"\n")
	code += fmt.Sprintf("   DB_USER        = \"%s\"\n", doc.DBControl.User)
	code += fmt.Sprintf("   DB_PASSWORD    = \"%s\"\n", doc.DBControl.Password)
	if (len(doc.DBControl.Schema) < 1) {
		code += fmt.Sprintf("   DB_SCHEMA      = \"%s%s\"\n", options.DBTablePrefix, schemaName)
	} else {
		code += fmt.Sprintf("   DB_SCHEMA      = \"%s\"\n", doc.DBControl.Schema);
	}
	code += fmt.Sprintf("   DB_HOST_MYSQL  = \"%s\"\n", doc.DBControl.Host)
	code += fmt.Sprintf("   DB_NAME_MYSQL  = \"%s\"\n", doc.DBControl.DBName)
	code += fmt.Sprintf(")\n")

	code += fmt.Sprintf("\n")

	code += fmt.Sprintf("type Persistence struct {\n")
	code += fmt.Sprintf("  db *sql.DB\n")
	code += fmt.Sprintf("}\n")
	code += fmt.Sprintf("\n")

	code += fmt.Sprintf("\n")

	// Private code - same for all persistance layers
	// This could be put in a template - it's the same for all DB layers
	code += fmt.Sprintf("func initMySQL() error {\n")
	code += fmt.Sprintf("  constr := fmt.Sprintf(\"%%s:%%s@/%%s?parseTime=true\",\n")
	code += fmt.Sprintf("       DB_USER,\n")
	code += fmt.Sprintf("       DB_PASSWORD,\n")
	code += fmt.Sprintf("       DB_NAME_MYSQL)\n")
	code += fmt.Sprintf("\n")
	code += fmt.Sprintf("  db, err := sql.Open(\"mysql\", constr)\n")
	code += fmt.Sprintf("  if err != nil {\n")
	code += fmt.Sprintf("    log.Panic(err)\n")
	code += fmt.Sprintf("    return err\n")
	code += fmt.Sprintf("  }\n")
	code += fmt.Sprintf("\n")
	code += fmt.Sprintf("  globalDataBase = db\n")
	code += fmt.Sprintf("  return nil\n")
	code += fmt.Sprintf("}\n")
	code += fmt.Sprintf("\n")
	code += fmt.Sprintf("func globalInitDb() error {\n")
	code += fmt.Sprintf("  return initMySQL()\n")
	code += fmt.Sprintf("}\n")
	code += fmt.Sprintf("\n")

	code += fmt.Sprintf("func NewPersistence() (*Persistence, error) {\n")
	code += fmt.Sprintf("  if globalDataBase == nil {\n")
	code += fmt.Sprintf("    err := globalInitDb()\n")
	code += fmt.Sprintf("    if err != nil {\n")
	code += fmt.Sprintf("      log.Println(\"globalInitDb, %%v\", err)\n")
	code += fmt.Sprintf("      return nil, err\n")
	code += fmt.Sprintf("    }\n")
	code += fmt.Sprintf("  }\n")
	code += fmt.Sprintf("  p := &Persistence{db: globalDataBase}\n")
	code += fmt.Sprintf("  return p, nil\n")
	code += fmt.Sprintf("}\n")
	code += fmt.Sprintf("\n")

	return code
}

func generatePersistenceCodeForDefine(define *common.XMLDefine, options *common.Options, className string, converters bool) string {

	// Check if class name matches - perhaps use regexp here..
	if strings.Compare(className, "-") != 0 {
		if strings.Compare(className, define.Name) != 0 {
			return ""
		}
	}

	log.Printf("Generating persistence for class: %s\n", define.Name)

	//	code += fmt.Sprintf("   DB_SCHEMA      = \"%s%s\"\n", options.DBTablePrefix, schemaName)

	dbSchemaName := fmt.Sprintf("%s%s", options.DBTablePrefix, strings.ToLower(define.Name))
	if define.DBSchema != "" {
		dbSchemaName = fmt.Sprintf("%s%s", options.DBTablePrefix, define.DBSchema)
	}

	code := ""
	code += fmt.Sprintf("const DB_SCHEMA_%s = \"%s\"\n", strings.ToUpper(define.Name), dbSchemaName)

	//fmt.Printf("Generating code for type='%s', named='%s'\n", define.Type, define.Name)
	switch define.Type {
	case "class":
		code += generatePersistenceCreateCode(define)
		code += generatePersistenceFetchCode(define)
		code += generatePersistenceRetrieveCode(define)
		code += generatePersistenceUpdateCode(define)
		code += generatePersistenceDeleteCode(define)
		// if converters {
		// 	code += define.generateClassConverters()
		// }
		break
	case "enum": // No code for this one!!
		return ""
	default:
		log.Fatalf("[XMLDefine::generatePersistenceCode] Error, can't generate code for type '%s'\n", define.Type)
		break
	}

	// Create postfix on fetch query function
	createRetrieveFuncPostfix = true

	return code
}

func generateErrorCheck() string {
	code := ""
	code += fmt.Sprintf("  if err != nil {\n")
	code += fmt.Sprintf("    return err\n")
	code += fmt.Sprintf("  }\n")
	return code
}

func generateErrorCheckUserReturn(retval string) string {
	code := ""
	code += fmt.Sprintf("  if err != nil {\n")
	code += fmt.Sprintf("    return %s, err\n", retval)
	code += fmt.Sprintf("  }\n")
	return code
}

// func lastPersistedMethodName(define *common.XMLDefine) string {
// 	lastName := ""
// 	for _, f := range define.Methods {
// 		if f.noPersist == true {
// 			continue
// 		}
// 		lastName = f.Name
// 	}
// 	return lastName

// }

func lastFieldName(define *common.XMLDefine) string {
	lastName := ""
	for _, f := range define.Fields {
		if f.SkipPersistance == true {
			continue
		}
		lastName = f.Name
	}
	return lastName
}

func getSchemaName(define *common.XMLDefine) string {
	return (fmt.Sprintf("DB_SCHEMA_%s", strings.ToUpper(define.Name)))
}

func generatePersistenceReadWriteVarList(define *common.XMLDefine, varName string, skipAutoID bool) string {
	code := ""

	lastName := lastFieldName(define)

	for _, f := range define.Fields {
		if f.SkipPersistance == true {
			continue
		}
		if (f.DBAutoID == true) && (skipAutoID == true) {
			continue
		}
		if strings.Compare(f.Name, lastName) != 0 {
			code += fmt.Sprintf("      %s.%s,\n", varName, f.Name)
		} else {
			code += fmt.Sprintf("      %s.%s)\n", varName, f.Name)
		}
	}

	// TODO: Shit should not work on methods..
	// for _, f := range define.Methods {
	// 	if f.NoPersist == true {
	// 		continue
	// 	}
	// 	if (f.DBAutoID == true) && (skipAutoID == true) {
	// 		continue
	// 	}
	// 	if strings.Compare(f.Name, lastName) != 0 {
	// 		code += fmt.Sprintf("      %s.%s,\n", varName, f.Name)
	// 	} else {
	// 		code += fmt.Sprintf("      %s.%s)\n", varName, f.Name)
	// 	}
	// }
	code += fmt.Sprintf("\n")
	return code
}

func generatePersistenceCreateCode(define *common.XMLDefine) string {

	code := ""

	code += fmt.Sprintf("var ErrNoSuch%s = errors.New(\"No such %s\")\n", define.Name, define.Name)
	code += fmt.Sprintf("\n")
	code += fmt.Sprintf("const createUpdateVariables%s = \"", define.Name)

	if len(define.Fields) == 0 {
		log.Println("Define has no fields!!!!")
		log.Println(define)
	}

	//primaryFieldName := strings.ToLower(define.Name) + "id"
	primaryFieldName := strings.ToLower(define.Fields[0].Name)
	primaryIsAutoID := define.Fields[0].DBAutoID

	//lastName := define.lastPersistedMethodName()

	haveAtleastOneField := false
	for _, f := range define.Fields {
		if f.SkipPersistance == true {
			continue
		}
		dbFieldName := strings.ToLower(f.Name)
		if strings.Compare(primaryFieldName, dbFieldName) != 0 {
			code += fmt.Sprintf("%s=?,", dbFieldName)
			haveAtleastOneField = true
		}
	}

	// TODO: Check if we should support this... not quite sure..
	if !haveAtleastOneField {
		log.Printf("Class: '%s' has only one field or no field!! - this won't work, set attribute 'nonpersist=\"true\"' on class to generate lagnuage definition but no persistence code.", define.Name)
	}

	code = code[:len(code)-1]
	code += fmt.Sprintf("\"\n")
	code += fmt.Sprintf("\n")

	schemaName := getSchemaName(define) //fmt.Sprintf("DB_SCHEMA_%s", strings.ToUpper(define.Name))

	methodName := fmt.Sprintf("Create%s", define.Name)
	code += fmt.Sprintf("// %s creates a record in the DB\n", methodName)
	code += fmt.Sprintf("func (p* Persistence) %s(obj *%s) error {\n", methodName, define.Name)
	if primaryIsAutoID {
		// If autoid is enabled for the primary field we remove it from insert
		code += fmt.Sprintf("  stmt, err := p.db.Prepare(\"INSERT \"+%s+\" SET \"+createUpdateVariables%s)\n", schemaName, define.Name)
	} else {
		// autoid is not enabled, so we insert
		code += fmt.Sprintf("  stmt, err := p.db.Prepare(\"INSERT \"+%s+\" SET %s=?,\"+createUpdateVariables%s)\n", schemaName, primaryFieldName, define.Name)
	}
	code += fmt.Sprintf("  _, err = stmt.Exec(\n")
	//log.Println("lastName: ", lastName)
	code += generatePersistenceReadWriteVarList(define, "obj", true)
	code += generateErrorCheck()
	code += fmt.Sprintf("  return nil\n")
	code += fmt.Sprintf("}\n")
	code += fmt.Sprintf("\n")

	return code
}

func generatePersistenceFetchCode(define *common.XMLDefine) string {
	code := ""

	if createRetrieveFuncPostfix == false {
		code += fmt.Sprintf("func (p* Persistence) fetchFromQueryString(queryString string) ([]%s, error) {\n", define.Name)
	} else {
		code += fmt.Sprintf("func (p* Persistence) fetchFromQueryString%s(queryString string) ([]%s, error) {\n", define.Name, define.Name)
	}
	code += fmt.Sprintf("  rows,err := p.db.Query(queryString)\n")
	code += generateErrorCheckUserReturn("nil")

	code += fmt.Sprintf("\n")
	code += fmt.Sprintf("  list := make([]%s,0,0)\n", define.Name)
	code += fmt.Sprintf("\n")

	code += fmt.Sprintf("  for rows.Next() {\n")
	code += fmt.Sprintf("    res := %s{}\n", define.Name)
	code += fmt.Sprintf("    err := rows.Scan(\n")
	code += generatePersistenceReadWriteVarList(define, "&res", false)
	code += generateErrorCheckUserReturn("nil")
	code += fmt.Sprintf("    list = append(list, res)\n")
	code += fmt.Sprintf("  }\n")
	code += fmt.Sprintf("  return list, nil\n")
	code += fmt.Sprintf("}\n")
	code += fmt.Sprintf("\n")

	return code
}
func generatePersistenceRetrieveCode(define *common.XMLDefine) string {
	code := ""

	//fieldname := strings.ToLower(define.Name) + "id"
	fieldname := strings.ToLower(define.Fields[0].Name)

	schemaName := getSchemaName(define) // fmt.Sprintf("DB_SCHEMA_%s", strings.ToUpper(define.Name))

	methodName := fmt.Sprintf("Retrieve%sFromID", define.Name)
	code += fmt.Sprintf("// %s Retrieves a single record in the DB matching supplied ID\n", methodName)
	code += fmt.Sprintf("// ErrNoSuch%s is returned if no record is found\n", define.Name)
	code += fmt.Sprintf("func (p *Persistence) %s(ID string) (*%s, error) {\n", methodName, define.Name)
	code += fmt.Sprintf("  queryString := fmt.Sprintf(\"SELECT * FROM %%s WHERE %s='%%s'\",%s, ID)\n", fieldname, schemaName)

	if createRetrieveFuncPostfix == false {
		code += fmt.Sprintf("  result, err := p.fetchFromQueryString(queryString)\n")
	} else {
		code += fmt.Sprintf("  result, err := p.fetchFromQueryString%s(queryString)\n", define.Name)
	}
	code += fmt.Sprintf("\n")
	code += generateErrorCheckUserReturn("nil")
	code += fmt.Sprintf("\n")
	code += fmt.Sprintf("  if len(result) == 0 {\n")
	code += fmt.Sprintf("    log.Println(\"No %s found for id: %%s\", ID)\n", define.Name)
	code += fmt.Sprintf("    return nil, ErrNoSuch%s\n", define.Name)
	code += fmt.Sprintf("  }\n")
	code += fmt.Sprintf("\n")
	code += fmt.Sprintf("  return &result[0],nil\n")
	code += fmt.Sprintf("}\n")
	code += fmt.Sprintf("\n")

	return code
}
func generatePersistenceUpdateCode(define *common.XMLDefine) string {
	code := ""
	//fieldname := strings.ToLower(define.Name) + "id"
	fieldname := strings.ToLower(define.Fields[0].Name)

	schemaName := getSchemaName(define) //fmt.Sprintf("DB_SCHEMA_%s", strings.ToUpper(define.Name))

	code += fmt.Sprintf("var updateQuery%s = \"UPDATE \" + %s + \" SET \" + createUpdateVariables%s + \" WHERE %s=?\"\n", define.Name, schemaName, define.Name, fieldname)

	methodName := fmt.Sprintf("Update%s", define.Name)
	code += fmt.Sprintf("// %s Updates the structure in the db\n", methodName)
	code += fmt.Sprintf("func (p *Persistence) %s(obj *%s) error {\n", methodName, define.Name)
	code += fmt.Sprintf("  stmt, err := p.db.Prepare(updateQuery%s)\n", define.Name)
	code += generateErrorCheck()
	code += fmt.Sprintf("  _, err = stmt.Exec(\n")

	//mainKeyField := fmt.Sprintf("%sID", define.Name)
	mainKeyField := define.Fields[0].Name

	for _, f := range define.Fields {
		if f.SkipPersistance == true {
			continue
		}
		if strings.Compare(f.Name, mainKeyField) != 0 {
			code += fmt.Sprintf("    obj.%s,\n", f.Name)
		}
	}
	code += fmt.Sprintf("    obj.%s)\n", mainKeyField)
	code += fmt.Sprintf("\n")
	code += generateErrorCheck()
	code += fmt.Sprintf("\n")
	code += fmt.Sprintf("  return nil\n")
	code += fmt.Sprintf("}\n")
	code += fmt.Sprintf("\n")
	return code
}

func generatePersistenceDeleteCode(define *common.XMLDefine) string {
	code := ""

	//fieldname := strings.ToLower(define.Name) + "id"
	fieldname := strings.ToLower(define.Fields[0].Name)
	schemaName := getSchemaName(define) //fmt.Sprintf("DB_SCHEMA_%s", strings.ToUpper(define.Name))
	mainKeyField := fmt.Sprintf("%sID", define.Name)

	code += fmt.Sprintf("var deleteQuery%s = \"DELETE FROM \" + %s + \" WHERE %s=?\"\n", define.Name, schemaName, fieldname)
	code += fmt.Sprintf("\n")
	methodName := fmt.Sprintf("Delete%s", define.Name)
	code += fmt.Sprintf("// %s Deletes the structure in the db\n", methodName)
	code += fmt.Sprintf("func (p *Persistence) %s(%s string) error {\n", methodName, mainKeyField)

	code += fmt.Sprintf("  stmt, err := p.db.Prepare(deleteQuery%s)\n", define.Name)
	code += generateErrorCheck()
	code += fmt.Sprintf("\n")
	code += fmt.Sprintf("  result, err := stmt.Exec(%s)\n", mainKeyField)
	code += generateErrorCheck()
	code += fmt.Sprintf("\n")
	code += fmt.Sprintf("  affected, _ := result.RowsAffected()\n")
	code += fmt.Sprintf("  if affected == 0 {\n")
	code += fmt.Sprintf("    return ErrNoSuch%s\n", define.Name)
	code += fmt.Sprintf("  }")
	code += fmt.Sprintf("\n")
	code += fmt.Sprintf("  return nil\n")
	code += fmt.Sprintf("}\n")
	code += fmt.Sprintf("\n")
	return code
}
