package talker

import (
	"encoding/json"
	"fmt"
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

func (a Attr[T]) IsZero() bool {
	return !a.present
}

// String returns the string representation of the attribute.
func (a Attr[T]) String() string {
	return fmt.Sprintf("%v", a.value)
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

type EmptyBehavior int

const (
	ZeroEmpty EmptyBehavior = iota
	OmitEmpty
	NullEmpty
)

type Object struct {
	emptyBehavior EmptyBehavior
	attrs         map[string]string
}

var _ fmt.Stringer = Object{}

func NewObject() Object {
	return Object{attrs: map[string]string{}, emptyBehavior: ZeroEmpty}
}

func (o Object) WithEmptyBehavior(emptyBehavior EmptyBehavior) Object {
	o.emptyBehavior = emptyBehavior

	return o
}

func (o Object) Present() bool {
	if len(o.attrs) == 0 && o.emptyBehavior == OmitEmpty {
		return false
	}

	return true
}

func (o Object) Filled() bool {
	if len(o.attrs) == 0 && o.emptyBehavior != ZeroEmpty {
		return false
	}

	return true
}

func (o Object) With(key string, value any) Object {
	if p, ok := value.(Presenter); ok && !p.Present() {
		return o
	}

	if f, ok := value.(Filler); ok && !f.Filled() {
		o.attrs[key] = "null"

		return o
	}

	o.attrs[key] = fmt.Sprintf("%v", value)

	return o
}

func (o Object) WithQuoted(key string, value any) Object {
	if p, ok := value.(Presenter); ok && !p.Present() {
		return o
	}

	if f, ok := value.(Filler); ok && !f.Filled() {
		o.attrs[key] = "null"

		return o
	}

	o.attrs[key] = fmt.Sprintf(`"%s"`, value)

	return o
}

func (o Object) String() string {
	attrStr := "{"
	total := len(o.attrs)

	for key, value := range o.attrs {
		attrStr += fmt.Sprintf(`"%s":%v`, key, value)

		if total > 1 {
			attrStr += ","
		}

		total--
	}

	return attrStr + "}"
}

type Array struct {
	emptyBehavior EmptyBehavior
	values        []string
}

func NewArray() Array {
	return Array{values: []string{}, emptyBehavior: ZeroEmpty}
}

func (a Array) WithEmptyBehavior(emptyBehavior EmptyBehavior) Array {
	a.emptyBehavior = emptyBehavior

	return a
}

func (a Array) Present() bool {
	if len(a.values) == 0 && a.emptyBehavior == OmitEmpty {
		return false
	}

	return true
}

func (a Array) Filled() bool {
	if len(a.values) == 0 && a.emptyBehavior != ZeroEmpty {
		return false
	}

	return true
}

func (a Array) Add(value any) Array {
	if p, ok := value.(Presenter); ok && !p.Present() {
		return a
	}

	if f, ok := value.(Filler); ok && !f.Filled() {
		a.values = append(a.values, "null")

		return a
	}

	a.values = append(a.values, fmt.Sprintf("%v", value))

	return a
}

func (a Array) AddQuoted(value any) Array {
	if p, ok := value.(Presenter); ok && !p.Present() {
		return a
	}

	if f, ok := value.(Filler); ok && !f.Filled() {
		a.values = append(a.values, "null")

		return a
	}

	a.values = append(a.values, fmt.Sprintf(`"%s"`, value))

	return a
}

func (a Array) String() string {
	attrStr := "["
	total := len(a.values)

	for _, value := range a.values {
		attrStr += value

		if total > 1 {
			attrStr += ","
		}

		total--
	}

	return attrStr + "]"
}

// Text is a data structure that represents a text fmt.Stringer.
type Text struct {
	text string
}

var _ fmt.Stringer = Text{}

// NewText returns a new text fmt.Stringer.
func NewText(text string) Text {
	return Text{text: text}
}

// String returns the string representation of the text fmt.Stringer.
func (t Text) String() string {
	return t.text
}

// M is a data structure that represents a map.
type M map[string]any

// Element is a data structure that represents an element fmt.Stringer.
type Element struct {
	tag    string
	attrs  M
	childs []fmt.Stringer
}

var _ fmt.Stringer = Element{}

// NewElement returns a new element fmt.Stringer.
func NewElement(tag string) Element {
	return Element{tag: tag, attrs: M{}, childs: []fmt.Stringer{}}
}

// With adds an attribute to the element fmt.Stringer.
func (e Element) With(key string, value any) Element {
	e.attrs[key] = value

	return e
}

// Content adds child fmt.Stringers to the element fmt.Stringer.
func (e Element) Content(contents ...fmt.Stringer) Element {
	e.childs = contents

	return e
}

// Text adds text fmt.Stringers to the element fmt.Stringer.
func (e Element) Text(texts ...string) Element {
	e.childs = make([]fmt.Stringer, len(texts))

	for i, text := range texts {
		e.childs[i] = NewText(text)
	}

	return e
}

// String returns the string representation of the element fmt.Stringer.
func (e Element) String() string {
	attrStr := ""

	for key, value := range e.attrs {
		if p, ok := value.(Presenter); ok && !p.Present() {
			continue
		}

		attrStr += fmt.Sprintf(` %s="%v"`, key, value)
	}

	content := ""

	for _, child := range e.childs {
		content += fmt.Sprintf("%v", child)
	}

	if e.tag == "" {
		return content
	}

	return fmt.Sprintf("<%s%s>%s</%s>", e.tag, attrStr, content, e.tag)
}

// Cond is a data structure that represents a conditional fmt.Stringer.
type Cond struct {
	cond   bool
	values []fmt.Stringer
}

var _ fmt.Stringer = Cond{}

// If returns a new conditional fmt.Stringer.
func If(cond bool) Cond {
	return Cond{cond: cond, values: []fmt.Stringer{}}
}

// Then adds child fmt.Stringers to the conditional fmt.Stringer if the condition is true.
func (c Cond) Then(values ...fmt.Stringer) Cond {
	if c.cond {
		c.values = values
	}

	return c
}

// Else adds child fmt.Stringers to the conditional fmt.Stringer if the condition is false.
func (c Cond) Else(values ...fmt.Stringer) Cond {
	if !c.cond {
		c.values = values
	}

	return c
}

// String returns the string representation of the conditional fmt.Stringer.
func (c Cond) String() string {
	content := ""

	for _, child := range c.values {
		content += fmt.Sprintf("%v", child)
	}

	return content
}

// ForEach is a function that applies a function to each item in a list and returns a new list.
func ForEach[T any, U any](items []T, fn func(T) U) []U {
	newItems := make([]U, len(items))

	for i, item := range items {
		newItems[i] = fn(item)
	}

	return newItems
}

type Template[T any] struct {
	attrs  T
	childs []fmt.Stringer
	render func(T, []fmt.Stringer) fmt.Stringer
}

var _ fmt.Stringer = Template[any]{}

func NewTemplate[T any](render func(T, []fmt.Stringer) fmt.Stringer) Template[T] {
	var attrs T

	if render == nil {
		render = func(attrs T, childs []fmt.Stringer) fmt.Stringer {
			return NewElement("")
		}
	}

	return Template[T]{render: render, attrs: attrs, childs: []fmt.Stringer{}}
}

func (t Template[T]) With(attrs T) Template[T] {
	t.attrs = attrs

	return t
}

func (t Template[T]) Content(contents ...fmt.Stringer) Template[T] {
	t.childs = contents

	return t
}

func (t Template[T]) String() string {
	return fmt.Sprintf("%v", t.render(t.attrs, t.childs))
}
