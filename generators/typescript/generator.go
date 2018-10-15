package typescript

import "modelgenerator/common"

type CodeGenerator struct {
	Methods []common.AccessMethod
	Imports []common.XMLImport
}

type TSLangGenerators struct{}

func CreateTSLanguage() common.Language {
	tslang := TSLangGenerators{}
	return (common.Language)(&tslang)
}

func (lang *TSLangGenerators) GetModelGenerator() common.Generator {
	return createTSLangGenerator()
}

func (lang *TSLangGenerators) GetCrudGenerator() common.Generator {
	return nil
}

func (lang *TSLangGenerators) GetDBCreateGenerator() common.Generator {
	return nil
}

func createTSLangGenerator() common.Generator {
	codeGen := CodeGenerator{}
	return (common.Generator)(&codeGen)
}
