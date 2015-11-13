package dispatch

import (
	"encoding/json"
	"errors"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"reflect"
	"strconv"
	"time"
)

type ReadWriteStringBuffer string

func (to ReadWriteStringBuffer) Write(byte []byte) (n int, err error) {
	to = ReadWriteStringBuffer(byte)
	return len(to), err
}

func (from ReadWriteStringBuffer) Read(out []byte) (n int, err error) {
	out = []byte(from)
	return len(from), err
}

func ToString(k interface{}) (s string, e error) {
	in := k
	typeinfo := reflect.TypeOf(in)
	kind := typeinfo.Kind()

	if *Debug {
		Info.Printf("in %v TypeOf( %v ) %v\n", in, typeinfo, kind)
	}

	switch k.(type) {
	case string:
		s = k.(string)
	case int32, int64, int:
		s = strconv.FormatInt(k.(int64), 10)
	case uint32, uint64, uint:
		s = strconv.FormatUint(k.(uint64), 10)
	case bool:
		s = strconv.FormatBool(k.(bool))
	case float32:
		s = strconv.FormatFloat(k.(float64), 'f', -1, 32)
	case float64:
		s = strconv.FormatFloat(k.(float64), 'f', -1, 32)
	case interface{}:
		s = fmt.Sprintf("%v", k)
	default:
		text := fmt.Sprintf("Expected one of [string, {u,}int{,32,64} float{32,64}] but got: %T", k)
		e = errors.New(text)
		if *Debug {
			Elog.Printf("%s\n", text)
		}
	}
	return s, e
}

func Struct2Map(in interface{}) (map[string]interface{}, error) {
	typeinfo := reflect.TypeOf(in)
	kind := typeinfo.Kind()

	if *Debug {
		Info.Printf("in %v TypeOf( %v ) %v\n", in, typeinfo, kind)
	}

	o := make(map[string]interface{})
	var err error
	var value interface{}
	var name string
	s := reflect.TypeOf(in)
	v := reflect.ValueOf(in)
	// typeOfT := s.Type()
	for i := 0; i < v.NumField(); i++ {
		field := s.Field(i)
		name = field.Name
		if *Debug {
			Info.Printf("%d: %s\n", i, name)
		}
		defer func() {
			if r := recover(); r != nil {
				if *Debug {
					Info.Printf("%d: %s %s = %v %T\n", i, name, field.Type, v.Field(i).Interface(), err)
					Info.Printf("Name[%s] %v\n", name, field)
				}
			}
		}()
		// value, err = ToString(v.Field(i).Interface())
		if *Debug {
			Info.Printf("%d: %s %s = %v %v %T %v\n", i, name, field.Type, v.Field(i).Interface(), value, value, err)
		}

		switch field.Type {
		case reflect.TypeOf(time.Time{}),
			reflect.TypeOf(time.Location{}),
			reflect.TypeOf(time.Duration(0)),
			reflect.TypeOf(time.Weekday(0)),
			reflect.TypeOf(""),
			reflect.TypeOf(false),
			reflect.TypeOf(uint32(0)),
			reflect.TypeOf(uint64(0)),
			reflect.TypeOf(uint(0)),
			reflect.TypeOf(int(0)),
			reflect.TypeOf(int32(0)),
			reflect.TypeOf(int64(0)),
			reflect.TypeOf(float64(0)),
			reflect.TypeOf(float32(0)):
			value, err = ToString(v.Field(i).Interface())
		default:
			value, err = Mapify(v.Field(i).Interface())
		}

		o[name] = value
	}
	// o := make(map[string]interface{})
	// T := reflect.TypeOf(in)
	// var err error
	// var value interface{}
	// s := reflect.ValueOf(&in).Elem()
	// typeOfT := s.Type()
	// for i := 0; i < s.NumField(); i++ {
	// 	f := s.Field(i)
	// 	value = f.Interface()
	// 	Info.Printf("%d: %s %s = %v   %T %v %v\n", i,
	// 		typeOfT.Field(i).Name, f.Type(), f.Interface(), T, T, value)
	// 	o[name] = value
	// }
	// // Info.Printf("in %v TypeOf( %v )\n", in, reflect.TypeOf(in))
	return o, err
}

