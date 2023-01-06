package mapx

import "reflect"

var JSONEncoderFuncs = RegisterEncoder(EncoderFuncs{}, jsonAny)

func jsonAny(val any) (any, error) {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.Slice:
		l := v.Len()
		out := make([]any, 0, l)
		for i := 0; i < l; i++ {
			v, err := jsonAny(v.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case reflect.Map:
		out := make(map[string]any, v.Len())
		for iter := v.MapRange(); iter.Next(); {
			k, v := iter.Key(), iter.Value()

			val, err := jsonAny(v.Interface())
			if err != nil {
				return nil, err
			}
			out[k.String()] = val
		}
		return out, nil
	}

	return val, nil
}
