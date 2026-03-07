package main

import (
	"fmt"
)

// background is a helper that accepts an arbitrary function as a parameter and runs it in a
// in goroutine in the background.
func (app *application) background(fn func()) {
	// Increment the WaitGroup counter
	app.wg.Add(1)

	go func() {
		// Use defer to decrement the WaitGroup counter before the goroutine returns.
		defer app.wg.Done()

		// Recover from any panic
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		// Execute the arbitrary function that we passed as the parameter
		fn()
	}()
}
