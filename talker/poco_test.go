package talker_test

import (
	"errors"
	"testing"

	"github.com/Arsfiqball/csverse/talker"
)

func TestError(t *testing.T) {
	t.Run("one level with standard error", func(t *testing.T) {
		var err error

		stdErr := errors.New("test")
		talkerErr := talker.NewError("TEST", "test")

		err = talkerErr.Wrap(stdErr)

		if !errors.Is(err, stdErr) {
			t.Fatal("error is not stdErr")
		}

		if !errors.Is(err, talkerErr) {
			t.Fatal("error is not talkerErr")
		}
	})

	t.Run("two level", func(t *testing.T) {
		var err error

		namedErr1 := talker.NewError("TEST1", "test 1")
		namedErr2 := talker.NewError("TEST2", "test 2")

		err = namedErr2.Wrap(namedErr1)

		if !errors.Is(err, namedErr1) {
			t.Fatal("error is not namedErr1")
		}

		if !errors.Is(err, namedErr2) {
			t.Fatal("error is not namedErr2")
		}
	})

	t.Run("five level with standard error", func(t *testing.T) {
		var err error

		stdErr := errors.New("test")
		namedErr1 := talker.NewError("TEST1", "test 1")
		namedErr2 := talker.NewError("TEST2", "test 2")
		namedErr3 := talker.NewError("TEST3", "test 3")
		namedErr4 := talker.NewError("TEST4", "test 4")

		err = namedErr1.Wrap(stdErr)
		err = namedErr2.Wrap(err)
		err = namedErr3.Wrap(err)
		err = namedErr4.Wrap(err)

		if !errors.Is(err, stdErr) {
			t.Fatal("error is not stdErr")
		}

		if !errors.Is(err, namedErr1) {
			t.Fatal("error is not namedErr1")
		}

		if !errors.Is(err, namedErr2) {
			t.Fatal("error is not namedErr2")
		}

		if !errors.Is(err, namedErr3) {
			t.Fatal("error is not namedErr3")
		}

		if !errors.Is(err, namedErr4) {
			t.Fatal("error is not namedErr4")
		}

		// Test using verbose flag (-v) to print stack trace
		// for _, tc := range talker.ErrorDataFrom(err, 10) {
		// 	t.Log(tc)
		// }
	})
}

func funcThatPanics() {
	panic("test")
}

type someProxyInterface interface {
	SomeProxyMethod()
}

type someProxyStruct struct{}

func newSomeProxyInterface() someProxyInterface {
	return &someProxyStruct{}
}

func (s *someProxyStruct) SomeProxyMethod() {
	funcThatPanics()
}

func someProxyFunc() {
	newSomeProxyInterface().SomeProxyMethod()
}

func TestRecover(t *testing.T) {
	t.Run("panic", func(t *testing.T) {
		erp := talker.NewError("ERR_RECOVERED_PANIC", "panic")

		func() {
			defer talker.RecoverAs(&erp, 10)

			someProxyFunc()
		}()

		// if erp.Error() != "test" {
		// 	t.Fatal("message is not 'test'")
		// }

		// Test using verbose flag (-v) to print stack trace
		// for _, s := range talker.ErrorDataFrom(erp, 10) {
		// 	t.Log(s)
		// }
	})
}

func TestPower(t *testing.T) {
	// TODO: implement
}
