package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBeforeAndAfterFlowOrder(t *testing.T) {
	counter := 0

	app := App("app", "")

	app.Before = callChecker(t, 0, &counter)
	app.Command("c", "", func(c *Cmd) {
		c.Before = callChecker(t, 1, &counter)
		c.Command("cc", "", func(cc *Cmd) {
			cc.Before = callChecker(t, 2, &counter)
			cc.Action = callChecker(t, 3, &counter)
			cc.After = callChecker(t, 4, &counter)
		})
		c.After = callChecker(t, 5, &counter)
	})
	app.After = callChecker(t, 6, &counter)

	app.Run([]string{"app", "c", "cc"})
	require.Equal(t, 7, counter)
}

func TestBeforeAndAfterFlowOrderWhenOneBeforePanics(t *testing.T) {
	defer func() {
		recover()
	}()

	counter := 0

	app := App("app", "")

	app.Before = callChecker(t, 0, &counter)
	app.Command("c", "", func(c *Cmd) {
		c.Before = callChecker(t, 1, &counter)
		c.Command("cc", "", func(cc *Cmd) {
			cc.Before = callCheckerAndPanic(t, 42, 2, &counter)
			cc.Action = func() {
				t.Fatalf("should not have been called")
			}
			cc.After = func() {
				t.Fatalf("should not have been called")
			}
		})
		c.After = callChecker(t, 3, &counter)
	})
	app.After = callChecker(t, 4, &counter)

	app.Run([]string{"app", "c", "cc"})
	require.Equal(t, 5, counter)
}

func TestBeforeAndAfterFlowOrderWhenOneAfterPanics(t *testing.T) {
	defer func() {
		e := recover()
		require.Equal(t, 42, e)
	}()

	counter := 0

	app := App("app", "")

	app.Before = callChecker(t, 0, &counter)
	app.Command("c", "", func(c *Cmd) {
		c.Before = callChecker(t, 1, &counter)
		c.Command("cc", "", func(cc *Cmd) {
			cc.Before = callChecker(t, 2, &counter)
			cc.Action = callChecker(t, 3, &counter)
			cc.After = callCheckerAndPanic(t, 42, 4, &counter)
		})
		c.After = callChecker(t, 5, &counter)
	})
	app.After = callChecker(t, 6, &counter)

	app.Run([]string{"app", "c", "cc"})
	require.Equal(t, 7, counter)
}

func TestBeforeAndAfterFlowOrderWhenMultipleAftersPanic(t *testing.T) {
	defer func() {
		e := recover()
		require.Equal(t, 666, e)
	}()

	counter := 0

	app := App("app", "")

	app.Before = callChecker(t, 0, &counter)
	app.Command("c", "", func(c *Cmd) {
		c.Before = callChecker(t, 1, &counter)
		c.Command("cc", "", func(cc *Cmd) {
			cc.Before = callChecker(t, 2, &counter)
			cc.Action = callChecker(t, 3, &counter)
			cc.After = callCheckerAndPanic(t, 42, 4, &counter)
		})
		c.After = callChecker(t, 5, &counter)
	})
	app.After = callCheckerAndPanic(t, 666, 6, &counter)

	app.Run([]string{"app", "c", "cc"})
	require.Equal(t, 7, counter)
}

func TestCommandAction(t *testing.T) {

	called := false

	app := App("app", "")

	app.Command("a", "", ActionCommand(func() { called = true }))

	app.Run([]string{"app", "a"})

	require.True(t, called, "commandAction should be called")

}

func callChecker(t *testing.T, wanted int, counter *int) func() {
	return func() {
		t.Logf("checker: wanted: %d, got %d", wanted, *counter)
		require.Equal(t, wanted, *counter)
		*counter++
	}
}

func callCheckerAndPanic(t *testing.T, panicValue interface{}, wanted int, counter *int) func() {
	return func() {
		t.Logf("checker: wanted: %d, got %d", wanted, *counter)
		require.Equal(t, wanted, *counter)
		*counter++
		panic(panicValue)
	}
}
