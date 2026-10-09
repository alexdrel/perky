using Perky;

var count = Context.Key(0);
count.Bind("wrong");
count.Bind(() => "wrong");
