package perky_test

import (
	"context"
	"fmt"

	"github.com/alexdrel/perky"
)

func Example() {
	theme := perky.Key("light")
	retryLimit := perky.Key(3)

	// Attach values to an ordinary Go context at the request boundary.
	native := perky.Attach(context.Background(), theme.Bind("dark"), retryLimit.Bind(5))
	fmt.Println("request:", theme.Get(native), retryLimit.Get(native))

	// Application-only code can carry the values without cancellation or deadlines.
	app := perky.Detach(native)
	preview := app.With(theme.Bind("light"))
	fmt.Println("preview:", theme.Get(preview), retryLimit.Get(preview))
	fmt.Println("original:", theme.Get(app), retryLimit.Get(app))

	// Output:
	// request: dark 5
	// preview: light 5
	// original: dark 5
}
