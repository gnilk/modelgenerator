package golang

//
// Generates persistence base code for the POGO class
// Simple CRUD functions - you have to extend your self
//
//
// 	"io/ioutil"
//
import (
	"fmt"
	"log"
	"modelgenerator/common"
	"strings"
)

func (generator *DBGenerator) GenerateCode(doc common.XMLDoc, options *common.Options) string {
	code := ""

	// className string, source string, splitInFiles bool, converters bool, verbose int, outputDir string

	if options.SplitInFiles != true {
		code += generateDBCreateHeader(doc, options.Filename)
		// generate code for all defines
		for i := 0; i < len(doc.Defines); i++ {
			if doc.Defines[i].SkipPersistance == true {
				continue
			}
			if strings.Compare(options.PersistenceClass, "-") != 0 {
				for j := 0; j < len(options.AllPersistenceClasses); j++ {
					options.PersistenceClass = options.AllPersistenceClasses[j]
					code += generateDBCreateCodeForDefine(&doc.Defines[i], options)

					//code += doc.Defines[i].generatePersistenceCode(options.AllPersistenceClasses[j], options.Converters)
				}
			} else {
				code += generateDBCreateCodeForDefine(&doc.Defines[i], options)
				//code += doc.Defines[i].generatePersistenceCode(options.PersistenceClass, options.Converters)
			}

		}
	} else {
		log.Panicln("SPLIT IN FILES NOT SUPPORTED!!!!")
	}

	return code
}

func generateDBCreateHeader(doc common.XMLDoc, source string) string {
	code := ""
	if len(doc.DBControl.DBName) > 0 {
		code = fmt.Sprintf("USE `%s`;\n", doc.DBControl.DBName)
	} else {
		code = "USE `nagini`;\n"
	}

	return code
}

func getDBTableName(define *common.XMLDefine, options *common.Options) string {
	return fmt.Sprintf("%s%s", options.DBTablePrefix, strings.ToLower(define.Name))
}

func generateDBCreateCodeForDefine(define *common.XMLDefine, options *common.Options) string {
	//options.PersistenceClass, options.Converters)
	// Check if class name matches - perhaps use regexp here..
	if strings.Compare(options.PersistenceClass, "-") != 0 {
		if strings.Compare(options.PersistenceClass, define.Name) != 0 {
			return ""
		}
	}

	if options.Verbose > 0 {
		log.Printf("Generating DB Create Statements for class: %s\n", define.Name)
	}

	code := ""

	switch define.Type {
	case "class":
		code += generateDBCreateCodeForClass(define, options)
		break
	case "enum":
		// silent skip enum type - this is not an error, we just don't put them in the DB
		break
	default:
		fmt.Printf("Error, can't generate code for type '%s'\n", define.Type)
		break
	}

	return code
}

func generateDBCreateCodeForClass(define *common.XMLDefine, options *common.Options) string {
	code := "\n"

	if options.GenerateDropStatement == true {
		code += fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", getDBTableName(define, options))
	}

	if options.IsUpgrade != true {
		code += fmt.Sprintf("CREATE TABLE `%s` (\n", getDBTableName(define, options))
	}

	code += generateDBFieldCode(define, options)
	// code += define.generateDBCreateCodeForField(options, define.Guids, "varchar(36)", true)
	// code += define.generateDBCreateCodeForStrings(options, define.Strings)
	// code += define.generateDBCreateCodeForField(options, define.Ints, "int(11)", false)
	// code += define.generateDBCreateCodeForField(options, define.Bools, "int(11)", false)
	// code += define.generateDBCreateCodeForField(options, define.Times, "datetime", false)
	// code += define.generateDBCreateCodeForField(options, define.Lists, "mediumblob", false) // This needs a special thing
	// code += define.generateDBCreateCodeForField(options, define.Enums, "int(11)", false)

	// When not upgrading we need to close table creation statement
	if !options.IsUpgrade {
		// Insert primary key - this defaults to first GUID - could add XML attribute to class in order to define this
		primaryKey := define.Fields[0]
		code += fmt.Sprintf("  PRIMARY KEY(`%s`)\n", strings.ToLower(primaryKey.Name))
		code += fmt.Sprintf(") ENGINE=InnoDB DEFAULT CHARSET=utf8;\n")
	}

	return code
}

