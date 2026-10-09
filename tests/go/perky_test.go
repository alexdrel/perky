package perky_test

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alexdrel/perky"
)

var _ perky.Values = context.Background()
var _ perky.Values = perky.Context{}

func TestDefaultsAndIdentity(t *testing.T) {
	theme := perky.Key("light")
	other := perky.Key("light")
	alias := theme
	ctx := perky.New(theme.Bind("dark"))
	if theme.Get(ctx) != "dark" || alias.Get(ctx) != "dark" || other.Get(ctx) != "light" {
		t.Fatal("keys must have independent reference identities; aliases preserve identity")
	}
	if theme.Get(perky.Context{}) != "light" || theme.Get(context.Background()) != "light" {
		t.Fatal("empty contexts must resolve defaults")
	}
	if ctx.Value("unrelated") != nil || ctx.Value([]int{1}) != nil {
		t.Fatal("standalone contexts must expose only Perky values")
	}
}

func TestInheritance(t *testing.T) {
	theme := perky.Key("light")
	retries := perky.Key(3)
	root := perky.New(theme.Bind("dark"), retries.Bind(4), retries.Bind(5))
	preview := root.With(theme.Bind("light"))
	cautious := root.With(retries.Bind(1))
	nested := preview.With(retries.Bind(0))
	if retries.Get(root) != 5 || theme.Get(preview) != "light" || retries.Get(preview) != 5 {
		t.Fatal("inheritance or last-binding-wins failed")
	}
	if theme.Get(cautious) != "dark" || retries.Get(cautious) != 1 || retries.Get(nested) != 0 {
		t.Fatal("nested or sibling overrides failed")
	}
	if theme.Get(root) != "dark" || retries.Get(root) != 5 || theme.Get(perky.New()) != "light" {
		t.Fatal("derivation must not modify parents or independent roots")
	}
}

func TestAttachAndDetach(t *testing.T) {
	type requestKey struct{}
	deadline := time.Now().Add(time.Hour)
	parent, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	parent = context.WithValue(parent, requestKey{}, "request")
	theme := perky.Key("light")
	retries := perky.Key(3)
	first := perky.Attach(parent, theme.Bind("dark"))
	native := perky.Attach(first, retries.Bind(5), retries.Bind(6))
	standalone := perky.Detach(native)
	if theme.Get(native) != "dark" || theme.Get(standalone) != "dark" || retries.Get(standalone) != 6 {
		t.Fatal("Attach and Detach must preserve all Perky bindings")
	}
	if retries.Get(first) != 3 || native.Value(requestKey{}) != "request" {
		t.Fatal("Attach must preserve native values without changing its parent")
	}
	if got, ok := native.Deadline(); !ok || !got.Equal(deadline) {
		t.Fatal("Attach must preserve deadlines")
	}
	if standalone.Value(requestKey{}) != nil {
		t.Fatal("Detach must discard unrelated native values")
	}
	cancel()
	if native.Err() != context.Canceled || native.Done() != parent.Done() {
		t.Fatal("Attach must preserve cancellation")
	}
	if theme.Get(standalone.With(retries.Bind(1))) != "dark" || retries.Get(standalone) != 6 {
		t.Fatal("detached contexts must remain usable independently after cancellation")
	}
	if theme.Get(perky.Detach(context.Background())) != "light" || theme.Get(perky.Attach(parent)) != "light" {
		t.Fatal("contexts without bindings must retain defaults")
	}
}

func TestLiveValues(t *testing.T) {
	reads := 0
	count := perky.LiveKey(func() int { reads++; return reads })
	root := perky.New()
	live := root.With(count.LiveBind(func() int { reads++; return reads * 10 }))
	detached := perky.Detach(perky.Attach(context.Background(), count.LiveBind(func() int { reads++; return reads * 100 })))
	if reads != 0 {
		t.Fatal("construction and detachment must not invoke getters")
	}
	if count.Get(root) != 1 || count.Get(root) != 2 || count.Get(live) != 30 || count.Get(detached) != 400 {
		t.Fatal("getters must run on each read")
	}
	if count.Get(live.With(count.Bind(0))) != 0 || reads != 4 {
		t.Fatal("fixed bindings must shadow live sources")
	}
	if count.Get(root.With(count.Bind(7)).With(count.LiveBind(func() int { return 8 }))) != 8 {
		t.Fatal("live bindings must shadow fixed sources")
	}
}

func TestNilAndFunctionValues(t *testing.T) {
	number := 3
	pointer := perky.Key(&number)
	anything := perky.Key[any]("default")
	ctx := perky.New(pointer.Bind(nil), anything.Bind(nil))
	if pointer.Get(ctx) != nil || anything.Get(ctx) != nil || pointer.Get(perky.New()) != &number {
		t.Fatal("explicit nil must not be mistaken for a missing binding")
	}
	now := perky.Key(func() int { return 1 })
	if now.Get(ctx)() != 1 || now.Get(ctx.With(now.Bind(func() int { return 2 })))() != 2 {
		t.Fatal("functions are fixed values unless explicitly supplied as getters")
	}
	liveFunction := perky.LiveKey(func() func() int { return func() int { return 3 } })
	if liveFunction.Get(ctx)() != 3 {
		t.Fatal("live getters may return functions")
	}
}

func TestConcurrentDerivation(t *testing.T) {
	count := perky.Key(0)
	root := perky.New(count.Bind(1))
	var workers sync.WaitGroup
	for i := 0; i < 32; i++ {
		workers.Add(1)
		go func(value int) {
			defer workers.Done()
			for j := 0; j < 100; j++ {
				if count.Get(root.With(count.Bind(value))) != value || count.Get(root) != 1 {
					t.Error("concurrent derivation changed another context")
					return
				}
			}
		}(i)
	}
	workers.Wait()
}

func TestInvalidTypes(t *testing.T) {
	output, err := exec.Command("go", "test", "./testdata/invalid").CombinedOutput()
	if err == nil {
		t.Fatal("invalid binding types and native-context use must fail to compile")
	}
	for _, message := range []string{"as int value", "does not implement context.Context"} {
		if !strings.Contains(string(output), message) {
			t.Fatalf("missing compiler rejection %q:\n%s", message, output)
		}
	}
}
