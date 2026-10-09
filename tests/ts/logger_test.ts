import { assertEquals } from "@std/assert";
import { Context } from "../../perky.ts";
import { Now, WriteLine } from "../../examples/ts/environment.ts";
import { Colors, log, LogLevel } from "../../examples/ts/logger.ts";

Deno.test("a logger test supplies its own clock and output", () => {
  const lines: string[] = [];
  const ctx = new Context(
    Now(() => 1_000, false),
    WriteLine((line) => lines.push(line), false),
    Colors(false),
    LogLevel("debug"),
  );

  log(ctx, "debug", "Connecting");
  log(ctx, "info", "Connected");

  assertEquals(lines, [
    "1970-01-01T00:00:01.000Z DEBUG Connecting",
    "1970-01-01T00:00:01.000Z INFO Connected",
  ]);
});

Deno.test("a quieter context inherits the clock and output, leaving its parent intact", () => {
  const lines: string[] = [];
  const ctx = new Context(
    Now(() => 1_000, false),
    WriteLine((line) => lines.push(line), false),
    Colors(false),
  );
  const quiet = ctx(LogLevel("warn"));

  log(quiet, "info", "Hidden");
  log(quiet, "warn", "Retrying");
  log(ctx, "info", "Still visible");

  assertEquals(lines, [
    "1970-01-01T00:00:01.000Z WARN Retrying",
    "1970-01-01T00:00:01.000Z INFO Still visible",
  ]);
});

Deno.test("colors can be enabled for one call's context", () => {
  const lines: string[] = [];
  const ctx = new Context(
    Now(() => 1_000, false),
    WriteLine((line) => lines.push(line), false),
    Colors(false),
  );

  log(ctx(Colors(true)), "warn", "Retrying");
  log(ctx, "warn", "Retrying");

  assertEquals(lines, [
    "1970-01-01T00:00:01.000Z \x1b[33mWARN\x1b[0m Retrying",
    "1970-01-01T00:00:01.000Z WARN Retrying",
  ]);
});