func _Mapify(in interface{}) (interface{}, error) {
	typeinfo := reflect.TypeOf(in)
	kind := typeinfo.Kind()

	if *Debug {
		Info.Printf("in %v TypeOf( %v ) %v\n", in, typeinfo, kind)
	}

	var err error
	switch kind.String() {
	case "struct":
		switch in.(type) {
		case time.Time, time.Duration, time.Weekday, time.Location:
			return ToString(in)
		}
		return Struct2Map(in)
	case "slice":
		var lhs interface{}
		var err error
		var slc []interface{}
		for _, ifc := range in.([]interface{}) {
			lhs, err = Mapify(ifc)
			if err != nil {
				slc = append(slc, lhs)
			}
		}
		return slc, err
	default:

		switch in.(type) {
		case string, bool, uint32, uint64, uint, int, int32, int64, float64, float32, time.Time:
			return in, nil
		case []struct{}:
			var T []map[string]interface{}
			for _, s := range in.([]struct{}) {
				var m map[string]interface{}
				m, err = Struct2Map(s)
				T = append(T, m)
			}
			return T, nil
		case struct{}:
			return Struct2Map(in)
		case map[interface{}]interface{}:
			o := make(map[string]interface{})
			for k, v := range in.(map[interface{}]interface{}) {
				sk := ""
				switch k.(type) {
				case *interface{}:
					if in != nil {
						sk = fmt.Sprintf("%v", in)
					}
				default:
					sk, err = ToString(k)
				}
				v, err = Mapify(v)
				if err != nil {
					return nil, err
				}
				o[sk] = v
			}
			return o, nil
		case *interface{}:
			sk := ""
			if in != nil {
				sk = fmt.Sprintf("%v", in)
			}
			return sk, nil
		case []interface{}:
			in1 := in.([]interface{})
			len1 := len(in1)
			o := make([]interface{}, len1)
			for i := 0; i < len1; i++ {
				o[i], err = Mapify(in1[i])
				if err != nil {
					return nil, err
				}
			}
			return o, nil
		default:
			if *Debug {
				text := fmt.Sprintf("Expected map got %T", in)
				Info.Printf("\n%s\n", text)
			}
			return in, nil
		}
	}
	if *Debug {
		text := fmt.Sprintf("Expected map got %T", in)
		Info.Printf("\n%s\n", text)
	}
	return in, nil
}

// func Mapify(in interface{}) (map[string]interface{}, error) {
func Mapify(in interface{}) (interface{}, error) {

	typeinfo := reflect.TypeOf(in)
	kind := typeinfo.Kind()

	if *Debug {
		Info.Printf("in %v TypeOf( %v ) %v\n", in, typeinfo, kind)
	}

	switch kind.String() {
	case "struct":
		return Struct2Map(in)
	case "slice":
		var lhs interface{}
		var err error
		var slc []interface{}
		for _, ifc := range in.([]interface{}) {
			lhs, err = Mapify(ifc)
			if err != nil {
				slc = append(slc, lhs)
			}
		}
		return slc, err
	default:
		switch in.(type) {
		case map[string]interface{}:
			o := make(map[string]interface{})
			for k, v := range in.(map[string]interface{}) {
				o[k], _ = _Mapify(v)
			}
			return o, nil
		case map[interface{}]interface{}:
			o := make(map[string]interface{})
			for k, v := range in.(map[interface{}]interface{}) {
				if sk, err := ToString(k); err == nil {
					v, err = _Mapify(v)
					if err != nil {
						return nil, err
					}
					o[sk] = v
				} else {
					return nil, err
				}
			}
			return o, nil
		case []interface{}:
			o := make(map[string]interface{})
			for k, v := range in.([]interface{}) {
				if sk, err := ToString(k); err == nil {
					v, err = _Mapify(v)
					if err != nil {
						return nil, err
					}
					o[sk] = v
				} else {
					return nil, err
				}
			}
			return o, nil
		}
	}

	text := fmt.Sprintf("Expected map got %T", in)
	if *Debug {
		Elog.Printf("\n\n%s\n\n", text)
	}
	return nil, errors.New(text)
}

