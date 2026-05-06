# `pcre2` Module Reference

The `pcre2` module provides regular expression support using the PCRE2 library.

## Global Functions

- `compile(pattern, flags=0)`: Compiles a regex pattern into a `Pattern` object.
- `search(pattern, string, flags=0)`: Searches for the first match of a pattern.
- `match(pattern, string, flags=0)`: Matches a pattern against the beginning of a string.
- `findall(pattern, string, flags=0)`: Returns all matches in a string.
- `sub(pattern, repl, string, count=0, flags=0)`: Replaces occurrences of a pattern.
- `split(pattern, string, maxsplit=0, flags=0)`: Splits a string by occurrences of a pattern.

## Constants

- `PCRE2_ANCHORED`
- `PCRE2_CASELESS` (Alias: `IGNORECASE`, `I`)
- `PCRE2_MULTILINE` (Alias: `MULTILINE`, `M`)
- `PCRE2_DOTALL` (Alias: `DOTALL`, `S`)

## Classes

### `Pattern(pattern, flags=0)`
- `search(subject, pos=0)`: Returns a `Match` object or `None`.
- `match(subject, pos=0)`: Returns a `Match` object if it matches at the start.
- `findall(subject, pos=0)`: Returns all matches as a list.
- `sub(repl, subject, count=0)`: Performs substitutions.
- `split(subject, maxsplit=0)`: Performs splitting.
- `free()`: Frees the compiled regex memory.

### `Match`
- `group(index=0)`: Returns the string matched by the specified group.
- `groups()`: Returns a tuple of all captured groups.
- `start(index=0)`, `end(index=0)`: Returns the start and end offsets of a match.
- `span(index=0)`: Returns `(start, end)`.
- `free()`: Frees match-specific data memory.
