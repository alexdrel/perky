# Perky

Perky is a small TypeScript library for managing contextual values such as settings, modes, policies, and services.

Each module can declare its own context keys without modifying a shared interface or registering them centrally. Functions receive a `Context` argument and use imported keys to access their values. Callers can create derived contexts that override selected values without affecting the original.

Keys are ordinary TypeScript declarations, so features retain ownership of their settings, and IDE operations such as Find All References and Rename work naturally.

Values can be fixed or supplied by getters, allowing a context to expose settings managed and updated elsewhere.

## Getting started

Perky exports a single runtime symbol, `Context`, which provides key creation, context construction, and the `Context` type.

```ts
import { Context } from "@alexdrel/perky";

const Theme = Context.key<"light" | "dark">("light");
const ShowGrid = Context.key(false);

function render(ctx: Context, page: Page) {
    return renderPage(page, {
        theme: ctx(Theme),
        grid: ctx(ShowGrid),
    });
}

const ctx = new Context();

render(ctx, page);

const preview = ctx(
    Theme("dark"),
    ShowGrid(true),
);

render(preview, page);

ctx(Theme);      // "light"
preview(Theme);  // "dark"
```

A key identifies a contextual value and defines its type and default. Calling a key with a value or getter creates a binding. Calling a context with a key reads that value; calling it with bindings creates a derived context.

| Expression | Meaning |
|---|---|
| `Theme` | Key identity |
| `Theme("dark")` | Fixed binding |
| `Theme(() => "dark")` | Getter binding |
| `ctx(Theme)` | Read |
| `ctx(Theme("dark"))` | Derive |

`new Context(...)` creates an independent context, optionally with initial bindings.

Neither constructing a binding nor deriving a context changes any existing context.

## Keys and ownership

As a project grows, settings and services tend to accumulate. Passing them individually through unrelated functions becomes tedious, while collecting them into a common `Services` or `ApplicationContext` interface requires different features to coordinate changes to the same definition.

Perky avoids the common definition. A key is declared by the module that owns its meaning.

For example, a rendering module might declare:

```ts
// rendering/settings.ts
import { Context } from "@alexdrel/perky";

export const Theme =
    Context.key<"light" | "dark">("light");

export const ShowGrid = Context.key(false);
```

A billing module can independently introduce its own keys:

```ts
// billing/settings.ts
import { Context } from "@alexdrel/perky";

export const RetryLimit = Context.key(3);
export const DryRun = Context.key(false);
```

A logging module might expose its logger through a context key:

```ts
// logging/settings.ts
import { Context } from "@alexdrel/perky";

export const Logger = Context.key(defaultLogger);
```

Neither module needs to know about the others when declaring its keys. Adding a new key doesn't require changing a shared interface, updating constructors, or registering it with a central service.

Consumers can import keys from any of these modules. For example, billing can use its own settings alongside the logger owned by the logging module:

```ts
// billing/charge.ts
import { Context } from "@alexdrel/perky";
import { DryRun, RetryLimit } from "./settings";
import { Logger } from "../logging/settings";

async function charge(ctx: Context, invoice: Invoice) {
    const logger = ctx(Logger);

    if (ctx(DryRun)) {
        logger.info("Simulating charge");
        return simulateCharge(invoice);
    }

    return chargeWithRetries(invoice, ctx(RetryLimit));
}
```

The logging module doesn't need to know which consumers use its key, and the billing module doesn't need to know how the logger was constructed. Both use the same context without sharing a central declaration of its contents.

Each call to `Context.key` creates a distinct identity. Two modules can declare keys with identical names and types without collisions. Importing or aliasing a key preserves its identity.

Because a key is an ordinary exported TypeScript value, the usual IDE operations work on its declaration and references. Find All References can locate reads and explicit bindings without searching for string-based identifiers or consulting a registry.

Keys are not limited to application-wide configuration. They can represent any value that belongs to an execution environment, including temporary modes, request-specific values, and services.

## Contexts

A context is an immutable mapping from keys to value sources.

```ts
const root = new Context();

const normal = root(
    Theme("light"),
    RetryLimit(3),
);

const dark = normal(Theme("dark"));
const cautious = normal(RetryLimit(1));
const both = dark(RetryLimit(1));
```

A derived context inherits its parent's bindings, except where it explicitly overrides them.

```ts
normal(Theme);       // "light"
dark(Theme);         // "dark"
cautious(Theme);     // "light"
both(RetryLimit);    // 1
```

