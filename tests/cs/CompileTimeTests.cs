using System.Diagnostics;
using System.Runtime.CompilerServices;
using Xunit;

namespace PerkyTests;

public class CompileTimeTests
{
    [Fact]
    public async Task IncompatibleBindingsFailToCompile()
    {
        var host = Environment.GetEnvironmentVariable("DOTNET_HOST_PATH") ?? "dotnet";
        var start = new ProcessStartInfo(host)
        {
            RedirectStandardOutput = true,
            RedirectStandardError = true
        };
        start.ArgumentList.Add("build");
        start.ArgumentList.Add(FixturePath());
        start.ArgumentList.Add("--nologo");
        using var process = Process.Start(start)!;
        var stdout = process.StandardOutput.ReadToEndAsync();
        var stderr = process.StandardError.ReadToEndAsync();
        await process.WaitForExitAsync();
        var output = await stdout + await stderr;

        Assert.NotEqual(0, process.ExitCode);
        Assert.Contains("CS1503", output); // A string cannot bind to a numeric key.
        Assert.Contains("CS0029", output); // A getter must return the key's value type.
    }

    private static string FixturePath([CallerFilePath] string source = "") =>
        Path.Combine(Path.GetDirectoryName(source)!, "testdata", "invalid", "Invalid.csproj");
}
