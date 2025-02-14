package exoclick

import (
	"bytes"
	"fmt"
	"reflect"
)

var dateType = reflect.TypeOf(CustomDate{})

func Stringify(message interface{}) string {
	var buf bytes.Buffer
	v := reflect.ValueOf(message)
	stringifyValue(&buf, v)
	return buf.String()
}

func stringifyValue(w *bytes.Buffer, val reflect.Value) {
	if val.Kind() == reflect.Ptr && val.IsNil() {
		w.WriteString("<nil>")
		return
	}

	v := reflect.Indirect(val)

	switch v.Kind() {
	case reflect.String:
		fmt.Fprintf(w, `"%s"`, v)
	case reflect.Slice:
		w.WriteByte('[')
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				w.WriteByte(' ')
			}

			stringifyValue(w, v.Index(i))
		}

		w.WriteByte(']')
		return
	case reflect.Struct:
		if v.Type().Name() != "" {
			w.WriteString(v.Type().String())
		}

		if v.Type() == dateType {
			fmt.Fprintf(w, "{%s}", v.Interface())
			return
		}

		w.WriteByte('{')

		var sep bool
		for i := 0; i < v.NumField(); i++ {
			fv := v.Field(i)
			if fv.Kind() == reflect.Ptr && fv.IsNil() {
				continue
			}
			if fv.Kind() == reflect.Slice && fv.IsNil() {
				continue
			}
			if fv.Kind() == reflect.Map && fv.IsNil() {
				continue
			}

			if sep {
				w.WriteString(", ")
			} else {
				sep = true
			}

			w.WriteString(v.Type().Field(i).Name)
			w.WriteByte(':')
			stringifyValue(w, fv)
		}

		w.WriteByte('}')
	default:
		if v.CanInterface() {
			fmt.Fprint(w, v.Interface())
		}
	}
}