func MapRecursePrint(m interface{}) {
	var depth int
	var nl bool = true
	_MapRecursePrint(m, &nl, &depth)
}

func _MapRecursePrint(m interface{}, nl *bool, depth *int) {
	in := m
	typeinfo := reflect.TypeOf(in)
	kind := typeinfo.Kind()

	if *Debug {
		Info.Printf("in %v TypeOf( %v ) %v\n", in, typeinfo, kind)
	}

	switch m.(type) {
	case bool, int32, float32, float64, int64, int, string, uint32, uint64, uint, time.Time:
		if *depth > 0 {
			spaces := fmt.Sprintf("%*s", *depth*2, " ")
			Plain.Printf("%s%-v ", spaces, m)
		} else {
			Plain.Printf("%-v ", m)
		}
	case map[string]interface{}:
		for k, v := range m.(map[string]interface{}) {
			if *depth > 0 {
				spaces := fmt.Sprintf("%*s", *depth*2, " ")
				Plain.Printf("%s%-v ", spaces, k)
			} else {
				Plain.Printf("%-v ", k)
			}
			*nl = false
			__MapRecursePrint(v.(interface{}), nl, depth)
		}
	case []interface{}:
		for _, o := range m.([]interface{}) {
			__MapRecursePrint(o, nl, depth)
		}
	default:
		if *depth > 0 {
			spaces := fmt.Sprintf("%*s", *depth*2, " ")
			Plain.Printf("%s%-v ", spaces, m)
		} else {
			Plain.Printf("%-v ", m)
		}
	}
	if *depth == 0 {
		Plain.Println()
		*nl = true
	}
}

func __MapRecursePrint(m interface{}, nl *bool, depth *int) {
	in := m
	typeinfo := reflect.TypeOf(in)
	kind := typeinfo.Kind()

	if *Debug {
		Info.Printf("in %v TypeOf( %v ) %v\n", in, typeinfo, kind)
	}

	*depth++
	switch m.(type) {
	case bool, int32, float32, float64, int64, int, string, uint32, uint64, uint, time.Time:
		if *depth > 0 {
			spaces := fmt.Sprintf("%*s", *depth*2, " ")
			Plain.Printf("%s%-v ", spaces, m)
		} else {
			Plain.Printf("%-v ", m)
		}
	case []interface{}:
		_MapRecursePrint(m, nl, depth)
	case map[string]interface{}:
		_MapRecursePrint(m, nl, depth)
	default:
		if *depth > 0 {
			spaces := fmt.Sprintf("%*s", *depth*2, " ")
			Plain.Printf("%s%-v ", spaces, m)
		} else {
			Plain.Printf("%-v ", m)
		}
	}
	*depth--
}

var spaces string = fmt.Sprintf("%*s", 2, " ")

func TransformData(in interface{}) (out interface{}, err error) {
	switch in.(type) {
	case *interface{}:
		sk := ""
		if in != nil {
			sk = fmt.Sprintf("%v", in)
		}
		return sk, nil
	case map[string]string:
		return in.(interface{}), nil
	case map[interface{}]interface{}:
		o := make(map[string]interface{})
		for k, v := range in.(map[interface{}]interface{}) {
			sk, err := ToString(k)
			if err != nil {
				return nil, err
			}
			v, err = TransformData(v)
			if err != nil {
				return nil, err
			}
			o[sk] = v
		}
		return o, nil
	case []interface{}:
		in1 := in.([]interface{})
		len1 := len(in1)
		o := make([]interface{}, len1)
		for i := 0; i < len1; i++ {
			o[i], err = TransformData(in1[i])
			if err != nil {
				return nil, err
			}
		}
		return o, nil
	default:
		return in, nil
	}
	return in, nil
}

func Yamlify(data interface{}) string {
	data, err := TransformData(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	s, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return string(s)
}

func Jsonify(data interface{}) string {
	var err error
	data, err = TransformData(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	s, err := json.MarshalIndent(data, "", spaces)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return string(s)
}

func Json2Yaml(input []byte) string {
	var data interface{}
	var err = json.Unmarshal(input, &data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	data, err = TransformData(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}

	output, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return string(output)
}
