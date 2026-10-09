// Package environment owns the application's clock and output.
package environment

import (
	"fmt"
	"time"

	"github.com/alexdrel/perky"
)

var Now = perky.LiveKey(time.Now)
var WriteLine = perky.Key(func(line string) { fmt.Println(line) })
