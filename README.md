# Perky

Perky is a small library for TypeScript, Go, and C# that lets components declare and use typed
contextual values without coordinating through a shared context definition.

Applications often need to provide settings and services to components deep in their call hierarchy.
Passing these values individually through intermediate functions creates repetitive plumbing, while
maintaining a common context or service interface requires unrelated components to coordinate their
dependencies.

With Perky, each component declares the contextual values it owns as typed keys. Functions receive
an explicit context and use those keys to retrieve values. Callers can supply different values by
deriving new contexts, without changing existing ones or maintaining a central list of dependencies.

## Example: Request processing

Consider an HTTP endpoint that generates a report. The diagnostics module uses a request identifier
for logging and follows the application's current logging level:

```ts
// diagnostics.ts
import { Context } from "@alexdrel/perky";
import { settings } from "./settings";

export const RequestId = Context.key("unknown");
export const LogLevel = Context.key(() => settings.logLevel);

export function log(ctx: Context, message: string, level: "info" | "debug" = "info") {
  if (level === "debug" && ctx(LogLevel) !== "debug") return;
  console.info(`[${ctx(RequestId)}] ${message}`);
}
```

Separately, the formatting module uses a locale to format numbers:

```ts
// formatting.ts
import { Context } from "@alexdrel/perky";

export const Locale = Context.key("en-US");

export function formatNumber(ctx: Context, value: number) {
  return value.toLocaleString(ctx(Locale));
}
```

Neither module needs to know about the other's settings. The report generator uses both components,
passing along the context without declaring their dependencies:

```ts
// reports.ts
import { Context } from "@alexdrel/perky";
import { log } from "./diagnostics";
import { formatNumber } from "./formatting";

export function summarize(ctx: Context, values: number[]) {
  log(ctx, "Generating report", "debug");

  const total = values.reduce((sum, value) => sum + value, 0);
  return `Total: ${formatNumber(ctx, total)}`;
}
```

The endpoint supplies the values associated with each request:

```ts
// endpoint.ts
import { Context } from "@alexdrel/perky";
import { RequestId } from "./diagnostics";
import { Locale } from "./formatting";
import { summarize } from "./reports";

const base = new Context();

export function handle(request: ReportRequest) {
  const ctx = base(
    RequestId(request.id),
    Locale(request.user.locale),
  );

  return summarize(ctx, request.values);
}
```

Each request gets its own context, derived from the same base. Diagnostics and formatting read their
respective values, while the report generator simply passes the context through. The report's data
remains an ordinary function argument.

Not every contextual value needs to be fixed when a context is created. The diagnostics module
defines `LogLevel` using a getter, so logging follows changes to the application's settings without
rebuilding contexts. A caller can still override that value for a particular context using
`LogLevel("debug")`.

The keys belong to the modules that use them, rather than to a shared application interface. If
diagnostics later needs another contextual setting, it can introduce a new key without changing the
report generator or any intermediate function signatures.

Keys are ordinary exported values, so imports, Find All References, and Rename work as expected.
Contexts are immutable and passed explicitly, allowing independent requests, jobs, and tests to use
different values without modifying shared state.

Perky makes contextual values almost as convenient to use as globals, while allowing each request,
job, or test to have its own values without interfering with others.

## Where Perky helps

Perky can serve as a lightweight alternative to dependency injection frameworks for many
applications. Components declare keys for the services they consume, while application
initialization creates and binds the actual implementations. Functions obtain their dependencies
from context without extensive constructor plumbing, central service interfaces, or registration
infrastructure.

Testing is an especially natural fit. Individual tests can substitute services, clocks, or settings
through derived contexts without modifying global state or restoring it afterward. The same
mechanism supports request and background-job processing, where each execution may need its own
environment, and independently developed components that need to share services without coordinating
their declarations.

Contexts can also expose existing configuration through live getters, allowing local overrides
without changing how the underlying settings are managed.

Perky doesn't prescribe which values belong in a context. Settings and services used across multiple
components are natural candidates, while values central to an operation, such as the invoice being
processed or the amount being charged, are generally clearer as explicit parameters. The choice
remains with the developer.

## Getting started

Each implementation follows the same model: a key defines a typed value and its default, while an
explicitly passed context supplies any overrides. The examples below create a context with an
initial binding and pass it to a function that reads the key.

