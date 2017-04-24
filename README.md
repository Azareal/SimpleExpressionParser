# Simple Expression Parser

A parser for parsing simple expressions which can be placed within Discord Bot commands or for incorporating a minimal amount of programming logic in another context. The underlying data is stored as an untyped  string ready for it's body to be lazily evaluated.

This fits with Discord's model where the data provided could be anything, and the algorithm might not know what the type is until a specific command or function requests a specific type and it can check to see whether the data satisfies those requirements.

I'm currently in the process of porting this over from Azareal/SajuukBot and turning it into a dependency on that end.

# Syntax
```
switch("ha")
{
	"ha":"hey there!",
	"hm":"lolol",
	default:"default text"
}
```
...would print... `hey there!`

This is equivalent to:

```
switch("ha")
{
	"ha":"hey there!",
	"hm":"lolol",
	"default text"
}

if(true)
{
	"run the code in this block"
}

if(true): "run the code on this line"

```

# Progress

* Primitives.

** Strings. Complete

** Integers. Incomplete. Integer arithmetic needs to be implemented.

** Floats / Decimals. Still in the planning stages.

** Maps. Incomplete. You can't access a specific key within it yet.

* Control Structures

** If statements. Complete, we now have support for both line capture statements, and block capture statements :)

** If-else statements. A temporary implementation is up. Subject to change.

** Switches. Fully implemented. Arbitrary expressions can be used as labels, and it has a syntax similar to that of a map.

** Loops. This is currently being planned out.

* Operators

** Assignment Operator. Still in the planning stages.

** Concatenation. Complete. Two adjacent items without any seperator will fuse with each other. Is this the best way of doing it? We might need to rethink this.

** AND and OR boolean logic. Complete.

* I/O

** Printing. Loose strings are currently printed directly to the main output, as this is mostly used in my bot, this means the chat channel it's responding to. We might need to rethink this at some point.
