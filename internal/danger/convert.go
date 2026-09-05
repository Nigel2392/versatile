package danger

import (
	re "reflect"
)

func init() {
	for k, v := range unsafeKinds {
		for k2, rev := range v {
			if !rev {
				continue
			}

			if _, ok := unsafeKinds[k2]; !ok {
				unsafeKinds[k2] = make(map[re.Kind]bool)
			}
			unsafeKinds[k2][k] = true
		}
	}
}

// if the bool mapvalue is true, a reverse of the same
// rule will be created (i.e. Int is also /not/ safe to convert to String)
var unsafeKinds = map[re.Kind]map[re.Kind]bool{
	re.String: {
		re.Int:    true,
		re.Int8:   true,
		re.Int16:  true,
		re.Int32:  true,
		re.Int64:  true,
		re.Uint:   true,
		re.Uint8:  true,
		re.Uint16: true,
		re.Uint32: true,
		re.Uint64: true,
	},
	re.Slice: {
		re.Array: false,
	},
}

func IsSafeConversion(from, to re.Type) bool {
	if from.Kind() == to.Kind() {
		return true
	}
	if unsafeTo, ok := unsafeKinds[from.Kind()]; ok {
		if _, ok := unsafeTo[to.Kind()]; ok {
			return false
		}
	}
	return true
}