Derivation doesn't modify the parent or its siblings. If the same key is bound more than once during construction, the last binding takes precedence.

An independent context can be created at any time:

```ts
const isolated = new Context(
    RetryLimit(10),
);
```

It starts from the keys' defaults rather than inheriting bindings from another context.

Contexts are ordinary values. They can be stored, passed between functions, returned, or captured by closures. Multiple asynchronous operations can use different contexts without requiring ambient state or special propagation mechanisms.

The immutability applies to bindings. It doesn't freeze objects supplied as values or prevent externally owned state from changing.

## Fixed and live values

A key's default is usually a constant:

```ts
const RetryLimit = Context.key(3);
```

Some settings change during execution, such as a log level or a user's display preferences. For these, a key can use a getter:

```ts
// theme.ts
type ThemeName = "light" | "dark";

let currentTheme: ThemeName = "light";

export const Theme = Context.key<ThemeName>(
    () => currentTheme
);

export function setTheme(theme: ThemeName) {
    currentTheme = theme;
}
```

The getter is evaluated whenever the key is read:

```ts
const ctx = new Context();

ctx(Theme); // "light"

setTheme("dark");

ctx(Theme); // "dark"
```

The type returned by `ctx(Theme)` is still `ThemeName`, not a function or wrapper.

This allows the module owning a setting to retain control over its storage and modification. Perky only provides access to the current value. The source might be a local variable, an object property, or an existing configuration system.

Bindings follow the same convention:

```ts
const fixed = ctx(
    Theme("light"),
);

const live = ctx(
    Theme(() => preferences.theme),
);
```

`fixed(Theme)` always returns `"light"`. `live(Theme)` evaluates its getter each time.

An explicit binding shadows the inherited source, regardless of whether that source is fixed or live. A derived context can therefore replace a live setting with a fixed value, or vice versa.

### Reading live values

Getters aren't cached by Perky. Two reads of the same key may produce different values if the underlying source changes.

When several operations need a consistent value, read it once:

```ts
const theme = ctx(Theme);

renderHeader(theme);
renderBody(theme);
```

Perky doesn't implement subscriptions or reactive updates. A UI framework or state-management library can provide those independently.

### Callable values

Functions passed to `Context.key` or to a key are interpreted as getters. This keeps the common live-setting case concise:

```ts
const LogLevel = Context.key(() => settings.logLevel);

const local = ctx(
    LogLevel(() => preferences.logLevel),
);
```

When the value itself is a function, `false` as the second argument makes that intention explicit:

```ts
const Handler = Context.key(defaultHandler, false);

const specialized = ctx(
    Handler(customHandler, false),
);

specialized(Handler)("Hello");
```

A getter returning a function uses the ordinary form:

```ts
const Handler = Context.key(() => currentHandler);
```

Here `ctx(Handler)` returns the current handler without invoking it.

The rules are the same when declaring defaults and constructing bindings:

| Form | Meaning |
|---|---|
| `Context.key(value)` | Fixed default |
| `Context.key(() => value)` | Getter default |
| `Context.key(fn, false)` | Callable fixed default |
| `Key(value)` | Fixed binding |
| `Key(() => value)` | Getter binding |
| `Key(fn, false)` | Callable fixed binding |

Passing `false` as the second argument is also available for non-callable values, though it normally isn't needed.

## Type safety and defaults

Each key carries its value type.

```ts
const RetryLimit = Context.key(3);

ctx(RetryLimit);       // number
RetryLimit(5);         // valid
RetryLimit("five");    // TypeScript error
```

A getter must also return a compatible value.

Every key has a default. Reading a key that has no binding in the current context or any of its ancestors returns that default.

When absence is meaningful, it can be part of the key's type:

```ts
const Database = Context.key<Database | null>(null);
```

A function that requires a database can check for its presence:

```ts
async function findUser(ctx: Context, id: string) {
    const db = ctx(Database);

    if (!db) {
        throw new Error("Database unavailable");
    }

    return db.findUser(id);
}
```

Perky doesn't track which keys have been explicitly bound in the type of a context. All contexts share the same `Context` type.

This is intentional. The purpose of independent keys is to allow new contextual values to be introduced without modifying the types of existing functions or contexts.

The type system checks values against their keys, but it doesn't prove that every service required by an operation has been configured.

## Passing context

Perky doesn't maintain a global current context. Functions that use or propagate contextual values receive a context explicitly.

