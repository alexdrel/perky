import { Context } from "../perky.ts";
import { Now } from "./environment.ts";
import { Colors, log, LogLevel } from "./logger.ts";

interface Document {
  name: string;
  text: string;
}

function inspectDocument(ctx: Context, document: Document) {
  log(ctx, "debug", `Checking ${document.name}`);
  const words = document.text.trim().split(/\s+/).filter(Boolean).length;
  if (!words) {
    log(ctx, "warn", `${document.name} is empty`);
    return;
  }
  log(ctx, "info", `${document.name}: ${words} words`);
}

function inspectDocuments(ctx: Context, documents: Document[]) {
  for (const document of documents) inspectDocument(ctx, document);
}

const documents = [
  { name: "hello.txt", text: "Hello from Perky" },
  { name: "draft.txt", text: "" },
];

// Use a fixed clock so the displayed output is reproducible. Output stays on the console.
const ctx = new Context(Now(() => 1_000, false), Colors(false));

console.log("Normal:");
inspectDocuments(ctx, documents);

console.log("\nDebug:");
inspectDocuments(ctx(LogLevel("debug")), documents);

console.log("\nWarnings only:");
inspectDocuments(ctx(LogLevel("warn")), documents);

console.log("\nNormal again:");
inspectDocuments(ctx, documents);

console.log("\nAll colors:");
inspectDocuments(ctx(LogLevel("debug"), Colors(true)), documents);

console.log("\nLive clock:");
inspectDocuments(new Context(LogLevel("debug")), documents);
