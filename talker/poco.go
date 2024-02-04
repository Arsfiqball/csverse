package talker

import (
	"context"
	"fmt"
	"runtime"
)

// Error is a custom error type that can be used to wrap errors and add additional information.
// This error must be created with the NewError function.
type Error struct {
	code       string
	info       string
	declaredAt string
	wrappedAt  string
	data       interface{}
	parent     error
}

// NewError creates a new Error with the given code and default info.
// The default info can be overridden with the WithInfo method.
// Other error can be wrapped with the Wrap method.
// Additional data can be added with the WithData method.
// Example:
//
//	Err001 := talker.NewError("ERR_001", "Something went wrong") // root error definition
//
//	func doSomething() error {
//		// ...
//		return Err001.Wrap(errors.New("wrapped error")).WithInfo("Something went wrong in doSomething")
//	}
func NewError(code string, defaultInfo string) Error {
	var caller string

	_, file, line, ok := runtime.Caller(1)

	if ok {
		caller = fmt.Sprintf("%s:%d", file, line)
	}

	return Error{code: code, info: defaultInfo, declaredAt: caller}
}

// Wrap wraps the given error with the current error.
func (e Error) Wrap(err error) Error {
	var caller string

	_, file, line, ok := runtime.Caller(1)

	if ok {
		caller = fmt.Sprintf("%s:%d", file, line)
	}

	e.parent = err
	e.wrappedAt = caller

	return e
}

// WithInfo adds additional information to the error.
func (e Error) WithInfo(message string) Error {
	e.info = message

	return e
}

// Info returns the additional information of the error.
func (e Error) Info() string {
	return e.info
}

// WithData adds additional data to the error.
func (e Error) WithData(data interface{}) Error {
	e.data = data

	return e
}

// Data returns the additional data of the error.
func (e Error) Data() interface{} {
	return e.data
}

// Error returns the string representation of the error.
func (e Error) Error() string {
	return e.info
}

// Is checks if the error is of the given type.
func (e Error) Is(target error) bool {
	if target == nil {
		return false
	}

	pocoErr, ok := target.(Error)

	if ok && e.code == pocoErr.code {
		return true
	}

	return false
}

// Unwrap returns the parent error.
func (e Error) Unwrap() error {
	return e.parent
}

// ErrorData is a data structure that represents an error.
// It can be used to serialize the error to JSON.
type ErrorData struct {
	Code     string      `json:"code"`
	Info     string      `json:"info"`
	Location string      `json:"location"`
	Data     interface{} `json:"data"`
	Child    *ErrorData  `json:"child"`
}

// ErrorDataFrom creates an ErrorData from the given error.
// If the error has a parent, it will be included in the ErrorData as a child.
// The depth parameter specifies how many levels of children to include.
// If the depth is 0, no children will be included.
// Example:
//
//	errContainer := talker.NewError("ERR_001", "Something went wrong")
//	errContainer = errContainer.Wrap(errors.New("wrapped error"))
//	errData := talker.ErrorDataFrom(errContainer, 10)
//	fmt.Println(errData)
func ErrorDataFrom(err error, depth int) *ErrorData {
	pocoErr, ok := err.(Error)
	if !ok {
		return nil
	}

	location := pocoErr.declaredAt

	if pocoErr.wrappedAt != "" {
		location = pocoErr.wrappedAt
	}

	result := &ErrorData{
		Code:     pocoErr.code,
		Info:     pocoErr.info,
		Location: location,
		Data:     pocoErr.data,
	}

	if depth > 0 && pocoErr.parent != nil {
		result.Child = ErrorDataFrom(pocoErr.parent, depth-1)
	}

	return result
}

// Recover recovers from a panic and converts it to an Error.
// The depth parameter specifies how many levels of the stack trace to include.
// Example:
//
//	func main() {
//		errContainer := talker.NewError("ERR_001", "Something went wrong")
//		func() {
//			defer talker.RecoverAs(&errContainer, 10)
//			// ... do something that can panic
//		}()
//
//		errData := talker.ErrorDataFrom(errContainer, 10)
//	}
func RecoverAs(out *Error, depth int) {
	if out == nil {
		return
	}

	const skip = 2
	if r := recover(); r != nil {
		pocoErr := *out // Copy the original poco.Error
		pocoErr.info = fmt.Sprintf("%v", r)

		for i := skip; i < depth; i++ {
			pc, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}

			name := "unknown"

			fn := runtime.FuncForPC(pc)
			if fn != nil {
				name = fn.Name()
			}

			childErr := Error{
				code:       pocoErr.code,
				info:       fmt.Sprintf("stack %d", i-skip),
				declaredAt: fmt.Sprintf("%s:%d", file, line),
				wrappedAt:  fmt.Sprintf("%s:%d %s", file, line, name),
				parent:     &pocoErr,
			}

			pocoErr = childErr
		}

		*out = pocoErr
	}
}

type Params map[string]any

// Span starts a new span with the given name and attributes.
// The span will be ended when the returned function is called.
// Example:
//
//	func doSomething(ctx context.Context) {
//		ctx, end := talker.Span(ctx, "doSomething", talker.Params{"key": "value"})
//		defer end()
//		// ... do something
//	}
func Span(ctx context.Context, name string, params Params) (context.Context, func()) {
	pwr, ok := ctx.Value(powerContextKey).(Power)
	if !ok {
		return ctx, func() {}
	}

	var ends []func()

	for _, hook := range pwr.spanHooks {
		newCtx, end := hook(ctx, name, params)

		ends = append(ends, end)
		ctx = newCtx
	}

	return ctx, func() {
		for _, end := range ends {
			end()
		}
	}
}

// Event sends an event with the given name and attributes.
// Example:
//
//	func doSomething(ctx context.Context) {
//		talker.Event(ctx, "doSomething", talker.Params{"key": "value"})
//		// ... do something
//	}
func Event(ctx context.Context, name string, attrs map[string]any) {
	pwr, ok := ctx.Value(powerContextKey).(Power)
	if !ok {
		return
	}

	for _, hook := range pwr.eventHooks {
		hook(ctx, name, attrs)
	}
}

// SpanHook is a function that can be used to hook into the Span function.
type SpanHook func(ctx context.Context, name string, attrs map[string]any) (context.Context, func())

// EventHook is a function that can be used to hook into the Event function.
type EventHook func(ctx context.Context, name string, attrs map[string]any)

// Power is a configuration for the Span and Event functions.
// This "Power" must be created with the NewPower function.
type Power struct {
	spanHooks  []SpanHook
	eventHooks []EventHook
}

// NewPower creates a new Power.
func NewPower() Power {
	return Power{}
}

// WithSpanHook adds a SpanHook to the Power.
func (c Power) WithSpanHook(hook SpanHook) Power {
	c.spanHooks = append(c.spanHooks, hook)

	return c
}

// WithEventHook adds an EventHook to the Power.
func (c Power) WithEventHook(hook EventHook) Power {
	c.eventHooks = append(c.eventHooks, hook)

	return c
}

// PowerContextKey is a context key for the Power.
type PowerContextKey string

const powerContextKey = PowerContextKey("power_context")

// Context adds the Power to the context.
func (c Power) Context(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, powerContextKey, c)
}
