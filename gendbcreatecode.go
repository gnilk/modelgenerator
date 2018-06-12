package main

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
	"strings"
)

func generateDBCreateCode(doc XMLDoc, options *Options) string {
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
					code += doc.Defines[i].generateDBCreateCode(options)

					//code += doc.Defines[i].generatePersistenceCode(options.AllPersistenceClasses[j], options.Converters)
				}
			} else {
				code += doc.Defines[i].generateDBCreateCode(options)
				//code += doc.Defines[i].generatePersistenceCode(options.PersistenceClass, options.Converters)
			}

		}
	} else {
		log.Panicln("SPLIT IN FILES NOT SUPPORTED!!!!")
	}

	return code
}

func generateDBCreateHeader(doc XMLDoc, source string) string {
	code := ""
	if len(doc.DBControl.DBName) > 0 {
		code = fmt.Sprintf("USE `%s`;\n", doc.DBControl.DBName)
	} else {
		code = "USE `nagini`;\n"
	}

	return code
}

func (field *XMLDataTypeField) getDBColumnName(options *Options) string {
	return fmt.Sprintf("%s", strings.ToLower(field.Name))

}

func (field *XMLDataTypeField) additionalDBCreateStatement(options *Options) string {
	res := ""

	if field.DBAutoID {
		res = res + " AUTO_INCREMENT"
	}
	return res
}

func (define *XMLDefine) getDBTableName(options *Options) string {
	return fmt.Sprintf("%s%s", options.DBTablePrefix, strings.ToLower(define.Name))
}

func (define *XMLDefine) generateDBCreateCode(options *Options) string {
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
		code += define.generateDBCreateCodeForClass(options)
		break
	default:
		fmt.Printf("Error, can't generate code for type '%s'\n", define.Type)
		break
	}

	return code
}

func (define *XMLDefine) generateDBCreateCodeForClass(options *Options) string {
	code := "\n"

	if options.GenerateDropStatement == true {
		code += fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", define.getDBTableName(options))
	}

	if options.IsUpgrade != true {
		code += fmt.Sprintf("CREATE TABLE `%s` (\n", define.getDBTableName(options))
	}

	code += define.generateDBFieldCode(options)
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

func (define *XMLDefine) generateDBFieldCode(options *Options) string {
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
					define.getDBTableName(options),
					field.getDBColumnName(options),
					//field.getDBType(options),
					field.TypeMapping(options.CurrentDoc.DBTypeMappings),
					defaultValue)
			}
		} else {
			if firstField {
				code += fmt.Sprintf("  `%s` %s NOT NULL %s,\n",
					field.getDBColumnName(options),
					field.TypeMapping(options.CurrentDoc.DBTypeMappings),
					field.additionalDBCreateStatement(options))
				firstField = false
			} else {
				code += fmt.Sprintf("  `%s` %s NOT NULL %s,\n",
					field.getDBColumnName(options),
					field.TypeMapping(options.CurrentDoc.DBTypeMappings),
					field.additionalDBCreateStatement(options))
			}
		}
	}
	return code
}

func (define *XMLDefine) generateDBCreateCodeForField(options *Options, list []XMLDataTypeField, typeString string, isPrimary bool) string {
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
					define.getDBTableName(options),
					field.getDBColumnName(options),
					typeString,
					defaultValue)
			}
		} else {
			if firstField {
				code += fmt.Sprintf("  `%s` %s NOT NULL,\n", field.getDBColumnName(options), typeString)
				firstField = false
			} else {
				code += fmt.Sprintf("  `%s` %s DEFAULT NULL,\n", field.getDBColumnName(options), typeString)
			}
		}
	}
	return code
}

func (define *XMLDefine) generateDBCreateCodeForStrings(options *Options, list []XMLDataTypeField) string {
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
					define.getDBTableName(options),
					field.getDBColumnName(options),
					fieldSize,
					defaultValue)
			}

		} else {
			code += fmt.Sprintf("  `%s` varchar(%d) DEFAULT NULL,\n",
				field.getDBColumnName(options),
				fieldSize)
		}

	}

	return code
}

func (define *XMLDefine) generateDBCreateCodeForFieldWithoutNULL(list []XMLDataTypeField, typeString string) string {
	code := ""
	for _, field := range list {
		name := strings.ToLower(field.Name)
		code += fmt.Sprintf("  `%s` %s,\n", name, typeString)
	}
	return code
}
