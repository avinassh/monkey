package object

type ObjectType string

const (
	NULL_OBJ         = "NULL"
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}
