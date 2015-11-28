package queryutil

import (
	"reflect"
	"strconv"
)

type elementNamer struct {
	prefix string
}

func newElementNamer(prefix string) elementNamer {
	return elementNamer{prefix: prefix}
}

func (n elementNamer) Name(i int) string {
	if n.prefix == "" {
		return strconv.Itoa(i + 1)
	} else {
		return n.prefix + "." + strconv.Itoa(i+1)
	}
}

type mapNamer struct {
	elementNamer elementNamer
	kname, vname string
}

func (n mapNamer) name(i int, finalName string) string {
	return n.elementNamer.Name(i) + "." + finalName
}

func (n mapNamer) KeyName(i int) string {
	return n.name(i, n.kname)
}

func (n mapNamer) ValueName(i int) string {
	return n.name(i, n.vname)
}

func newMapNamer(prefix string, tag reflect.StructTag) mapNamer {
	kname := tag.Get("locationNameKey")
	if kname == "" {
		kname = "key"
	}
	vname := tag.Get("locationNameValue")
	if vname == "" {
		vname = "value"
	}
	return mapNamer{kname: kname, vname: vname, elementNamer: newElementNamer(prefix)}
}
