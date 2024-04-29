package codegen

import "fmt"

// dict accepts a series of arguments, and from those arguments creates a string-keyed map.
// the first argument is the key, and the second argument is the value, repeating until no more arguments are left.
// when an odd number of arguments are provided, the value is set as empty string.
func dict(v ...any) map[string]any {
	dict := map[string]any{}
	lenv := len(v)
	for i := 0; i < lenv; i += 2 {
		key := strval(v[i])
		if i+1 >= lenv {
			dict[key] = ""
			continue
		}
		dict[key] = v[i+1]
	}
	return dict
}

// dictGet returns the value in map d corresponding to the key key.
// if no value exists, an empty string is returned
func dictGet(d map[string]any, key string) any {
	if val, ok := d[key]; ok {
		return val
	}
	return ""
}

// dictSet takes in a dict, a string key, and a value, and sets the value at that key in the dict to the value argument
func dictSet(d map[string]any, key string, value any) map[string]any {
	d[key] = value
	return d
}

// dictSet takes in a dict and a string key, and removes the value at that key in the dict
func dictUnset(d map[string]any, key string) map[string]any {
	delete(d, key)
	return d
}

// dictHasKey takes in a dict and a key and returns if a value exists at that key in the dict
func dictHasKey(d map[string]any, key string) bool {
	_, ok := d[key]
	return ok
}

// dictPluck takes in a key and a series of dicts and returns any values that are in the provided dicts at that key
func dictPluck(key string, d ...map[string]any) []any {
	res := []any{}
	for _, dict := range d {
		if val, ok := dict[key]; ok {
			res = append(res, val)
		}
	}
	return res
}

// dictKeys takes in a series of dicts and returns all keys in those dicts.
// keys can be repeated if they are in multiple dicts.
func dictKeys(dicts ...map[string]any) []string {
	k := []string{}
	for _, dict := range dicts {
		for key := range dict {
			k = append(k, key)
		}
	}
	return k
}

// dictValues takes in a series of dicts and returns all values in those dicts.
func dictValues(dicts ...map[string]any) []any {
	values := []any{}
	for _, dict := range dicts {
		for _, value := range dict {
			values = append(values, value)
		}
	}

	return values
}

// dictPick takes in a dict and a series of string keys.
// It returns a new dict containing keys & values corresponding to the provided keys
func dictPick(dict map[string]any, keys ...string) map[string]any {
	res := map[string]any{}
	for _, k := range keys {
		if v, ok := dict[k]; ok {
			res[k] = v
		}
	}
	return res
}

// dictOmit takes in a dict and a series of string keys.
// It returns a new dict containing all keys & values excluding those corresponding to the provided keys
func dictOmit(dict map[string]any, keys ...string) map[string]any {
	res := map[string]any{}

	omit := make(map[string]bool, len(keys))
	for _, k := range keys {
		omit[k] = true
	}

	for k, v := range dict {
		if _, ok := omit[k]; !ok {
			res[k] = v
		}
	}
	return res
}

// dictDig takes in a series of keys, a default value, and a dict.
// From those arguments, it returns the value corresponding to the recursive lookup of the dict.
// in the situation where no value is found at the key, the default is returned
func dictDig(ps ...any) (any, error) {
	if len(ps) < 3 {
		return nil, fmt.Errorf("dig needs at least three arguments")
	}
	dict := ps[len(ps)-1].(map[string]any)
	def := ps[len(ps)-2]
	ks := make([]string, len(ps)-2)
	for i := 0; i < len(ks); i++ {
		ks[i] = ps[i].(string)
	}

	return digFromDict(dict, def, ks)
}

// helper for dictDig that performs the recursive search
func digFromDict(dict map[string]any, d any, ks []string) (any, error) {
	k, ns := ks[0], ks[1:]
	step, has := dict[k]
	if !has {
		return d, nil
	}
	if len(ns) == 0 {
		return step, nil
	}
	dat, ok := step.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("cannot dig type %T", step)
	}
	return digFromDict(dat, d, ns)
}

// helper to convert any type to a string
func strval(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
