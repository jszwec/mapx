package mapx

import (
	"reflect"
	"strings"
	"time"
)

type tag struct {
	name      string
	prefix    string
	tagname   string
	empty     bool
	omitEmpty bool
	ignore    bool
	inline    bool
	raw       bool
}

func parseTag(tagname string, field reflect.StructField) (t tag) {
	t.tagname = tagname
	t.raw = isKnownStruct(walkType(field.Type))

	tags := strings.Split(field.Tag.Get(tagname), ",")
	if len(tags) == 1 && tags[0] == "" {
		t.name = field.Name
		t.empty = true
		return
	}

	switch tags[0] {
	case "-":
		t.ignore = true
		return
	case "":
		t.name = field.Name
	default:
		t.name = tags[0]
	}

	for _, tagOpt := range tags[1:] {
		switch tagOpt {
		case "omitempty":
			t.omitEmpty = true
		case "inline":
			if walkType(field.Type).Kind() == reflect.Struct {
				t.inline = true
				t.prefix = tags[0]
			}
		case "raw":
			t.raw = true
		}
	}
	return
}

func walkType(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ
}

var timeType = reflect.TypeOf((*time.Time)(nil)).Elem()

func isKnownStruct(typ reflect.Type) bool {
	// we will possibly list more here.
	return timeType == typ
}
