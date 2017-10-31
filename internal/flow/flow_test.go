package flow

import (
	"github.com/stretchr/testify/require"

	"testing"
)

func TestStepCallsDo(t *testing.T) {
	called := false
	step := &Step{
		Do: func() {
			called = true
		},
	}

	step.Run(nil)

	require.True(t, called, "Step's do wasn't called")
}

func TestStepCallsSuccessAfterDo(t *testing.T) {
	calls := 0
	step := &Step{
		Do: func() {
			require.Equal(t, 0, calls, "Do should be called first")
			calls++
		},
		Success: &Step{
			Do: func() {
				require.Equal(t, 1, calls, "Success should be called second")
				calls++
			},
		},
		Error: &Step{
			Do: func() {
				t.Fatalf("Error should not have been called")
			},
		},
	}

	step.Run(nil)

	require.Equal(t, 2, calls, "Both do and success should be called")
}

func TestStepCallsErrorIfDoPanics(t *testing.T) {
	defer func() { recover() }()
	calls := 0
	step := &Step{
		Do: func() {
			require.Equal(t, 0, calls, "Do should be called first")
			calls++
			panic(42)
		},
		Success: &Step{
			Do: func() {
				t.Fatalf("Success should not have been called")
			},
		},
		Error: &Step{
			Do: func() {
				require.Equal(t, 1, calls, "Error should be called second")
				calls++
			},
		},
	}

	step.Run(nil)

	require.Equal(t, 2, calls, "Both do and error should be called")
}

func TestStepCallsOsExitIfAskedTo(t *testing.T) {
	exitCalled := false
	defer exitShouldBeCalledWith(t, 42, &exitCalled)()

	step := &Step{}

	step.Run(exit(42))

	require.True(t, exitCalled, "should have called exit")
}

func TestStepRethrowsPanic(t *testing.T) {
	defer func() {
		require.Equal(t, 42, recover(), "should panicked with the same value")
	}()

	step := &Step{}

	step.Run(42)

	t.Fatalf("Should have panicked")
}

func TestStepShouldNopIfNoSuccessNorPanic(t *testing.T) {
	defer exitShouldNotCalled(t)()

	step := &Step{}

	step.Run(nil)
}

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
