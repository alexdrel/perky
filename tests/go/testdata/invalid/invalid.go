// This fixture is expected to fail compilation; Go ignores testdata in ./... discovery.
package invalid

import (
	"context"
	"github.com/alexdrel/perky"
)

var count = perky.Key(0)
var wrongBinding = count.Bind("wrong")
var wrongGetter = count.LiveBind(func() string { return "wrong" })
var wrongNative context.Context = perky.New()
