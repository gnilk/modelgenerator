package common

import (
	"fmt"
	"strings"
)

func (field *XMLDataTypeField) dump() {
	fmt.Printf("  Name: %s\n", field.Name)
	fmt.Printf("  Default: %s\n", field.Default)
}

//TypeMapping maps a field's type to potential mappings from the specific table
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

func (field *XMLDataTypeField) TypeMappingLang(typeMappings []XMLTypeMapping, lang string) string {
	for _, mapping := range typeMappings {
		if strings.Compare(mapping.Lang, lang) == 0 {
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
	}
	return field.Type
}

//
// TODO: These should be moved out of here
//
func (field *XMLDataTypeField) GetDBColumnName(options *Options) string {
	return fmt.Sprintf("%s", strings.ToLower(field.Name))

}

func (field *XMLDataTypeField) AdditionalDBCreateStatement(options *Options) string {
	res := ""

	if field.DBAutoID {
		res = res + " AUTO_INCREMENT"
	}
	return res
}