### TypeScript

Install from npm:

```sh
npm install @alexdrel/perky
```

```ts
import { Context } from "@alexdrel/perky";

const Theme = Context.key("light");

function render(ctx: Context) {
  console.log(ctx(Theme));
}

const ctx = new Context(Theme("dark"));
render(ctx); // "dark"
```

`Context.key` creates a typed key with a default. Calling a key produces a binding; calling a
context with a key reads its value. The context constructor accepts initial bindings.

### Go

Add the module to your project:

```sh
go get github.com/alexdrel/perky
```

```go
package main

import (
    "fmt"

    "github.com/alexdrel/perky"
)

var Theme = perky.Key("light")

func render(ctx perky.Context) {
    fmt.Println(Theme.Get(ctx))
}

func main() {
    ctx := perky.New(Theme.Bind("dark"))
    render(ctx) // dark
}
```

Go uses `Key`, `Bind`, `New`, and `Get` rather than TypeScript's callable syntax. This example uses
a standalone `perky.Context`; the same keys also work with Go's standard `context.Context`, as
described later.

### C#

The C# implementation targets .NET 10. Download the `.nupkg` from
[GitHub Releases](https://github.com/alexdrel/perky/releases), or include [Perky.cs](Perky.cs)
directly in your project.

```csharp
using System;
using Perky;

var Theme = Context.Key("light");

var ctx = new Context(Theme.Bind("dark"));
Render(ctx); // dark

void Render(Context context) {
    Console.WriteLine(context.Get(Theme));
}
```

C# uses `Context.Key` to declare keys and `Get` to read them. Shared keys are usually declared as
`static readonly Key<T>` fields in the classes that own them.

In every language, keys without an override resolve to their defaults. Contexts can be passed
through application code without declaring a separate interface listing the keys they contain.

## Keys and bindings

A key belongs to the component that declares it. It has a type, a default, and an identity
independent of its name or value. Other components import or reference that key rather than
redeclaring it or adding it to a shared context interface.

### TypeScript

```ts
// diagnostics.ts
export const RequestId = Context.key("unknown");

// formatting.ts
export const Locale = Context.key("en-US");
export const DisplayMode = Context.key<"compact" | "full">("full");
```

A call such as `RequestId("req-42")` creates a binding. Its type is checked against the key:
`DisplayMode("compact")` is valid, while `DisplayMode("verbose")` is a TypeScript error.

### Go

```go
// package diagnostics
var RequestId = perky.Key("unknown")

// package formatting
var Locale = perky.Key("en-US")
var DisplayMode = perky.Key("full")
```

`RequestId.Bind("req-42")` creates a typed binding. The factory normally infers the type;
`perky.KeyRef[T]` is available when a field, parameter, or result needs to name the key's type
explicitly.

### C#

```csharp
public static class Diagnostics
{
    public static readonly Key<string> RequestId = Context.Key("unknown");
}

public static class Formatting
{
    public static readonly Key<string> Locale = Context.Key("en-US");
}
```

Shared keys are commonly `static readonly` fields. A local `var` declaration works just as well when
the key is private to an operation or test. Bindings use `Diagnostics.RequestId.Bind("req-42")`.

Every key declaration creates a distinct identity, even if two keys have the same type and default.
Importing or referencing an existing key preserves its identity. No registration step is needed, and
a context does not have to know which keys may eventually be read from it.

## Contexts and overrides

Contexts are immutable environments of key bindings. A new context starts with its keys' defaults,
and a derived context inherits its parent's bindings except where explicitly overridden. A context
can receive several bindings at once; if a key appears more than once, the last binding wins.

The following examples give an application a retry policy and derive a context for an incoming
request. The request data remains an ordinary argument to the processing function.

### TypeScript

```ts
const RetryLimit = Context.key(3);
const app = new Context(RetryLimit(5));

function handle(request: ReportRequest) {
  const ctx = app(RequestId(request.id));
  return processRequest(ctx, request.data);
}
```

### Go

```go
var RetryLimit = perky.Key(3)
var app = perky.New(RetryLimit.Bind(5))

func handle(request ReportRequest) {
    ctx := app.With(diagnostics.RequestId.Bind(request.ID))
    processRequest(ctx, request.Data)
}
```

### C#

```csharp
var retryLimit = Context.Key(3);
var app = new Context(retryLimit.Bind(5));

void Handle(ReportRequest request)
{
    var ctx = app.With(Diagnostics.RequestId.Bind(request.Id));
    ProcessRequest(ctx, request.Data);
}
```

Derivation does not modify the application context or any previous request context. The new context
retains inherited bindings such as the retry policy without copying or restating them. Callers can
also create an independent root whenever they need one (`new Context()`, `perky.New()`, or
`new Context()` in C#, respectively).

Bindings are immutable, but the objects they refer to need not be. If a key contains a service or a
mutable object, Perky does not clone or freeze that object. An override changes subsequent key
reads; it does not retroactively change an object that was previously retrieved and retained.

## Fixed and live values

A key can have a fixed value or a getter that supplies its current value on each read. This is
useful for settings already managed elsewhere: Perky exposes them through a key without taking over
their storage. A more local binding can replace a live source with a fixed value, or vice versa.
Getters are not cached.

### TypeScript

```ts
const LogLevel = Context.key(() => settings.logLevel);

const ctx = new Context(LogLevel("debug")); // Fixed override
console.log(ctx(LogLevel)); // "debug"
```

A function passed to `Context.key` or to a key is normally a getter. For example,
`LogLevel(() => preferences.logLevel)` creates a live binding whose getter runs on each read.

### Go

```go
var LogLevel = perky.LiveKey(func() string { return settings.LogLevel })

ctx := perky.New(LogLevel.Bind("debug")) // Fixed override
level := LogLevel.Get(ctx)
```

Go separates fixed and live sources explicitly: `perky.Key` and `Key.Bind` accept fixed values,
while `perky.LiveKey` and `Key.LiveBind` accept getters. For example,
`LogLevel.LiveBind(func() string { return preferences.LogLevel })` creates a live override.

### C#

```csharp
var logLevel = Context.Key(() => Settings.LogLevel);

var ctx = new Context(logLevel.Bind("debug")); // Fixed override
var level = ctx.Get(logLevel);
```

`Context.Key(Func<T>)` and `Key<T>.Bind(Func<T>)` create live sources. For example,
`logLevel.Bind(() => Preferences.LogLevel)` supplies a live override. Getters run on reads, not when
the binding is constructed.

### When the value is a function

TypeScript and C# need to distinguish a function stored _as a value_ from a function called _to
obtain a value_. Go's `Key`/`LiveKey` distinction already makes this explicit.

**TypeScript** uses `false` for a fixed callable:

```ts
const Handler = Context.key(defaultHandler, false);
const ctx = new Context(Handler(customHandler, false));

ctx(Handler)("Hello");
```

**Go** stores functions directly with `Key` and `Bind`:

```go
var Handler = perky.Key(defaultHandler)
ctx := perky.New(Handler.Bind(customHandler))

Handler.Get(ctx)("Hello")
```

**C#** uses the declared delegate type to resolve the overload:

```csharp
var handler = Context.Key<Action<string>>(Console.WriteLine);
var ctx = new Context(handler.Bind((Action<string>)CustomHandler));

ctx.Get(handler)("Hello");
```

For a C# function with no arguments, specify its delegate type to store it as a value:

```csharp
var now = Context.Key(() => DateTimeOffset.UtcNow); // Key<DateTimeOffset>, live
var clock = Context.Key<Func<DateTimeOffset>>(() => DateTimeOffset.UtcNow); // Fixed function
```

A getter returning a function is also supported. In TypeScript, omit `false` and return the function
from the getter; in Go, use `LiveKey` or `LiveBind`; in C#, use a getter returning the delegate
type. Explicit null values are valid bindings and are distinct from missing keys. In C#, use a named
`value: null` argument or a typed cast where a null could otherwise be confused with a null getter.

## Passing contexts

Perky uses explicit context arguments. A function can read its own keys or pass the context on to
another component without declaring all the keys that component needs. This remains ordinary
function calling, including when callbacks or asynchronous operations are involved.

### TypeScript

```ts
async function processBatch(ctx: Context, jobs: Job[]) {
  await Promise.all(jobs.map((job) => processJob(ctx, job)));
}
```

### Go

```go
func processBatch(ctx perky.Context, jobs []Job) {
    for _, job := range jobs {
        processJob(ctx, job)
    }
}
```

### C#

```csharp
Task ProcessBatch(Context ctx, IEnumerable<Job> jobs)
{
    return Task.WhenAll(jobs.Select(job => ProcessJob(ctx, job)));
}
```

No global current-context pointer, thread-local storage, or automatic async propagation is involved.
A function receives the environment its caller chose. This is also why a function's signature need
not change when one of its downstream components introduces another key.

## Go context integration

Go's Perky keys can also be read from the standard `context.Context`, which already travels through
many HTTP and library APIs. `perky.Attach` adds Perky bindings to a native context without changing
its cancellation, deadlines, or unrelated values:

```go
func handle(req *http.Request) {
    native := perky.Attach(
        req.Context(),
        diagnostics.RequestId.Bind(req.Header.Get("X-Request-ID")),
    )
    processRequest(native)

    appCtx := perky.Detach(native)
    render(appCtx)
}

func processRequest(ctx context.Context) {
    log.Printf("Processing %s", diagnostics.RequestId.Get(ctx))
    // Cancellation and deadlines are also available through ctx.
}

func render(ctx perky.Context) {
    fmt.Println(diagnostics.RequestId.Get(ctx))
    // No cancellation or deadline API is available on ctx.
}
```

`Key.Get` works with either context type. `perky.Detach` extracts the Perky bindings without copying
them, evaluating getters, or retaining the native Go context. The resulting `perky.Context` can be
extended with `With`, but deliberately does **not** implement `context.Context`: it has no
cancellation or deadline API. A standalone context can also be constructed directly using
`perky.New`.

Application code decides whether to keep the native Go context or work with the standalone Perky
context. The two types do not need to be passed together.

## Testing

Derived contexts make it possible to replace just the values relevant to a test, leaving application
settings and other tests untouched. A clock is a simple example: production reads the real clock
through a getter, while a test binds a fixed instant.

### TypeScript

```ts
const Now = Context.key(() => Date.now());

const testCtx = app(Now(1_700_000_000_000));
const result = generateReport(testCtx, sampleData);
```

### Go

```go
var Now = perky.LiveKey(time.Now)

fixed := time.Unix(1_700_000_000, 0)
testCtx := app.With(Now.Bind(fixed))
result := generateReport(testCtx, sampleData)
```

### C#

```csharp
var now = Context.Key(() => DateTimeOffset.UtcNow);

var fixedTime = DateTimeOffset.FromUnixTimeSeconds(1_700_000_000);
var testCtx = app.With(now.Bind(fixedTime));
var result = GenerateReport(testCtx, sampleData);
```

A test can similarly replace a logger, database, or other service by binding its key to a fake
implementation. No global overrides need to be installed or restored. Mutable service instances and
getter state remain the application's responsibility, including their behavior under concurrent
access.

## Examples and development

The implementations are small and dependency-free. See the source and tests for the exact runtime
details, including key identity and how inherited bindings are stored.

### TypeScript

Use Deno 2.9 or later for the repository's development commands:

```sh
deno task test:ts   # Type checking, linting, formatting, and runtime tests
deno task build:ts  # Build JavaScript and TypeScript declarations in dist/
deno task pack:ts   # Validate and create an npm package tarball
```

The [document inspector](examples/ts/inspect.ts) and [logger tests](tests/ts/logger_test.ts) show a
larger example with different contextual environments. The package is published as
[`@alexdrel/perky`](https://www.npmjs.com/package/@alexdrel/perky).

### Go

```sh
go test -race ./...
```

The [usage test](tests/go/example_test.go) provides an executable example. The `KeyRef[T]` name is
available for APIs that pass keys around as typed values; the generic alias requires Go 1.24 or
later.

### C#

```sh
dotnet test tests/cs/Perky.Tests.csproj
dotnet run --project examples/cs/Perky.Example.csproj
dotnet build Perky.csproj --configuration Release
deno task pack:cs # Create dist/cs/Perky.<version>.nupkg for a GitHub release
```

The [C# example](examples/cs/Program.cs) demonstrates context-based logging and clock substitution.
The implementation is in [Perky.cs](Perky.cs).

The combined `deno task test`, `deno task build`, and `deno task fmt` commands run across all three
languages.

## Background

Perky grew out of the typed-context design in [KVS](https://github.com/alexdrel/KVS), an
experimental TypeScript dialect, but is a standalone library with implementations in three
languages.
