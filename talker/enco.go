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

// Node is a data structure that represents a node in a tree.
type Node interface {
	String() string
	Childs() []Node
}

// Text is a data structure that represents a text node.
type Text struct {
	text string
}

var _ Node = Text{}

// NewText returns a new text node.
func NewText(text string) Text {
	return Text{text: text}
}

// String returns the string representation of the text node.
func (t Text) String() string {
	return t.text
}

// Childs returns the child nodes of the text node.
func (t Text) Childs() []Node {
	return nil
}

// M is a data structure that represents a map.
type M map[string]any

// Element is a data structure that represents an element node.
type Element struct {
	tag    string
	attrs  M
	childs []Node
}

var _ Node = Element{}

// NewElement returns a new element node.
func NewElement(tag string) Element {
	return Element{tag: tag, attrs: M{}, childs: []Node{}}
}

// With adds an attribute to the element node.
func (e Element) With(key string, value any) Element {
	e.attrs[key] = value

	return e
}

// Content adds child nodes to the element node.
func (e Element) Content(contents ...Node) Element {
	e.childs = contents

	return e
}

// Text adds text nodes to the element node.
func (e Element) Text(texts ...string) Element {
	e.childs = make([]Node, len(texts))

	for i, text := range texts {
		e.childs[i] = NewText(text)
	}

	return e
}

// Childs returns the child nodes of the element node.
func (e Element) Childs() []Node {
	return e.childs
}

// String returns the string representation of the element node.
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

// Cond is a data structure that represents a conditional node.
type Cond struct {
	cond   bool
	values []Node
}

var _ Node = Cond{}

// If returns a new conditional node.
func If(cond bool) Cond {
	return Cond{cond: cond, values: []Node{}}
}

// Then adds child nodes to the conditional node if the condition is true.
func (c Cond) Then(values ...Node) Cond {
	if c.cond {
		c.values = values
	}

	return c
}

// Else adds child nodes to the conditional node if the condition is false.
func (c Cond) Else(values ...Node) Cond {
	if !c.cond {
		c.values = values
	}

	return c
}

// Childs returns the child nodes of the conditional node.
func (c Cond) Childs() []Node {
	return c.values
}

// String returns the string representation of the conditional node.
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
	childs []Node
	render func(T, []Node) Node
}

var _ Node = Template[any]{}

func NewTemplate[T any](render func(T, []Node) Node) Template[T] {
	var attrs T

	if render == nil {
		render = func(attrs T, childs []Node) Node {
			return NewElement("")
		}
	}

	return Template[T]{render: render, attrs: attrs, childs: []Node{}}
}

func (t Template[T]) With(attrs T) Template[T] {
	t.attrs = attrs

	return t
}

func (t Template[T]) Content(contents ...Node) Template[T] {
	t.childs = contents

	return t
}

func (t Template[T]) Childs() []Node {
	return t.childs
}

func (t Template[T]) String() string {
	return fmt.Sprintf("%v", t.render(t.attrs, t.childs))
}
