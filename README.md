# Simple Expression Parser

A parser for parsing simple expressions which can be placed within Discord Bot commands or for incorporating a minimal amount of programming logic in another context. The underlying data is stored as an untyped  string ready for it's body to be lazily evaluated.

This fits with Discord's model where the data provided could be anything, and the algorithm might not know what the type is until a specific command or function requests a specific type and it can check to see whether the data satisfies those requirements.

I'm currently in the process of porting this over from Azareal/SajuukBot and turning it into a dependency on that end.
