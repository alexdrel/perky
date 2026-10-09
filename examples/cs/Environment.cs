using Perky;

namespace PerkyExample;

// Application-owned effects are independent of the logger's settings.
public static class AppEnvironment
{
    public static readonly Key<DateTimeOffset> Now = Context.Key(() => DateTimeOffset.UtcNow);
    public static readonly Key<Action<string>> WriteLine = Context.Key<Action<string>>(Console.WriteLine);
}
