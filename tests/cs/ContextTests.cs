using Perky;
using Xunit;

namespace PerkyTests;

public class ContextTests
{
    [Fact]
    public void KeysHaveIndependentIdentitiesAndAliasesPreserveIdentity()
    {
        var theme = Context.Key("light");
        var other = Context.Key("light");
        var alias = theme;
        var ctx = new Context(theme.Bind("dark"));

        Assert.Equal("dark", ctx.Get(alias));
        Assert.Equal("light", ctx.Get(other));
        Assert.Equal("light", new Context().Get(theme));
    }

    [Fact]
    public void ChildrenInheritOverridesWithoutChangingParentsOrSiblings()
    {
        var theme = Context.Key("light");
        var retries = Context.Key(3);
        var root = new Context(theme.Bind("dark"), retries.Bind(4), retries.Bind(5));
        var preview = root.With(theme.Bind("light"));
        var cautious = root.With(retries.Bind(1));

        Assert.Equal("light", preview.Get(theme));
        Assert.Equal(5, preview.Get(retries));
        Assert.Equal("dark", cautious.Get(theme));
        Assert.Equal(1, cautious.Get(retries));
        Assert.Equal(0, preview.With(retries.Bind(0)).Get(retries));
        Assert.Equal("dark", root.Get(theme));
        Assert.Equal(5, root.Get(retries));
    }

    [Fact]
    public void GetterDefaultsAndBindingsAreLazyAndReadEveryTime()
    {
        var reads = 0;
        var count = Context.Key(() => ++reads);
        var root = new Context();
        var live = root.With(count.Bind(() => ++reads * 10));

        Assert.Equal(0, reads);
        Assert.Equal(1, root.Get(count));
        Assert.Equal(2, root.Get(count));
        Assert.Equal(30, live.Get(count));
        Assert.Equal(0, live.With(count.Bind(0)).Get(count));
        Assert.Equal(3, reads);
        Assert.Equal(8, root.With(count.Bind(7)).With(count.Bind(() => 8)).Get(count));
    }

    [Fact]
    public void NullOverridesAreDifferentFromMissingBindings()
    {
        var name = Context.Key<string?>("default");
        var ctx = new Context(name.Bind((string?)null));

        Assert.Null(ctx.Get(name));
        Assert.Null(ctx.With().Get(name));
        Assert.Equal("default", new Context().Get(name));
    }

    [Fact]
    public void OverloadsDistinguishGettersFromDelegateValues()
    {
        Func<int> clock = () => 42;
        Key<int> live = Context.Key(clock);
        Key<Func<int>> fixedClock = Context.Key<Func<int>>(clock);
        var ctx = new Context();

        Assert.Equal(42, ctx.Get(live));
        Assert.Same(clock, ctx.Get(fixedClock));
        Assert.Equal(7, ctx.With(fixedClock.Bind(() => 7)).Get(fixedClock)());
        Assert.Same(clock, ctx.With(fixedClock.Bind(() => clock)).Get(fixedClock));

        Func<string, int> handler = text => text.Length;
        Key<Func<string, int>> handlerKey = Context.Key(handler);
        Key<Func<string, int>> getterReturningHandler = Context.Key(() => handler);
        Assert.Same(handler, ctx.Get(handlerKey));
        Assert.Same(handler, ctx.Get(getterReturningHandler));
        Assert.Same(handler, ctx.With(handlerKey.Bind(handler)).Get(handlerKey));
        Assert.Equal(42, ctx.With(live.Bind(clock)).Get(live));
        Action<string> output = _ => { };
        var outputKey = Context.Key(output);
        Assert.Same(output, ctx.Get(outputKey));
    }

    [Fact]
    public void MutableValuesKeepTheirIdentity()
    {
        var items = new List<string>();
        var key = Context.Key(items);
        var ctx = new Context();
        items.Add("visible");

        Assert.Same(items, ctx.Get(key));
        Assert.Equal("visible", Assert.Single(ctx.Get(key)));
    }

    [Fact]
    public async Task AsyncOperationsCaptureIndependentContexts()
    {
        var mode = Context.Key("normal");
        var root = new Context();
        async Task<string> Read(Context ctx)
        {
            await Task.Yield();
            return ctx.Get(mode);
        }

        var results = await Task.WhenAll(Read(root.With(mode.Bind("a"))), Read(root.With(mode.Bind("b"))));
        Assert.Equal(new[] { "a", "b" }, results);
        Assert.Equal("normal", root.Get(mode));
    }

    [Fact]
    public void ConcurrentReadsAndDerivationsLeaveTheParentUntouched()
    {
        var count = Context.Key(0);
        var root = new Context(count.Bind(1));
        Parallel.For(0, 100, value =>
        {
            Assert.Equal(value, root.With(count.Bind(value)).Get(count));
            Assert.Equal(1, root.Get(count));
        });
    }
}
