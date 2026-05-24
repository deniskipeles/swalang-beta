Here is the reference table mapping Python keywords and essential built-in functions to their corresponding Swahili translations, along with brief descriptions of their purposes.

### Python Keywords (Swahili Translation)

| English Keyword | Swahili Translation | Description |
| :--- | :--- | :--- |
| `False` | `Uongo` | Boolean false value. |
| `None` | `Tupu` | Represents the absence of a value (null). |
| `True` | `Kweli` | Boolean true value. |
| `and` | `na` | Logical AND operator. |
| `as` | `kama` | Creates an alias while importing or in context managers. |
| `assert` | `hakikisha` | Debugging aid that tests a condition. |
| `async` | `sambamba` | Declares an asynchronous function (coroutine). |
| `await` | `subiri` | Pauses execution of a coroutine until the awaited result is ready. |
| `break` | `vunja` | Exits the innermost active loop. |
| `class` | `ainisha` | Defines a new user-defined class. |
| `continue` | `endelea` | Skips the current iteration of a loop and moves to the next. |
| `def` | `fafanua` | Defines a new function. |
| `del` | `futa` | Deletes objects, variables, or elements from a collection. |
| `elif` | `ikiwaingine` | Evaluates another condition if the previous `if` condition was false. |
| `else` | `ingine` | Executes code if all preceding conditions are false. |
| `except` | `isipokuwa` | Catches exceptions raised within a `try` block. |
| `finally` | `mwishowe` | Executes code guaranteed to run after `try` and `except` blocks. |
| `for` | `kwa` | Iterates over items of any sequence or iterable. |
| `from` | `kutoka` | Specifies the module from which to import specific attributes or functions. |
| `global` | `kote` | Declares that a variable inside a function has global scope. |
| `if` | `ikiwa` | Evaluates a conditional expression. |
| `import` | `leta` | Imports modules or dynamic shared libraries into the current namespace. |
| `in` | `katika` | Tests for membership in a sequence or collection. |
| `is` | `ni` | Tests object identity (checks if two variables refer to the same object). |
| `lambda` | `lambda` | Creates an anonymous inline function. |
| `nonlocal` | `siomitaa` | Declares that a variable inside a nested function belongs to the outer enclosing scope. |
| `not` | `sio` | Logical NOT operator. |
| `or` | `au` | Logical OR operator. |
| `pass` | `pita` | Null operation; used as a placeholder statement. |
| `raise` | `ibua` | Explicitly triggers or propagates an exception. |
| `return` | `rudisha` | Exits a function and returns a value. |
| `try` | `jaribu` | Defines a block of code to monitor for exceptions. |
| `while` | `wakati` | Loops as long as a condition remains true. |
| `with` | `pamoja` | Wraps execution of a block with methods defined by a context manager. |
| `yield` | `zalisha` | Pauses a generator function and returns a value to the generator's caller. |

***

### Key Built-in Functions (Swahili Translation)

