# Versatile

Versatile is a Go library providing flexible type conversion, comparison, and reflection utilities. It aims to bridge strict type boundaries safely at runtime, making it easier to work with dynamic data, function signatures, and interface abstractions.

## Core Features

- **Flexible Comparison (`compare.go`)**: Compare values with customizable looseness, handling pointers, type conversions, zero values, and driver values seamlessly. Supports custom equality via interfaces.
- **Function Casting (`func.go`)**: Safely cast functions and methods from one signature to another, automatically resolving argument types, dropping extraneous returns, and injecting parameters like `context.Context` and pre-bound arguments.
- **Dynamic Setters (`setters.go`)**: Reflectively assign and convert values from diverse source types into destination pointers, with support for SQL scanners, standard `driver.Valuer`, string parsing, and complex types.
- **Reflection Utilities (`django_reflect.go`)**: Deep zero-value checking and safe reflection-based type conversions.

## Usage Examples

### Flexible Comparison

The `Equals` function allows comparing values that wouldn't normally pass standard equality checks.  
You can use `FLAG_EQ` constants to dictate the looseness of the comparison, such as:

- `EQ_NONE`: Exact equality only.
- `EQ_TYPE_CONVERT`: Converts types automatically (e.g., `int8` to `int64`, `string` to `[]byte`).
- `EQ_IGNORE_PTR`: Ignores pointer boundaries, deferencing automatically.
- `EQ_ZEROS`: Considers `nil` slices/maps and empty slices/maps as equal.
- `EQ_DRIVER_VALUE`: Checks for and calls `driver.Valuer` on types if applicable.
- `EQ_DFLT`: Includes `EQ_DRIVER_VALUE`, `EQ_ZEROS`, `EQ_IGNORE_PTR`, and `EQ_TYPE_CONVERT`.

It also respects interfaces! If a struct implements `EqualityChecker` (`Equal(any) bool`) or `EqualityChecker2` (`Equals(any) bool`), those methods will be used.

```go
package main

import (
    "fmt"
    "github.com/Nigel2392/versatile"
)

// Implementing EqualityChecker
type MyType struct { Val int }

func (m MyType) Equal(other any) bool {
    if o, ok := other.(MyType); ok {
        return m.Val == o.Val
    }
    return false
}

func main() {
    // Auto-convert numeric types
    res := versatile.Equals(1, int64(1), versatile.EQ_TYPE_CONVERT)
    fmt.Println(res) // true

    // Compare strings to byte slices
    res = versatile.Equals("test", []byte("test"), versatile.EQ_TYPE_CONVERT)
    fmt.Println(res) // true

    // Ignore pointers (dereferences automatically)
    val := 1
    res = versatile.Equals(1, &val, versatile.EQ_IGNORE_PTR)
    fmt.Println(res) // true

    // Treat nil slices/maps as equal to empty slices/maps
    res = versatile.Equals([]int(nil), []int{}, versatile.EQ_ZEROS)
    fmt.Println(res) // true
    
    // Using custom interfaces
    a, b := MyType{Val: 10}, MyType{Val: 10}
    res = versatile.Equals(a, b)
    fmt.Println(res) // true
}
```

### Function Casting

`CastFunc` allows wrapping functions to match a desired signature. It handles type coercion, interface unpacking (`any` to specific types and vice versa), variadic functions, and dropping unused return variables.  
You can use options like `WithContext` or `WithFuncArgs` to partially apply arguments on creation!

```go
package main

import (
    "context"
    "fmt"
    "github.com/Nigel2392/versatile"
)

func main() {
    // Cast and convert arguments automatically (float64 -> int)
    src := func(a, b float64) float64 { return a + b }
    out, _ := versatile.CastFunc[func(int, int) float64](src)
    fmt.Println(out(2, 5)) // 7

    // Inject context seamlessly 
    srcCtx := func(ctx context.Context, a int) int { return a * 2 }
    outCtx, _ := versatile.CastFunc[func(int) int](srcCtx, versatile.WithContext(context.Background()))
    fmt.Println(outCtx(5)) // 10

    // Pre-bind discrete arguments using WithFuncArgs
    srcMany := func(ctx context.Context, a, b, c int64) float32 { return float32(a + b + c) }
    outMany, _ := versatile.CastFunc[func(int8, int8) int64](
        srcMany, 
        versatile.WithContext(context.Background()),
        versatile.WithFuncArgs(10), // pre-binds 10 to 'a'
    )
    fmt.Println(outMany(5, 5)) // 20

    // Discard extra return values (e.g., returning only the error)
    srcMulti := func(s string, n int) (string, int, error) { return s + "!", n, nil }
    outErrOnly, _ := versatile.CastFunc[func(string, int) error](srcMulti)
    err := outErrOnly("hello", 3)
    fmt.Println(err) // <nil>

    // Interface unwrapping: mapping 'any' to discrete structs/types
    srcAny := func(a int, b string) string { return fmt.Sprintf("%d:%s", a, b) }
    outAny, _ := versatile.CastFunc[func(any, any) string](srcAny)
    fmt.Println(outAny(42, "test")) // "42:test"
}
```

### Dynamic Setters

`ScanTo` sets a value into a destination pointer, parsing and converting the source data as required based on the provided flags (`SF_SQL_SCANNER`, `SF_STRCONV`, `SF_REFLECTCONV`).

```go
package main

import (
    "fmt"
    "github.com/Nigel2392/versatile"
)

func main() {
    var dstInt int
    var dstBool bool
    var dstFloat float64
    var dstBytes []byte

    // Parse a string directly into an int pointer
    versatile.ScanTo(&dstInt, "123", versatile.SF_STRCONV)
    fmt.Println(dstInt) // 123

    // Parse strings to booleans
    versatile.ScanTo(&dstBool, "true", versatile.SF_STRCONV)
    fmt.Println(dstBool) // true

    // Number to Float conversions
    versatile.ScanTo(&dstFloat, 123, versatile.SF_DEFAULT)
    fmt.Println(dstFloat) // 123.0

    // String to byte slice
    versatile.ScanTo(&dstBytes, "hello", versatile.SF_DEFAULT)
    fmt.Println(string(dstBytes)) // "hello"
}
```

### Deep Zero Checking

`IsZero` handles nested and complex types to determine if they are empty or contain only zero values. It supports interfaces by inspecting `IsZeroer` if implemented.

```go
package main

import (
    "fmt"
    "github.com/Nigel2392/versatile"
)

func main() {
    fmt.Println(versatile.IsZero(0))           // true
    fmt.Println(versatile.IsZero([]int{}))     // true
    fmt.Println(versatile.IsZero(nil))         // true
    fmt.Println(versatile.IsZero([]int{0, 0})) // true
    fmt.Println(versatile.IsZero(map[string]int{})) // true
}
```