```ts
function handleRequest(ctx: Context, request: Request) {
    return processOrder(ctx, request.order);
}
```

A function doesn't have to declare which individual keys it uses. Adding a key to an implementation therefore doesn't change its signature or the signatures of intermediate functions.

Callbacks capture contexts through ordinary JavaScript closures:

```ts
function processBatch(ctx: Context, orders: Order[]) {
    return Promise.all(
        orders.map(order => processOrder(ctx, order))
    );
}
```

The callback retains the particular context it captured, even if it runs later.

No asynchronous context tracking, thread-local storage, or special callback handling is involved.

### What belongs in context?

Context is useful for values determined by the surrounding environment rather than by an individual operation.

For example:

```ts
processOrder(ctx, order);
```

The order is an explicit input. The context may supply the logger, retry policy, database, or other settings used while processing it.

A contextual value can be as small as a boolean mode or as substantial as a service instance. Perky doesn't impose a distinction between configuration and dependencies.

Ordinary parameters remain appropriate when a value is central to the operation itself.

## Testing

Contexts are convenient for testing because a test can override only the values relevant to it.

For a complete small example, run [the document inspector](examples/inspect.ts). It passes a context from the caller through batch inspection and individual document inspection to [the console logger](examples/logger.ts). The logger owns its level and color settings, while [the application environment](examples/environment.ts) owns the clock and output. It prints normal, debug, and warnings-only logging, then the unchanged parent context's behavior and all three log colors. A final run uses the real clock.

```sh
deno task example
```

The [logger tests](examples/logger_test.ts) replace `Date.now` and console output through their own contexts, then derive quieter or colored contexts without changing the originals.

```sh
deno test examples/logger_test.ts
```

```ts
const testCtx = new Context(
    Database(fakeDatabase),
    DryRun(true),
    RetryLimit(0),
);

await processOrder(testCtx, sampleOrder);
```

A test can also derive its environment from an existing context:

```ts
const testCtx = baseCtx(
    Database(fakeDatabase),
);
```

Neither operation changes `baseCtx` or any other context.

There are no global overrides to restore after the test. Mutable service instances remain subject to their own state and lifetime rules, as they would with ordinary dependency injection.

## Relationship to dependency injection

Perky can distribute services without requiring them to be passed individually or stored in a centrally defined service interface.

```ts
const Database = Context.key<Database | null>(null);
const Logger = Context.key(defaultLogger);

const ctx = new Context(
    Database(database),
    Logger(logger),
);
```

Functions retrieve whichever services they use through the context argument.

Unlike a dependency injection container, Perky doesn't create service instances, resolve constructor dependencies, manage scopes of object lifetime, or rebuild object graphs.

Service construction remains ordinary application code. Perky handles only the bindings through which those services are accessed.

This also means overriding a database binding doesn't reconstruct an existing repository object that was created using a different database. It changes subsequent reads of that key, not objects that previously captured its value.

Perky is equally usable for settings that have nothing to do with services. Its abstraction is the contextual value, not the dependency graph.

## Scope

Perky deliberately leaves several responsibilities to other parts of an application.

It doesn't provide configuration persistence, remote settings management, subscriptions, or state mutation. Getter-backed keys can expose values from systems that provide those capabilities.

It doesn't construct or dispose of resources. Context lifetime and resource lifetime are independent.

It doesn't automatically propagate context through function calls. The context argument is explicit.

There is no global key registry, central configuration type, or requirement to declare a function's individual contextual dependencies.

The library is concerned with key identity, typed values, inheritance, and overrides.

---

## Implementation notes

The implementation is small and dependency-free. This section records its runtime and typing behavior.

### Development and packaging

Use Deno 2.9 or later:

```sh
deno task test   # Type checks, lint, formatting, and runtime tests
deno task build  # JavaScript and TypeScript declarations in dist/
deno task pack   # Validate, build, and create an npm tarball
```

The npm package exports an ES module and declarations. It has no runtime dependencies. `npm pack` also runs validation and the build through its `prepack` hook. Package name, version, and metadata live in `package.json`; publication is a separate release step. No JSR publishing configuration is included.

For local Deno use, import `Context` from `./perky.ts`.

### Public API

Export one runtime symbol, `Context`, which is both constructible and usable as a TypeScript type.

```ts
import { Context } from "@alexdrel/perky";

const Mode = Context.key("normal");

const ctx: Context = new Context(
    Mode("debug"),
);

ctx(Mode);              // string
ctx(Mode("normal"));    // Context
```

