package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alexdrel/perky"
	"github.com/alexdrel/perky/examples/go/environment"
	"github.com/alexdrel/perky/examples/go/logger"
)

type document struct {
	name string
	text string
}

func inspectDocument(ctx perky.Context, doc document) {
	logger.Log(ctx, logger.Debug, "Checking "+doc.name)
	words := len(strings.Fields(doc.text))
	if words == 0 {
		logger.Log(ctx, logger.Warn, doc.name+" is empty")
		return
	}
	logger.Log(ctx, logger.Info, fmt.Sprintf("%s: %d words", doc.name, words))
}

func inspectDocuments(ctx perky.Context, documents []document) {
	for _, doc := range documents {
		inspectDocument(ctx, doc)
	}
}

func inspectRequest(ctx context.Context, documents []document) {
	// The logger can use the native context directly.
	logger.Log(ctx, logger.Info, "Request accepted")
	// Document inspection needs only Perky values, so release the native context chain.
	inspectDocuments(perky.Detach(ctx), documents)
}

func main() {
	documents := []document{
		{name: "hello.txt", text: "Hello from Perky"},
		{name: "draft.txt", text: ""},
	}

	// Start at a Go request boundary, replacing the clock for reproducible output.
	native := perky.Attach(context.Background(),
		environment.Now.Bind(time.Unix(1, 0)),
		logger.Colors.Bind(false),
	)
	fmt.Println("Native request:")
	inspectRequest(native, documents)

	ctx := perky.Detach(native)
	fmt.Println("\nDebug:")
	inspectDocuments(ctx.With(logger.LogLevel.Bind(logger.Debug)), documents)

	fmt.Println("\nWarnings only:")
	inspectDocuments(ctx.With(logger.LogLevel.Bind(logger.Warn)), documents)

	fmt.Println("\nNormal again:")
	inspectDocuments(ctx, documents)

	fmt.Println("\nAll colors:")
	inspectDocuments(ctx.With(logger.LogLevel.Bind(logger.Debug), logger.Colors.Bind(true)), documents)

	fmt.Println("\nLive clock:")
	inspectDocuments(perky.New(logger.LogLevel.Bind(logger.Debug)), documents)
}
