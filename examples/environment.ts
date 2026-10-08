import { Context } from "../perky.ts";

// Application-owned effects can be replaced without changing their global implementations.
export const Now = Context.key(Date.now, false);
export const WriteLine = Context.key<(line: string) => void>(
  console.log,
  false,
);
