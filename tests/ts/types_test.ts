import { Context } from "../../perky.ts";

function typings() {
  const ctx: Context = new Context();
  const Count = Context.key(0);
  const count: number = ctx(Count);
  Count(5);
  // @ts-expect-error Wrong value type.
  Count("five");
  // @ts-expect-error Wrong getter result.
  Count(() => "five");
  const Theme = Context.key<"light" | "dark">(() => "light");
  const theme: "light" | "dark" = ctx(Theme);
  // @ts-expect-error Explicit union is retained.
  Theme("other");
  const fn = (input: string) => input.length;
  const Handler = Context.key(fn, false);
  const handler: typeof fn = ctx(Handler);
  Handler(fn, false);
  Handler(() => fn);
  const FixedZero = Context.key(() => 1, false);
  const fixedZero: () => number = ctx(FixedZero);
  const inferredZero: number = ctx(Context.key(() => 1));
  Context.key(1, false);
  Context.key(1, undefined);
  Count(1, false);
  // @ts-expect-error false is the only explicit flag.
  Context.key(fn, true);
  // @ts-expect-error false is the only explicit flag.
  Handler(fn, true);
  // @ts-expect-error A fixed binding must match the key's value type.
  Count(() => 1, false);
  // @ts-expect-error A callable must be supplied with a false argument.
  Handler(fn);
  const GetterHandler = Context.key(() => fn);
  const inferred: typeof fn = ctx(GetterHandler);
  // @ts-expect-error A function taking arguments is not a getter.
  Context.key(fn);
  const derived: Context = ctx(Count(1), Theme("dark"));
  // @ts-expect-error Derivation requires a binding.
  ctx();
  // @ts-expect-error Reads accept exactly one key.
  ctx(Count, Theme);
  // @ts-expect-error Keys are not constructor bindings.
  new Context(Count);
  return { count, theme, handler, inferred, derived, fixedZero, inferredZero };
}

void typings;