| English Built-in | Swahili Translation | Description |
| :--- | :--- | :--- |
| `print` | `andika` | Outputs text to the standard output stream. |
| `input` | `ingiza` | Reads a line of input from standard input. |
| `len` | `urefu` | Returns the number of items in an object or sequence. |
| `int` | `nambak` -> `nambakamili` | Converts a compatible value to an integer. |
| `float` | `nambad` -> `nambadesimali` | Converts a compatible value to a floating-point number. |
| `str` | `maneno` | Converts an object into its user-friendly string representation. |
| `bool` | `ukweli` | Converts an object into a Boolean value. |
| `list` | `orodha` | Creates a mutable, ordered list. |
| `dict` | `kamusi` | Creates an associative array of key-value pairs. |
| `tuple` | `safu` | Creates an immutable sequence of elements. |
| `set` | `seti` | Creates a mutable collection of unique, hashable elements. |
| `range` | `pengo` | Generates a sequence of integers with a defined start, stop, and step. |
| `abs` | `thamani_kamili` | Returns the absolute value of a numeric value. |
| `round` | `viringisha` | Rounds a number to a specified number of decimal digits. |
| `pow` | `kipeo` | Calculates the exponentiation of a number, optionally modulo a third number. |
| `divmod` | `mgao_baki` | Returns a tuple containing the quotient and remainder of a division. |
| `sum` | `jumla` | Sums the elements of an iterable. |
| `min` | `kiwango_chini` | Returns the smallest item from an iterable or set of arguments. |
| `max` | `kiwango_juu` | Returns the largest item from an iterable or set of arguments. |
| `isinstance` | `ni_sampuli` | Checks if an object is an instance of a specified class or classes. |
| `issubclass` | `ni_aina_dogo` | Checks if a class is a subclass of another specified class. |
| `id` | `id` | Returns the unique memory address (identity) of an object. |
| `hasattr` | `ina_sifa` | Checks if an object has a named attribute. |
| `getattr` | `pata_sifa` | Retrieves the value of a named attribute from an object. |
| `setattr` | `weka_sifa` | Sets the value of a named attribute on an object. |
| `delattr` | `futa_sifa` | Deletes a named attribute from an object. |
| `callable` | `inayoitika` | Checks if an object is callable (such as a function or method). |
| `type` | `aina` | Returns the type object of a specified value. |
| `help` | `msaidizi` | Invokes the built-in helper utility for an object. |






---





### Special Variables & Lifecycle Dunders

| English Special Name | Swahili Translation | Description |
| :--- | :--- | :--- |
| `self` | `nafsi` | Represents the instance of the class itself. |
| `__init__` | `__anza__` | Constructor method called when an instance of a class is created. |
| `__new__` | `__tengeneza__` | Allocator method called to build the raw instance object before initialization. |
| `__str__` | `__maneno__` | Returns a user-friendly, informal string representation of the object. |
| `__repr__` | `__uwakilishi__` | Returns an formal, unambiguous string representation of the object. |
| `__call__` | `__ita__` | Allows an instance of a class to be called like a standard function. |
| `__len__` | `__urefu__` | Invoked by the built-in `len()` function to determine object length. |
| `__bool__` | `__ukweli__` | Evaluates the truthiness of the object in conditional statements. |
| `__hash__` | `__heshi__` | Calculates an integer hash value, allowing the object to be used as a dictionary key or set element. |

***

### Collection & Context Manager Dunders

| English Special Name | Swahili Translation | Description |
| :--- | :--- | :--- |
| `__contains__` | `__ina__` | Invoked by the `in` and `not in` membership operators. |
| `__iter__` | `__rudia__` | Returns an iterator object for traversing the collection. |
| `__next__` | `__ifuatazo__` | Returns the next item from an active iterator. |
| `__getitem__` | `__pata_kipengele__` | Handles reading elements using bracket notation: `obj[key]`. |
| `__setitem__` | `__weka_kipengele__` | Handles assigning elements using bracket notation: `obj[key] = val`. |
| `__delitem__` | `__futa_kipengele__` | Handles deleting elements using bracket notation: `del obj[key]`. |
| `__enter__` | `__ingia__` | Sets up the runtime context before executing a `with` block. |
| `__exit__` | `__toka__` | Cleans up the runtime context and handles errors after a `with` block exits. |
| `__getattr__` | `__pata_sifa__` | Fallback called when a requested attribute is not found on the object. |
| `__setattr__` | `__weka_sifa__` | Called when setting an attribute value on an object: `obj.attr = val`. |
| `__delattr__` | `__futa_sifa__` | Called when deleting an attribute from an object: `del obj.attr`. |
| `__await__` | `__subiri__` | Returns an iterator used to drive an asynchronous coroutine step-by-step. |

***

### Operator Overloading Dunders (Math & Comparison)

