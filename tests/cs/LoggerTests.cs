using Perky;
using PerkyExample;
using Xunit;

namespace PerkyTests;

public class LoggerTests
{
    [Fact]
    public void ATestOverridesTimeOutputAndPolicyWithoutGlobalChanges()
    {
        var lines = new List<string>();
        var ctx = new Context(
            AppEnvironment.Now.Bind(DateTimeOffset.FromUnixTimeSeconds(1)),
            AppEnvironment.WriteLine.Bind(lines.Add),
            Logger.Colors.Bind(false),
            Logger.LogLevel.Bind(Level.Debug));

        Logger.Log(ctx, Level.Debug, "Connecting");
        var quiet = ctx.With(Logger.LogLevel.Bind(Level.Warn));
        Logger.Log(quiet, Level.Info, "Hidden");
        Logger.Log(quiet, Level.Warn, "Retrying");
        Logger.Log(ctx, Level.Info, "Still visible");
        Logger.Log(ctx.With(Logger.Colors.Bind(true)), Level.Warn, "Colored warning");

        Assert.Equal(new[] {
            "1970-01-01T00:00:01.000Z DEBUG Connecting",
            "1970-01-01T00:00:01.000Z WARN Retrying",
            "1970-01-01T00:00:01.000Z INFO Still visible",
            "1970-01-01T00:00:01.000Z \x1b[33mWARN\x1b[0m Colored warning"
        }, lines);
    }
}
