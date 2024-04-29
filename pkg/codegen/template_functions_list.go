package codegen

import (
	"fmt"
	"math"
	"reflect"
)

// inList is a helper function used to return if a given item (needle) is in the given list (haystack)
func inList(haystack []any, needle any) bool {
	for _, h := range haystack {
		if reflect.DeepEqual(needle, h) {
			return true
		}
	}
	return false
}

// list takes in a varadict series of items and returns them as a single list
func list(v ...any) []any {
	return v
}

// listAppend takes in a starting list and a given value and returns a new list.
// This new list contains all values of the previous list and the given value at the end of the new list.
// should the provided list be anything other than a slice or array, this will throw an error
func listAppend(list any, v any) ([]any, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		nl := make([]any, l)
		for i := 0; i < l; i++ {
			nl[i] = l2.Index(i).Interface()
		}

		return append(nl, v), nil

	default:
		return nil, fmt.Errorf("cannot push on type %s", tp)
	}
}

// listPrepend takes in a starting list and a given value and returns a new list.
// this new list contains the given value, followed by all values of the previous list.
// should the provided list be anything other than a slice or array, this will throw an error
func listPrepend(list any, v any) ([]any, error) {

	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		nl := make([]any, l)
		for i := 0; i < l; i++ {
			nl[i] = l2.Index(i).Interface()
		}

		return append([]any{v}, nl...), nil

	default:
		return nil, fmt.Errorf("cannot prepend on type %s", tp)
	}
}

// listFirst returns the first item in the list if there is at least one item in the list.
// This paris well with listRest.
// Should the provided list be anything other than a slice or array, this will throw an error
func listFirst(list any) (any, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		if l == 0 {
			return nil, nil
		}

		return l2.Index(0).Interface(), nil
	default:
		return nil, fmt.Errorf("cannot find first on type %s", tp)
	}
}

// listRest will return all items in the list in a new list excluding the first item. Order is maintained.
// This pairs well with listFirst
// Should the provided list be anything other than a slice or array, this will throw an error
func listRest(list any) ([]any, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		if l == 0 {
			return nil, nil
		}

		nl := make([]any, l-1)
		for i := 1; i < l; i++ {
			nl[i-1] = l2.Index(i).Interface()
		}

		return nl, nil
	default:
		return nil, fmt.Errorf("cannot find rest on type %s", tp)
	}
}

// listLast returns the last item in the list if there is at least one item in the list.
// This paris well with listInitial.
// Should the provided list be anything other than a slice or array, this will throw an error
func listLast(list any) (any, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		if l == 0 {
			return nil, nil
		}

		return l2.Index(l - 1).Interface(), nil
	default:
		return nil, fmt.Errorf("cannot find last on type %s", tp)
	}
}

// listInitial will return all items in the list in a new list excluding the last item. Order is maintained.
// This pairs well with listLast
// Should the provided list be anything other than a slice or array, this will throw an error
func listInitial(list any) ([]any, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		if l == 0 {
			return nil, nil
		}

		nl := make([]any, l-1)
		for i := 0; i < l-1; i++ {
			nl[i] = l2.Index(i).Interface()
		}

		return nl, nil
	default:
		return nil, fmt.Errorf("cannot find initial on type %s", tp)
	}
}

// listReverse will return all items in the list in a new list, but with their order reversed.
// Should the provided list be anything other than a slice or array, this will throw an error
func listReverse(v any) ([]any, error) {
	tp := reflect.TypeOf(v).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(v)

		l := l2.Len()
		// We do not sort in place because the incoming array should not be altered.
		nl := make([]any, l)
		for i := 0; i < l; i++ {
			nl[l-i-1] = l2.Index(i).Interface()
		}

		return nl, nil
	default:
		return nil, fmt.Errorf("cannot find reverse on type %s", tp)
	}
}

// listUniq will return a new list with duplicate items removed. Order is maintained excluding the removal of the duplicates.
// Should the provided list be anything other than a slice or array, this will throw an error
func listUniq(list any) ([]any, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		dest := []any{}
		var item any
		for i := 0; i < l; i++ {
			item = l2.Index(i).Interface()
			if !inList(dest, item) {
				dest = append(dest, item)
			}
		}

		return dest, nil
	default:
		return nil, fmt.Errorf("cannot find uniq on type %s", tp)
	}
}

// listWithout will return a new list with all items specified in the arguments removed. Order is maintained excluding the removed items.
// Should the provided list be anything other than a slice or array, this will throw an error
func listWithout(list any, omit ...any) ([]any, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		res := []any{}
		var item any
		for i := 0; i < l; i++ {
			item = l2.Index(i).Interface()
			if !inList(omit, item) {
				res = append(res, item)
			}
		}

		return res, nil
	default:
		return nil, fmt.Errorf("cannot find without on type %s", tp)
	}
}

// listHas will return true if the provided item exists in the provided list.
// Should the provided list be anything other than a slice or array, this will throw an error
func listHas(needle any, haystack any) (bool, error) {
	if haystack == nil {
		return false, nil
	}
	tp := reflect.TypeOf(haystack).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(haystack)
		var item any
		l := l2.Len()
		for i := 0; i < l; i++ {
			item = l2.Index(i).Interface()
			if reflect.DeepEqual(needle, item) {
				return true, nil
			}
		}

		return false, nil
	default:
		return false, fmt.Errorf("cannot find has on type %s", tp)
	}
}

// listSlice will return a slice of the provided list (note, not a new list).
// listSlice acknowledges up to 2 other args, the beginning and ending positions.
// If not set, they are assumed to be 0 and len() respectively.
// Should the provided list be anything other than a slice or array, this will throw an error
func listSlice(list any, indices ...int) (any, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		if l == 0 {
			return nil, nil
		}

		var start, end int
		if len(indices) > 0 {
			start = int(indices[0])
		}
		if len(indices) < 2 {
			end = l
		} else {
			end = int(indices[1])
		}

		return l2.Slice(start, end).Interface(), nil
	default:
		return nil, fmt.Errorf("list should be type of slice or array but %s", tp)
	}
}

// listConcat will return a new list containing all values from all provided lists.
// Should a provided item be anything other than a slice or array, this will throw an error
func listConcat(lists ...any) (any, error) {
	var res []any
	for idx, list := range lists {
		tp := reflect.TypeOf(list).Kind()
		switch tp {
		case reflect.Slice, reflect.Array:
			l2 := reflect.ValueOf(list)
			for i := 0; i < l2.Len(); i++ {
				res = append(res, l2.Index(i).Interface())
			}
		default:
			return nil, fmt.Errorf("cannot concat index %d type %s as list", idx, tp)
		}
	}
	return res, nil
}

// listChunk returns a list of lists, with each list having been divided into roughly equal parts.
// the first list will contain the first N items, the second teh second N etc.
// the final list will contain all remaining items as required.
func listChunk(size int, list any) ([][]any, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()

		//calculate and create number of lists to create
		cs := int(math.Floor(float64(l-1)/float64(size)) + 1)
		nl := make([][]any, cs)

		for i := 0; i < cs; i++ {
			clen := size
			if i == cs-1 {
				clen = int(math.Floor(math.Mod(float64(l), float64(size))))
				if clen == 0 {
					clen = size
				}
			}

			nl[i] = make([]any, clen)
			for j := 0; j < clen; j++ {
				ix := i*size + j
				nl[i][j] = l2.Index(ix).Interface()
			}
		}

		return nl, nil

	default:
		return nil, fmt.Errorf("cannot chunk type %s", tp)
	}
}
