package evaluator

import (
	"nala/object"
)

type MapofIDtoBuiltin map[string]*object.BuiltIn

// TODOs:
// Iterable interface: Array, String, Vector

// export builtins to REPL
var builtins = MapofIDtoBuiltin{
	"len":    object.GetBuiltinByName("len"),
	"type":   object.GetBuiltinByName("type"),
	"first":  object.GetBuiltinByName("first"),
	"last":   object.GetBuiltinByName("last"),
	"rest":   object.GetBuiltinByName("rest"),
	"push":   object.GetBuiltinByName("push"),
	"puts":   object.GetBuiltinByName("puts"),
	"putl":   object.GetBuiltinByName("putl"),
	"reads":  object.GetBuiltinByName("reads"),
	"keys":   object.GetBuiltinByName("keys"),
	"values": object.GetBuiltinByName("values"),
	"items":  object.GetBuiltinByName("items"),
	"ins":    object.GetBuiltinByName("ins"),
	"del":    object.GetBuiltinByName("del"),
	"copy":   object.GetBuiltinByName("copy"),
	"sb":     &object.BuiltIn{Fn: nil},
	"desc":   object.GetBuiltinByName("desc"),
	// "loadf":  &object.BuiltIn{Fn: nala_loadf},
}
