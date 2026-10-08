import { Context } from "../perky.ts";
import { Now, WriteLine } from "./environment.ts";

type Level = "debug" | "info" | "warn";

// The logger owns its settings; the application owns the clock and output.
export const LogLevel = Context.key<Level>("info");
export const Colors = Context.key(true);

const priority = { debug: 0, info: 1, warn: 2 };
const color = { debug: "\x1b[90m", info: "\x1b[36m", warn: "\x1b[33m" };

export function log(ctx: Context, level: Level, message: string) {
  if (priority[level] < priority[ctx(LogLevel)]) return;

  const time = new Date(ctx(Now)()).toISOString();
  const label = level.toUpperCase();
  const styled = ctx(Colors) ? `${color[level]}${label}\x1b[0m` : label;
  ctx(WriteLine)(`${time} ${styled} ${message}`);
}
