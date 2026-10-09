using System.Globalization;
using Perky;

namespace PerkyExample;

public enum Level { Debug, Info, Warn }

public static class Logger
{
    public static readonly Key<Level> LogLevel = Context.Key(Level.Info);
    public static readonly Key<bool> Colors = Context.Key(true);

    private static readonly string[] Color = ["\x1b[90m", "\x1b[36m", "\x1b[33m"];

    public static void Log(Context ctx, Level level, string message)
    {
        if (level < ctx.Get(LogLevel)) return;
        var timestamp = ctx.Get(AppEnvironment.Now).UtcDateTime
            .ToString("yyyy-MM-dd'T'HH:mm:ss.fff'Z'", CultureInfo.InvariantCulture);
        var label = level.ToString().ToUpperInvariant();
        if (ctx.Get(Colors)) label = $"{Color[(int)level]}{label}\x1b[0m";
        ctx.Get(AppEnvironment.WriteLine)($"{timestamp} {label} {message}");
    }
}
