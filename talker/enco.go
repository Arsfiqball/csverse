package talker

import (
	"database/sql/driver"
	"encoding/json"
)

// Attr is a data structure that represents an attribute.
// This can be used to represent optional fields in a struct.
// There are three states an attribute can be in:
// 1. Omitted: The attribute is not present.
// 2. Null: The attribute is present, but has no value.
// 3. Value: The attribute is present and has a value.
// Example:
//
//	type User struct {
//		ID   int
//		Name string
//		Age  talker.Attr[int]
//	}
//
//	user := User{
//		ID:   1,
//		Name: "John",
//		Age:  talker.Value(30),
//	}
//
//	fmt.Println(user.Age.Get()) // 30
//	fmt.Println(user.Age.Present()) // true
//	fmt.Println(user.Age.Filled()) // true
type Attr[T any] struct {
	present bool
	filled  bool
	value   T
}

// Present returns true if the attribute is present.
func (a Attr[T]) Present() bool {
	return a.present
}

// Filled returns true if the attribute is filled.
func (a Attr[T]) Filled() bool {
	return a.filled
}

// Get returns the value of the attribute.
func (a Attr[T]) Get() T {
	return a.value
}

// Value returns the value of the attribute.
func (a Attr[T]) Value() (driver.Value, error) {
	if !a.filled {
		return nil, nil
	}

	return a.Get(), nil
}

// Scan assigns a value from a database driver.
func (a *Attr[T]) Scan(value interface{}) error {
	if value == nil {
		*a = Null[T]()
		return nil
	}

	*a = Value[T](value.(T))

	return nil
}

// MarshalJSON returns the JSON encoding of the attribute.
func (a Attr[T]) MarshalJSON() ([]byte, error) {
	// Currently there is no convention for representing a missing attribute.
	// So we are representing it as null.
	// There is discussion about this in the community here: https://github.com/golang/go/discussions/63397
	if !a.present {
		return []byte("null"), nil
	}

	if !a.filled {
		return []byte("null"), nil
	}

	return json.Marshal(a.Get())
}

// UnmarshalJSON parses the JSON encoding and stores the result in the attribute.
func (a *Attr[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*a = Null[T]()
		return nil
	}

	var value T

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	*a = Value[T](value)

	return nil
}

// Omit returns an attribute that is omitted.
// This is useful when you want to represent an optional field that is not present
// or you want to reset the value of an attribute.
// Example:
//
//	type User struct {
//		ID   int
//		Name string
//		Age  talker.Attr[int]
//	}
//
//	user := User{
//		ID:   1,
//		Name: "John",
//		Age:  talker.Value(30),
//	}
//
//	user.Age = talker.Omit[int]()
//	fmt.Println(user.Age.Present()) // false
//	fmt.Println(user.Age.Filled()) // is also false
func Omit[T any]() Attr[T] {
	return Attr[T]{}
}

// Null returns an attribute that is present, but not filled.
// Example:
//
//	type User struct {
//		ID   int
//		Name string
//		Age  talker.Attr[int]
//	}
//
//	user := User{
//		ID:   1,
//		Name: "John",
//		Age:  talker.Null[int](),
//	}
//
//	fmt.Println(user.Age.Present()) // true, because it is present
//	fmt.Println(user.Age.Filled()) // false, because it is not filled
func Null[T any]() Attr[T] {
	return Attr[T]{present: true}
}

// Empty returns an attribute that is present and filled, but has zero value.
// Example:
//
//	type User struct {
//		ID   int
//		Name string
//		Age  talker.Attr[int]
//	}
//
//	user := User{
//		ID:   1,
//		Name: "John",
//		Age:  talker.Empty[int](),
//	}
//
//	fmt.Println(user.Age.Present()) // true, because it is present
//	fmt.Println(user.Age.Filled()) // true, because it is filled
//	fmt.Println(user.Age.Get()) // 0, because the zero value of int is 0
func Empty[T any]() Attr[T] {
	return Attr[T]{present: true, filled: true}
}

// Value returns an attribute that is present and filled with a value.
// Example:
//
//	type User struct {
//		ID   int
//		Name string
//		Age  talker.Attr[int]
//	}
//
//	user := User{
//		ID:   1,
//		Name: "John",
//		Age:  talker.Value(30),
//	}
//
//	fmt.Println(user.Age.Present()) // true, because it is present
//	fmt.Println(user.Age.Filled()) // true, because it is filled
//	fmt.Println(user.Age.Get()) // 30
func Value[T any](value T) Attr[T] {
	return Attr[T]{present: true, filled: true, value: value}
}

// Filler is an interface that tells if an attribute is filled.
type Filler interface {
	Filled() bool
}

// Filled returns true if all the attributes are filled.
// This is useful when you want to check if all the attributes are filled before performing an operation.
// Example:
//
//	if talker.Filled(user.Name, user.Age) {
//		// Perform operation.
//	}
func Filled(attrs ...Filler) bool {
	for _, attr := range attrs {
		if !attr.Filled() {
			return false
		}
	}

	return true
}

// Presenter is an interface that tells if an attribute is present.
type Presenter interface {
	Present() bool
}

// Present returns true if all the attributes are present.
// This is useful when you want to check if all the attributes are present before performing an operation.
// Example:
//
//	if talker.Present(user.Name, user.Age) {
//		// Perform operation.
//	}
func Present(attrs ...Presenter) bool {
	for _, attr := range attrs {
		if !attr.Present() {
			return false
		}
	}

	return true
}