The public runtime API consists of:

- `Context.key(source)` and `Context.key(value, false)` for creating keys.
- `new Context(...bindings)` for constructing independent contexts.
- `Key(source)` and `Key(value, false)` for constructing bindings.
- `ctx(Key)` for reading.
- `ctx(binding, ...bindings)` for deriving contexts.

A context read accepts exactly one key. Derivation requires at least one binding. Construction accepts zero or more bindings.

`Context.key` is a static callable property containing the key factory, with an optional `false` argument.

Internal helper functions and types need not be exported.

### Key representation

Every key receives a unique symbol when declared. Its identity is independent of its name or value type. Internally, the KVS tuple representation is extended to `[symbol, getterOrNull, value?]`. The callable key wraps its default tuple; bindings use the same tuple shape. Getter bindings omit the value; fixed bindings use a null getter.

A key is a callable object that creates bindings and carries a default source. Its runtime identity and default source are not writable through the public API.

Both defaults and overrides use the same internal representation:

```ts
type Binding<T> = readonly [symbol, (() => T) | null, value?: T];
```

A binding carries the key identity, a getter or null, and an optional fixed value.

Bindings are readonly tuples. Constructing a binding doesn't invoke a getter or modify an existing context.

At runtime, a callable argument is interpreted as a getter. Passing `false` as the second argument bypasses this interpretation and stores its argument directly.

The getter's result type, rather than its function type, is the key's value type.

### Frame representation

Use sparse prototype-linked frames.

A root frame is created with:

```ts
Object.create(null)
```

A derived frame is created with:

```ts
Object.create(parentFrame)
```

Each frame stores only its explicit overrides, indexed by the keys' unique symbols. An inherited binding is found through normal prototype lookup.

When no binding exists in the frame chain, resolution uses the key's default source.

The lookup must distinguish a missing binding from one containing `null` or `undefined`. Both are valid stored values.

Bindings supplied together are installed in order, with later bindings replacing earlier ones for the same key.

Private frames are fully constructed before use and aren't mutated afterward. The public context functions and keys are frozen once created; binding tuples remain plain arrays.

Getter sources are evaluated on each read. Fixed sources return their stored values unchanged.

No registry, key enumeration, default copying, or global current-context pointer is required.

### Callable Context

A context instance is a function closing over its private frame.

TypeScript should expose call signatures equivalent to:

```ts
interface Context {
    <T>(key: Key<T>): T;
    (binding: Binding, ...rest: Binding[]): Context;
}
```

A class/interface declaration merge can give the constructible `Context` export callable instance types. The constructor can return a function created by an internal factory.

The implementation must preserve the intended TypeScript call and construction signatures under strict mode.

Since instances are callable functions rather than ordinary class instances, normal class prototype identity and `instanceof Context` behavior shouldn't be assumed unless explicitly implemented.

### TypeScript inference

The typings need to distinguish source functions from stored callable values.

Expected examples:

```ts
const Count = Context.key(0);
// Key<number>

const Theme = Context.key(() => getTheme());
// Key<ThemeName>

const Handler = Context.key(defaultHandler, false);
// Key<typeof defaultHandler>

const local = ctx(Theme(() => currentTheme));
// Context

const handler = local(Handler);
// typeof defaultHandler
```

For function-valued keys, direct callable arguments must follow getter semantics. Fixed callable values require `false` as the second argument.

TypeScript overloads and inference should agree with the runtime interpretation rather than merely accepting both forms.

Internal `Key` and `Binding` types may be used in the public declarations without requiring users to import or name them.

### Tests

Runtime tests should cover independent key identities, aliases, defaults, fixed and getter bindings, inheritance, nested overrides, duplicate bindings, explicit nullish values, and independent roots.

Include tests for functions stored with a `false` second argument and getters that return functions.

Test closure capture and asynchronous execution with different explicitly passed contexts. Verify that deriving a context never modifies its parent or siblings.

Compile-time tests should cover inferred key types, explicit generic types, invalid bindings, callable-value cases, and the merged callable/constructible `Context` type.

Implementation should use ordinary JavaScript functions, symbols, and prototype-linked objects. Proxies, decorators, reflection, global state, and external dependencies aren't needed.

---

## Background

Perky grew out of the typed-context design in [KVS](https://github.com/alexdrel/KVS), an experimental TypeScript dialect, but is intended to stand independently as an ordinary TypeScript library.