| English Special Name | Swahili Translation | Description |
| :--- | :--- | :--- |
| `__abs__` | `__thamani_kamili__` | Invoked by the built-in `abs()` function. |
| `__round__` | `__viringisha__` | Invoked by the built-in `round()` function. |
| `__add__` / `__radd__` | `__jumlisha__` / `__rjumlisha__` | Implements the addition operator `+` (and its reflected version). |
| `__sub__` / `__rsub__` | `__toa__` / `__rtoa__` | Implements the subtraction operator `-` (and its reflected version). |
| `__mul__` / `__rmul__` | `__zidisha__` / `__rzidisha__` | Implements the multiplication operator `*` (and its reflected version). |
| `__truediv__` | `__gawa__` | Implements the standard division operator `/`. |
| `__mod__` | `__baki__` | Implements the modulo operator `%`. |
| `__pow__` | `__kipeo__` | Implements the exponentiation operator `**`. |
| `__eq__` | `__sawa__` | Implements the equality comparison operator `==`. |
| `__ne__` | `__sio_sawa__` | Implements the inequality comparison operator `!=`. |
| `__lt__` | `__chini_ya__` | Implements the "less than" operator `<`. |
| `__le__` | `__chini_au_sawa__` | Implements the "less than or equal" operator `<=`. |
| `__gt__` | `__juu_ya__` | Implements the "greater than" operator `>`. |
| `__ge__` | `__juu_au_sawa__` | Implements the "greater than or equal" operator `>=`. |

***

### Special Metadata Names

| English Special Name | Swahili Translation | Description |
| :--- | :--- | :--- |
| `__name__` | `__jina__` | Stores the name of the module, class, or function. |
| `__main__` | `__kuu__` | Representing the entry-point script environment during execution. |
| `__file__` | `__faili__` | Stores the file path of the currently executing module. |
| `__doc__` | `__maelezo__` | Stores the docstring (documentation) of an object. |



---



Here is the reference table mapping the remaining standard Python built-in functions (including `ascii`, `format`, `memoryview`, `frozenset`, `complex`, and the mathematical/FFI helpers) to their corresponding Swahili translations and descriptions.

### Additional Standard Built-ins (Swahili Translation)

| English Built-in | Swahili Translation | Description |
| :--- | :--- | :--- |
| `ascii` | `asiki` | Returns a readable string representation of an object, escaping non-ASCII characters. |
| `format` | `umbiza` | Formats a value into a specific representation governed by a format specifier. |
| `frozenset` | `seti_ganda` | Creates an immutable, hashable version of a standard set collection. |
| `memoryview` | `tazama_kumbukumbu` | Creates a memoryview object allowing safe, zero-copy access to the internal buffer of another object. |
| `complex` | `changamano` | Creates a complex number with real and imaginary parts. |
| `bytes` | `baiti` | Creates an immutable sequence of byte integers in the range `[0, 255]`. |
| `bytearray` | `safu_ya_baiti` | Creates a mutable sequence of byte integers in the range `[0, 255]`. |
| `bin` | `bainari` | Converts an integer into its binary representation string prefixed with `"0b"`. |
| `oct` | `oktali` | Converts an integer into its octal representation string prefixed with `"0o"`. |
| `hex` | `heksadesimali` | Converts an integer into its hexadecimal representation string prefixed with `"0x"`. |
| `hash` | `heshi` | Generates the numeric hash value of a hashable object. |
| `callable` | `inayoitika` | Evaluates whether the argument object can be called as a function. |
| `property` | `sifa` | Class decorator used to define managed properties (getter, setter, deleter). |
| `super` | `bora` | Returns a proxy object delegating method calls to parent or sibling classes in the MRO. |
| `staticmethod` | `mbinu_tuli` | Class decorator that converts a method into a static method (no implicit first argument). |
| `classmethod` | `mbinu_aina` | Class decorator that converts a method into a class method (passes class as implicit first argument). |
| `object` | `kitu` | The fundamental base class from which all Swalang/Pylearn classes inherit. |