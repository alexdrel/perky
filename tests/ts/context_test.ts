import { strictEqual as assertStrictEquals } from "node:assert";
import { test } from "node:test";
import { Context } from "../../perky.ts";

test("defaults, distinct identities, aliases, roots and nullish overrides", () => {
  const A = Context.key(1);
  const B = Context.key(1);
  const alias = A;
  const Nullable = Context.key<number | null | undefined>(3);
  const root = new Context(A(2));
  assertStrictEquals(root(alias), 2);
  assertStrictEquals(root(B), 1);
  assertStrictEquals(new Context()(A), 1);
  assertStrictEquals(root(Nullable(null))(Nullable), null);
  assertStrictEquals(root(Nullable(undefined))(Nullable), undefined);
});

test("nested inheritance, siblings and last binding wins", () => {
  const A = Context.key(0);
  const B = Context.key(false);
  const root = new Context(A(1), A(2));
  const child = root(A(3), A(4));
  const nested = child(B(true));
  assertStrictEquals(root(A), 2);
  assertStrictEquals(child(A), 4);
  assertStrictEquals(nested(A), 4);
  assertStrictEquals(nested(B), true);
  assertStrictEquals(root(B), false);
  assertStrictEquals(root(A(5))(A), 5);
  assertStrictEquals(child(A), 4);
});

test("live sources are lazy, uncached and shadowable", () => {
  let reads = 0;
  const Count = Context.key(() => ++reads);
  const root = new Context();
  const override = Count(() => ++reads * 10);
  const live = root(override);
  assertStrictEquals(reads, 0);
  assertStrictEquals(root(Count), 1);
  assertStrictEquals(root(Count), 2);
  assertStrictEquals(live(Count), 30);
  assertStrictEquals(live(Count(0))(Count), 0);
  assertStrictEquals(root(Count(7))(Count(() => 8))(Count), 8);
});

test("callable values and getter-returned functions", () => {
  const original = (text: string) => text.length;
  const replacement = (text: string) => text.length + 1;
  const Handler = Context.key(original, false);
  const LiveHandler = Context.key(() => replacement);
  const ctx = new Context();
  assertStrictEquals(ctx(Handler), original);
  assertStrictEquals(ctx(Handler(replacement, false))(Handler), replacement);
  assertStrictEquals(ctx(Handler(() => replacement))(Handler)("hi"), 3);
  assertStrictEquals(ctx(LiveHandler), replacement);
  const zero = () => 42;
  const FixedZero = Context.key(zero, false);
  assertStrictEquals(ctx(FixedZero), zero);
  assertStrictEquals(ctx(FixedZero(() => 7, false))(FixedZero)(), 7);
  assertStrictEquals(ctx(FixedZero(() => zero))(FixedZero), zero);
  const object = { count: 1 };
  const Value = Context.key(object, false);
  object.count++;
  assertStrictEquals(ctx(Value), object);
  assertStrictEquals(ctx(Value).count, 2);
});

test("asynchronous closures retain their explicit context", async () => {
  const Mode = Context.key("normal");
  const root = new Context();
  async function read(ctx: Context) {
    await Promise.resolve();
    return ctx(Mode);
  }
  const [a, b] = await Promise.all([
    read(root(Mode("a"))),
    read(root(Mode("b"))),
  ]);
  assertStrictEquals(a, "a");
  assertStrictEquals(b, "b");
  assertStrictEquals(root(Mode), "normal");
});

test("public contexts and keys are frozen; bindings are plain tuples", () => {
  const Count = Context.key(0);
  const ctx = new Context();
  assertStrictEquals(Object.isFrozen(ctx), true);
  assertStrictEquals(Object.isFrozen(ctx(Count(1))), true);
  assertStrictEquals(Reflect.set(ctx, "theme", "dark"), false);
  assertStrictEquals(Object.isFrozen(Count), true);
  assertStrictEquals(Object.isFrozen(Count(1)), false);
  assertStrictEquals(Array.isArray(Count(1)), true);
});
