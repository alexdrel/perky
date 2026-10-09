// gofmt lists unformatted files but exits successfully; make that fail checks on every OS.
const result = await new Deno.Command("gofmt", {
  args: ["-l", "perky.go", "tests/go", "examples/go"],
}).output();
await Deno.stdout.write(result.stdout);
await Deno.stderr.write(result.stderr);
Deno.exit(result.code || (result.stdout.length ? 1 : 0));
