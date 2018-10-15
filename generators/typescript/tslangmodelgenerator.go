package typescript

import (
	"fmt"
	"log"
	"modelgenerator/common"
)

func (generator *CodeGenerator) GenerateCode(doc common.XMLDoc, options *common.Options) string {
	code := ""
	if options.SplitInFiles == true {
		log.Printf("Split In Files not supported!\n")
		return code
	}

	code += fmt.Sprintf("//\n")
	code += fmt.Sprintf("// This file has been generated by ModelGenerator - do NOT edit!\n")
	code += fmt.Sprintf("//\n")

	for _, define := range doc.Defines {
		//log.Printf("Generate for define: %s\n", define.Name)
		code += generator.generateHeaderCodeForDefine(&define, options)
		//		code += generator.generateCode(&define, options)
	}

	return code
}

func (generator *CodeGenerator) generateHeaderCodeForDefine(define *common.XMLDefine, options *common.Options) string {
	code := ""

	log.Printf("Generating code for type='%s', named='%s'\n", define.Type, define.Name)
	switch define.Type {
	case "class":
		code += generator.generateClassCodeDefinition(define, options)
		break
	case "enum":
		code += generator.generateEnumCodeDefinition(define, options)
		break
	default:
		fmt.Printf("[TSLangModelGenerator::generateCode] Error, can't generate code for type '%s'\n", define.Type)
		break
	}

	return code

}

func (generator *CodeGenerator) generateClassCodeDefinition(define *common.XMLDefine, options *common.Options) string {
	code := ""
	// Begin class header
	code += fmt.Sprintf("class %s", define.Name)
	if define.Inherits != "" {
		code += fmt.Sprintf(" : public %s", define.Inherits)
	}
	code += fmt.Sprintf(" {\n") // end class header
	// fields
	code += fmt.Sprintf("public:\n")

	for _, field := range define.Fields {
		code += generator.fieldCode(options, &field)
		//generator.methodFromField(define, field, field.TypeMapping(options.CurrentDoc.GOTypeMappings), field.IsList)
	}

	code += fmt.Sprintf("};\n\n")
	return code
}

func (generator *CodeGenerator) fieldCode(options *common.Options, field *common.XMLDataTypeField) string {
	code := ""

	typePrefix := ""
	if field.IsList {
		typePrefix = typePrefix + "[]"
	}
	if field.IsPointer {
		typePrefix = typePrefix + "*"
	}
	code += fmt.Sprintf("    %s %s%s;\n", field.TypeMappingLang(options.CurrentDoc.AnyTypeMappings, "ts"), typePrefix, field.Name)

	return code
}

func (generator *CodeGenerator) generateEnumCodeDefinition(define *common.XMLDefine, options *common.Options) string {
	code := ""
	code += fmt.Sprintf("typedef enum {\n")
	for _, Int := range define.Ints {
		code += fmt.Sprintf("    %s = %d,\n", Int.Name, Int.Value)
	}
	code += fmt.Sprintf("} %s;\n\n", define.Name)
	return code
}

// func (generator *CodeGenerator) generateCode(options *common.Options, define *common.XMLDefine) string {
// 	return ""
// }
