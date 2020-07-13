package phppack

import (
	"reflect"
)

type packTag struct {
	Type string
	Size int
}

type packType struct {
	Name string
	Type reflect.Type
	tag  packTag
}
