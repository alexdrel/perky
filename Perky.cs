namespace Perky;

// A null getter means a fixed value, including null or a delegate.
internal readonly record struct Source<T>(Func<T>? Getter, T Value)
{
    internal T Read() => Getter is null ? Value : Getter();
}

/// <summary>A typed, independently owned context key. Declare it with Context.Key.</summary>
public sealed class Key<T>
{
    internal Source<T> Default { get; }

    internal Key(Source<T> source) => Default = source;

    /// <summary>Create a fixed binding, including delegate values.</summary>
    public Binding Bind(T value) => new(this, new Source<T>(null, value));

    /// <summary>Create a live binding. The getter is evaluated on each read.</summary>
    public Binding Bind(Func<T> getter)
    {
        ArgumentNullException.ThrowIfNull(getter);
        return new Binding(this, new Source<T>(getter, default!));
    }

}

/// <summary>An immutable association between a key and a fixed or live value source.</summary>
public sealed class Binding
{
    internal object Key { get; }
    internal object Source { get; }

    internal Binding(object key, object source) => (Key, Source) = (key, source);
}

/// <summary>An explicitly passed environment of typed values and immutable overrides.</summary>
public sealed class Context
{
    private readonly Context? parent;
    private readonly Dictionary<object, Binding> values = new();

    /// <summary>Declare a key with a fixed default. Specify T to store a no-argument function.</summary>
    public static Key<T> Key<T>(T value) => new(new Source<T>(null, value));

    /// <summary>Declare a key whose default getter runs on each read.</summary>
    public static Key<T> Key<T>(Func<T> getter)
    {
        ArgumentNullException.ThrowIfNull(getter);
        return new Key<T>(new Source<T>(getter, default!));
    }


    /// <summary>Create an independent context, optionally supplying initial bindings.</summary>
    public Context(params Binding[] bindings) : this(null, bindings) { }

    private Context(Context? parent, Binding[] bindings)
    {
        this.parent = parent;
        ArgumentNullException.ThrowIfNull(bindings);
        // One private frame holds only explicit overrides; later bindings win.
        foreach (var binding in bindings)
        {
            ArgumentNullException.ThrowIfNull(binding);
            values[binding.Key] = binding;
        }
    }

    /// <summary>Derive a child without changing this context or copying inherited keys.</summary>
    public Context With(params Binding[] bindings) => new(this, bindings);

    /// <summary>Read an override or the key's default, evaluating live getters each time.</summary>
    public T Get<T>(Key<T> key)
    {
        ArgumentNullException.ThrowIfNull(key);
        for (var frame = this; frame is not null; frame = frame.parent)
            if (frame.values.TryGetValue(key, out var binding))
                // Only this key can construct its bindings, so the source retains its type.
                return ((Source<T>)binding.Source).Read();
        return key.Default.Read();
    }
}
