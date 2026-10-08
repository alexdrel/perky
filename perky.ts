// deno-lint-ignore-file no-explicit-any

// Fixed values use a null getter; live values omit the value slot.
type Binding<T = unknown> = readonly [
  id: symbol,
  getter: (() => T) | null,
  value?: T,
];

// A function-valued key takes a getter returning that function, unless false is supplied.
type Input<T> = T extends (...args: any[]) => unknown ? () => T
  : T | (() => T);

interface Key<T> {
  (source: Input<T>, isGetter?: undefined): Binding<T>;
  (value: T, isGetter: false): Binding<T>;
  readonly default: Binding<T>;
}

function keyFunc<T>(getter: () => T, isGetter?: undefined): Key<T>;
function keyFunc<T>(
  value: T extends (...args: any[]) => unknown ? never : T,
  isGetter?: undefined,
): Key<T>;
function keyFunc<T>(value: T, isGetter: false): Key<T>;
function keyFunc<T>(input: T | (() => T), isGetter?: false): Key<T> {
  // Every key has its own identity, shared by its default and all its bindings.
  const id = Symbol();
  function bind(input: T | (() => T), isGetter?: false): Binding<T> {
    return isGetter !== false && typeof input === "function"
      ? [id, input as () => T]
      : [id, null, input as T];
  }
  bind.default = bind(input, isGetter);
  return Object.freeze(bind) as Key<T>;
}

type Frame = { [id: symbol]: Binding };

function context(parent: Frame | null, bindings: readonly Binding[]): Context {
  // One frame holds all supplied bindings; prototype lookup inherits the parent.
  const frame: Frame = Object.create(parent);
  for (const binding of bindings) {
    frame[binding[0]] = binding;
  }

  // The callable context captures this private frame.
  function contextFunc(first: Key<unknown> | Binding, ...rest: Binding[]) {
    if (Array.isArray(first)) {
      // ctx(Key(value), ...): derive one child context containing all bindings.
      return context(frame, [first as Binding, ...rest]);
    }
    // ctx(Key): read an inherited override or the key's default.
    const key = first as Key<unknown>;
    const binding = frame[key.default[0]] ?? key.default;
    // Evaluate live sources on every read; fixed values are returned unchanged.
    return binding[1] ? binding[1]() : binding[2];
  }

  // Freeze the public callable object; its private frame is never mutated after construction.
  return Object.freeze(contextFunc) as Context;
}

/** An explicitly passed environment of independently declared contextual values. */
export interface Context {
  <T>(key: Key<T>): T;
  (binding: Binding, ...rest: Binding[]): Context;
}

export class Context {
  /** Declare a key with a fixed default or a getter evaluated on every read. */
  static readonly key = keyFunc;

  /** Create an independent context, optionally supplying initial bindings. */
  constructor(...bindings: Binding[]) {
    return context(null, bindings);
  }
}
