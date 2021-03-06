package errors

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"runtime/debug"
	"testing"
)

func TestStackFormatMatches(t *testing.T) {

	defer func() {
		err := recover()
		if err != 'a' {
			t.Fatal(err)
		}

		e, s := Errorf("hi"), debug.Stack()
		bs := [][]byte{e.Stack(), s}

		// Grab the second line in the trace, which is the line above with Errorf on it
		//bs[0] = bytes.SplitN(bytes.SplitN(bs[0], []byte("\n"), 3)[1], []byte("+"), 2)[0]
		// Ignore the debug.Stack() and runtime.Stack() calls
		//bs[1] = bytes.SplitN(bytes.SplitN(bs[1], []byte("\n"), 6)[4], []byte("+"), 2)[0]
		bs[0] = bytes.SplitN(bytes.SplitN(bs[0], []byte("\n"), 3)[1], []byte("+"), 2)[0]
		// Ignore the debug.Stack() and runtime.Stack() calls
		bs[1] = bytes.SplitN(bytes.SplitN(bs[1], []byte("\n"), 6)[4], []byte("+"), 2)[0]

		if bytes.Compare(bs[0], bs[1]) != 0 {
			t.Errorf("Stack didn't match")
			t.Errorf("bs0=%s", bs[0])
			t.Errorf("bs1=%s", bs[1])
		}
	}()

	a()
}

func TestSkipWorks(t *testing.T) {

	defer func() {
		err := recover()
		if err != 'a' {
			t.Fatal(err)
		}

		bs := [][]byte{Wrap("hi", 2).Stack(), debug.Stack()}

		bs[0] = bytes.TrimRight(bytes.SplitN(bytes.SplitN(bs[0], []byte("\n"), 3)[1], []byte("+"), 2)[0], " ")
		bs[1] = bytes.TrimRight(bytes.SplitN(bytes.SplitN(bs[1], []byte("\n"), 10)[8], []byte("+"), 2)[0], " ")

		if bytes.Compare(bs[0], bs[1]) != 0 {
			t.Errorf("Stack didn't match")
			t.Errorf("bs0=%s", bs[0])
			t.Errorf("bs1=%s", bs[1])
		}
	}()

	a()
}

func TestNew(t *testing.T) {

	err := New("foo")

	if err.Error() != "foo" {
		t.Errorf("Wrong message")
	}

	err = New(fmt.Errorf("foo"))

	if err.Error() != "foo" {
		t.Errorf("Wrong message")
	}

	bs := [][]byte{New("foo").Stack(), debug.Stack()}

	bs[0] = bytes.SplitN(bytes.SplitN(bs[0], []byte("\n"), 3)[1], []byte("+"), 2)[0]
	bs[1] = bytes.SplitN(bytes.SplitN(bs[1], []byte("\n"), 6)[4], []byte("+"), 2)[0]

	if bytes.Compare(bs[0], bs[1]) != 0 {
		t.Errorf("Stack didn't match")
		t.Errorf("bs0=%s", bs[0])
		t.Errorf("bs1=%s", bs[1])
	}

	if err.ErrorStack() != err.TypeName()+": "+err.Error()+"\n\ngoroutine 1 [running]:\n"+string(err.Stack()) {
		t.Errorf("ErrorStack is in the wrong format")
		t.Errorf("es=%s", err.ErrorStack())
	}
}

func TestIs(t *testing.T) {

	if Is(nil, io.EOF) {
		t.Errorf("nil is an error")
	}

	if !Is(io.EOF, io.EOF) {
		t.Errorf("io.EOF is not io.EOF")
	}

	if !Is(io.EOF, New(io.EOF)) {
		t.Errorf("io.EOF is not New(io.EOF)")
	}

	if !Is(New(io.EOF), New(io.EOF)) {
		t.Errorf("New(io.EOF) is not New(io.EOF)")
	}

	if Is(io.EOF, fmt.Errorf("io.EOF")) {
		t.Errorf("io.EOF is fmt.Errorf")
	}

}

func TestWrapError(t *testing.T) {

	e := func() error {
		return Wrap("hi", 1)
	}()

	if e.Error() != "hi" {
		t.Errorf("Constructor with a string failed")
	}

	if Wrap(fmt.Errorf("yo"), 0).Error() != "yo" {
		t.Errorf("Constructor with an error failed")
	}

	if Wrap(e, 0) != e {
		t.Errorf("Constructor with an Error failed")
	}

	if Wrap(nil, 0).Error() != "<nil>" {
		t.Errorf("Constructor with nil failed")
	}
}

func TestWrapPrefixError(t *testing.T) {

	e := func() error {
		return WrapPrefix("hi", "prefix", 1)
	}()

	if e.Error() != "prefix: hi" {
		t.Errorf("Constructor with a string failed")
	}

	if WrapPrefix(fmt.Errorf("yo"), "prefix", 0).Error() != "prefix: yo" {
		t.Errorf("Constructor with an error failed")
	}

	prefixed := WrapPrefix(e, "prefix", 0)
	original := e.(*Error)

	if prefixed.Err != original.Err || !reflect.DeepEqual(prefixed.stack, original.stack) || !reflect.DeepEqual(prefixed.frames, original.frames) || prefixed.Error() != "prefix: prefix: hi" || !Is(prefixed, original) || !Is(original, prefixed) || prefixed.Unwrap().Error() != "hi" {
		t.Errorf("Constructor with an Error failed: original=%s, prefixed=%s", original.Error(), prefixed.Error())
	}

	if WrapPrefix(nil, "prefix", 0).Error() != "prefix: <nil>" {
		t.Errorf("Constructor with nil failed")
	}
}

type errType string

func (e errType) Error() string {
	return string(e)
}

func TestAs(t *testing.T) {
	original := errType("hi")
	var e errType
	if !As(New(original), &e) {
		t.Error("As failed to convert to errType")
	}
	if e != original {
		t.Error("As did not return original error")
	}
}

func ExampleErrorf(x int) (int, error) {
	if x%2 == 1 {
		return 0, Errorf("can only halve even numbers, got %d", x)
	}
	return x / 2, nil
}

func ExampleWrapError() (int, error) {
	// Wrap io.EOF with the current stack-trace and return it
	return 0, Wrap(io.EOF, 0)
}

func ExampleWrapError_skip() {
	defer func() {
		if err := recover(); err != nil {
			// skip 1 frame (the deferred function) and then return the wrapped err
			err = Wrap(err, 1)
		}
	}()
}

func ExampleIs(reader io.Reader, buff []byte) {
	_, err := reader.Read(buff)
	if Is(err, io.EOF) {
		return
	}
}

func ExampleNew(UnexpectedEOF error) error {
	// calling New attaches the current stacktrace to the existing UnexpectedEOF error
	return New(UnexpectedEOF)
}

func ExampleWrap() error {

	if err := recover(); err != nil {
		return Wrap(err, 1)
	}

	return a()
}

func ExampleError_Error(err error) {
	fmt.Println(err.Error())
}

func ExampleError_ErrorStack(err error) {
	fmt.Println(err.(*Error).ErrorStack())
}

func ExampleError_Stack(err *Error) {
	fmt.Println(err.Stack())
}

func ExampleError_TypeName(err *Error) {
	fmt.Println(err.TypeName(), err.Error())
}

func ExampleError_StackFrames(err *Error) {
	for _, frame := range err.StackFrames() {
		fmt.Println(frame.File, frame.LineNumber, frame.Package, frame.Name)
	}
}

func a() error {
	b(5)
	return nil
}

func b(i int) {
	c()
}

func c() {
	panic('a')
}