func generateDBFieldCode(define *common.XMLDefine, options *common.Options) string {
	code := ""
	firstField := true
	for _, field := range define.Fields {
		if field.SkipPersistance == true {
			continue
		}
		if options.IsUpgrade {
			if field.FromVersion >= options.FromVersion {
				defaultValue := field.Default
				if len(defaultValue) == 0 {
					// Ok with empty strings
					log.Printf("!WARNING!: Upgrade require field default values, check definition of '%s::%s'\n", define.Name, field.Name)
				}
				code += fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s NOT NULL DEFAULT '%s';\n",
					getDBTableName(define, options),
					field.GetDBColumnName(options),
					//field.getDBType(options),
					field.TypeMapping(options.CurrentDoc.DBTypeMappings),
					defaultValue)
			}
		} else {
			if firstField {
				code += fmt.Sprintf("  `%s` %s NOT NULL %s,\n",
					field.GetDBColumnName(options),
					field.TypeMapping(options.CurrentDoc.DBTypeMappings),
					field.AdditionalDBCreateStatement(options))
				firstField = false
			} else {
				code += fmt.Sprintf("  `%s` %s NOT NULL %s,\n",
					field.GetDBColumnName(options),
					field.TypeMapping(options.CurrentDoc.DBTypeMappings),
					field.AdditionalDBCreateStatement(options))
			}
		}
	}
	return code
}

func generateDBCreateCodeForField(define *common.XMLDefine, options *common.Options, list []common.XMLDataTypeField, typeString string, isPrimary bool) string {
	code := ""

	firstField := isPrimary
	for _, field := range list {
		if options.IsUpgrade {
			if field.FromVersion >= options.FromVersion {
				defaultValue := field.Default
				if len(defaultValue) == 0 {
					log.Fatalf("!Error: Upgrade require field default values!\nCheck definition of '%s::%s`\n", define.Name, field.Name)
				}
				code += fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s NULL DEFAULT %s;\n",
					getDBTableName(define, options),
					field.GetDBColumnName(options),
					typeString,
					defaultValue)
			}
		} else {
			if firstField {
				code += fmt.Sprintf("  `%s` %s NOT NULL,\n", field.GetDBColumnName(options), typeString)
				firstField = false
			} else {
				code += fmt.Sprintf("  `%s` %s DEFAULT NULL,\n", field.GetDBColumnName(options), typeString)
			}
		}
	}
	return code
}

func generateDBCreateCodeForStrings(define *common.XMLDefine, options *common.Options, list []common.XMLDataTypeField) string {
	code := ""

	for _, field := range list {
		fieldSize := field.DBSize
		//fmt.Printf("FieldSize: %d\n",fieldSize)
		if fieldSize == 0 {
			fieldSize = 128 // default field size, make this configurable option
		}
		if options.IsUpgrade {
			if field.FromVersion >= options.FromVersion {
				defaultValue := field.Default
				if len(defaultValue) == 0 {
					log.Fatalf("Upgrade require field default values!\nCheck definition of '%s::%s`\n", define.Name, field.Name)
				}

				code += fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` varchar(%d) NULL DEFAULT `%s`",
					getDBTableName(define, options),
					field.GetDBColumnName(options),
					fieldSize,
					defaultValue)
			}

		} else {
			code += fmt.Sprintf("  `%s` varchar(%d) DEFAULT NULL,\n",
				field.GetDBColumnName(options),
				fieldSize)
		}

	}

	return code
}

func generateDBCreateCodeForFieldWithoutNULL(define *common.XMLDefine, list []common.XMLDataTypeField, typeString string) string {
	code := ""
	for _, field := range list {
		name := strings.ToLower(field.Name)
		code += fmt.Sprintf("  `%s` %s,\n", name, typeString)
	}
	return code
}
