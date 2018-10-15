package mysql

import "modelgenerator/common"

type CrudGenerator struct {
	Imports []common.XMLImport
}

type DBGenerator struct{}

func CreateCrudGenerator() common.Generator {
	generator := CrudGenerator{}
	return (common.Generator)(&generator)
}

func CreateDBGenerator() common.Generator {
	generator := DBGenerator{}
	return (common.Generator)(&generator)
}
