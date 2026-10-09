using Perky;
using PerkyExample;

var documents = new[] { (Name: "hello.txt", Text: "Hello from Perky"), (Name: "draft.txt", Text: "") };
var ctx = new Context(AppEnvironment.Now.Bind(DateTimeOffset.FromUnixTimeSeconds(1)), Logger.Colors.Bind(false));

Console.WriteLine("Normal:");
InspectDocuments(ctx);

Console.WriteLine("\nDebug:");
InspectDocuments(ctx.With(Logger.LogLevel.Bind(Level.Debug)));

Console.WriteLine("\nWarnings only:");
InspectDocuments(ctx.With(Logger.LogLevel.Bind(Level.Warn)));

Console.WriteLine("\nNormal again:");
InspectDocuments(ctx);

Console.WriteLine("\nAll colors:");
InspectDocuments(ctx.With(Logger.LogLevel.Bind(Level.Debug), Logger.Colors.Bind(true)));

Console.WriteLine("\nLive clock:");
InspectDocuments(new Context(Logger.LogLevel.Bind(Level.Debug)));

void InspectDocuments(Context context)
{
    foreach (var document in documents) InspectDocument(context, document.Name, document.Text);
}

static void InspectDocument(Context ctx, string name, string text)
{
    Logger.Log(ctx, Level.Debug, $"Checking {name}");
    var words = text.Split((char[]?)null, StringSplitOptions.RemoveEmptyEntries).Length;
    if (words == 0)
    {
        Logger.Log(ctx, Level.Warn, $"{name} is empty");
        return;
    }
    Logger.Log(ctx, Level.Info, $"{name}: {words} words");
}
