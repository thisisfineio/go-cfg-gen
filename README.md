# go-cfg-gen
Generates a configuration file from a go struct. (Initially only simple types)

# Supported
Currently the following types are supported inside a struct

* Structs
* Pointers to structs
* Primitive types
..* int, int8, int16, int32, int64
..* uint, uint8, uint16, uint32, uint64
..* float32, float64
..* bool
..* string
* Slices of primitive types ([]string, etc), and slices of slices of primitive types ([][]string, etc)
..* Looking for input on how to handle arbitrary slice depths and slices of complex types
* Maps with keys of primitive types and values of primitive types
..* Like with slices, looking for ways to handle complex types for maps
* The empty interface

# Unsupported
* Slices of slices of slices ([][][]string) or greater
* Maps with complex value types

# Ignored types
* Uintptr
* unsafe.Pointer
* Non empty Interfaces
* Channels