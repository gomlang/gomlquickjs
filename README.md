# gomlquickjs

gomlquickjs is a native GoML port of QuickJS. It does not link to or call the
QuickJS C implementation.

The compatibility baseline is recorded in `UPSTREAM.toml`. The port is built
bottom-up: byte and Unicode primitives, lexical analysis, JavaScript values
and objects, parsing, bytecode compilation, the virtual machine, and built-in
objects.

The current implementation includes the QuickJS extended UTF-8 codec,
Unicode 17 identifier tables, character-range algebra, a recoverable lexical
and Pratt parsing frontend, UTF-16 JavaScript strings and property keys,
generation-checked object heap identities, ordinary objects and property
descriptors, explicit mark-and-sweep collection, bytecode validation and
execution, bytecode and native function calls, source-level functions,
lexical closures and recursion, arrow functions with lexical `this` and
`arguments`, default, rest, and recursive binding patterns across parameters,
separate parameter and function-body environments, per-iteration lexical loop bindings,
declarations, catch clauses, for-in/of, and destructuring assignments with
iterator closing and object-rest exclusion, iterable call and construct spread,
generator functions with `next`, injected `throw`, finally-aware `return`,
ordinary `yield`, and `yield *` completion delegation,
standard source-function length metadata, mapped and unmapped arguments objects
with strict callee poison accessors,
strict-directive propagation and strict/sloppy `this`, property-write, and deletion semantics,
contextual `await` and `yield` grammar across ordinary, async, generator, class, and module goals,
including top-level module await and early errors in parameters and exponentiation,
standard source and built-in function name and length descriptors with bound-name inference,
native arbitrary-precision BigInt arithmetic,
plain, interpolated, and tagged template literals with stable frozen template objects,
and context-sensitive regular-expression literals,
object literals with computed keys, methods and spread, sparse-array literals
with Array, code-point-aware String, and custom iterable spread, accessor execution,
`this`, constructor-specific `new`, `in`, protocol-aware `instanceof` with
Function.prototype Symbol.hasInstance, and a UTF-16
regular-expression engine with captures, backreferences, lookahead, and
lookbehind, including Unicode-aware named captures and named backreferences.
Nested regular-expression groups support locally added or removed `i`, `m`, and `s` modifiers.
Unicode-mode literals, dot atoms, character classes, ranges, and reverse assertions consume
complete code points while preserving UTF-16 capture offsets and lone surrogates.
Unicode property escapes include every Unicode 17 general category, aggregate category, Script,
and Script_Extensions value with canonical short and long aliases, plus every scalar binary
property implemented by upstream QuickJS and all of their canonical aliases.
Ignore-case matching uses the upstream QuickJS Unicode canonicalization tables for literals,
character classes, shorthand complements, and Unicode word boundaries.
String case conversion and NFC, NFD, NFKC, and NFKD normalization use generated upstream
QuickJS Unicode tables, including multi-code-point mappings and contextual final sigma.
Legacy non-Unicode patterns implement Annex B decimal, octal, and identity escapes while
Unicode mode rejects ambiguous numeric escapes.
RegExp match arrays expose null-prototype named `groups`, and String match, matchAll,
search, replace, replaceAll, and split dispatch to
well-known symbol protocols and RegExp state with captures, callback replacements,
named replacement substitutions, the `d` capture-indices flag with named index groups,
and global or sticky indexing. Mutually exclusive alternatives may reuse a capture
name and select the participating capture. RegExp includes QuickJS-compatible
`compile()` state replacement and normalized `toString()` output, and exposes the corresponding Symbol.match,
Symbol.matchAll, Symbol.search, Symbol.replace, and Symbol.split methods. All five symbol methods
use the dynamic RegExpExec path, including overridden `exec` methods and observable result fields.
MatchAll and split honor RegExp species constructors, and MatchAll returns a GC-rooted matcher
iterator with a shared QuickJS-compatible RegExp String Iterator prototype that preserves the source
RegExp lastIndex and advances Unicode empty matches by code point.
RegExp call and construct paths distinguish identity from cloning, preserve derived `newTarget`
prototypes, honor `Symbol.match`, dynamically coerce constructor, input, flags, and `lastIndex`
values, dispatch `test()` through an overridden `exec`, and keep `RegExp.prototype` as the
unbranded sentinel required by QuickJS.
Class declarations and expressions support base and derived
constructors, heritage prototype chains, `super()` and `super` property access,
instance and static methods, getters and setters, and public instance and static
fields, including computed public element names, private fields, methods and accessors with
lexical brands and private-name checks, plus static initialization blocks. The source compiler supports its implemented
declarations, expressions, classic, property-enumeration, and iterable loops, switch, throw, try/catch/finally, and
function subset, including labelled break and continue, continuous optional property,
computed, and call chains with grouped reference and delete semantics,
`new.target`, object-method home objects, and direct-eval inheritance of `this`,
`new.target`, and `super`, including static and runtime-computed eval var propagation into parameter and
function environments, plus nested `with` identifier reads, writes, updates, deletion, lexical shadowing,
and `Symbol.unscopables` across closures and direct eval. Script execution uses a context-persistent
global lexical environment with TDZ, mutable `let`, immutable `const`, nondeletable bindings,
QuickJS-compatible global var and function descriptors, atomic declaration checks, and runtime
identity checks that distinguish direct eval from a shadowing callable. Sloppy direct and indirect
global eval also preflight var and function declarations atomically before updating configurable
global properties. A qjs
command, native shortest-roundtrip ECMAScript number conversion, exact fixed,
exponential, precision, and 2–36 radix Number formatting, and a strict
UTF-16 JSON codec with reviver source contexts, VM-aware replacers, indentation,
accessors, toJSON dispatch, and raw JSON values are also available. URI percent encoding, UTC/ISO Date
arithmetic, SameValueZero Map/Set storage, resizable ArrayBuffer resizing and transfer,
growable SharedArrayBuffer storage, length-tracking and recoverable fixed typed-array views,
DataView memory primitives including 64-bit BigInt access, and QuickJS version-5 binary object serialization
are implemented natively. BJSON covers atom tables, IEEE numbers, UTF-16 strings, BigInt,
ordinary and array objects, Date and primitive wrappers, ArrayBuffer and typed-array views,
and optional cyclic/shared object references. A FIFO job queue and Promise
resolution core support chaining, thenable assimilation, finally, and the
standard combinators over synchronous iterables. Async declarations, expressions, arrows, object methods,
class methods, and exported functions return native Promise objects, suspend at
`await`, resume through the microtask queue, preserve receivers and `super`, and
route rejected awaited values through JavaScript catch/finally handlers. Async generators
queue Promise-returning `next`, `throw`, and `return` requests, distinguish `await` from
`yield`, await yielded values, and delegate to asynchronous or synchronous iterators with
`yield *`. Async-from-sync iterator adapters cache `next`, assimilate yielded values, close on
rejection, and share the intrinsic AsyncIterator prototype with async generators. `for await...of`
supports asynchronous and synchronous sources, declaration and assignment destructuring, and
awaited iterator closing across abrupt completions. The first
built-in slice provides global numeric predicates, primitive conversion
functions, Number and BigInt, primitive wrapper constructors and Object boxing,
VM-aware numeric, string, and property-key object-to-primitive conversion through Symbol.toPrimitive and method fallbacks,
well-known string tags with internal Date, RegExp, Error, arguments, and typed-array brands,
extended Math, Object static primitive coercion, Object.fromEntries and its rooted prototype, Reflect meta-object operations with explicit receivers and newTarget,
Array exotic length semantics, sparse generic callback iteration, forward and reverse reduction, species-aware concat, map, filter, flat, and flatMap, findLast and findLastIndex, fill, copyWithin, stable sort, and toReversed, toSorted, with, and toSpliced copying algorithms, UTF-16 String prototype algorithms,
Array.from and Array.of iterable construction,
String static constructors including `String.raw`, live iterable Map and Set constructors, iterators and forEach, Map and Set grouping, Map and WeakMap upsert methods, Set composition and relation methods, and a rooted Function prototype with call, Array
apply, call-and-construct forwarding bind, source-preserving toString for JavaScript functions,
native-code formatting for host and bound callables, and QuickJS-compatible source file, line,
and column metadata accessors, restricted caller and arguments accessors,
the shared Iterator prototype, Iterator.from and Iterator.concat, lazy map, filter, flatMap, take, and drop helpers,
and every, some, find, forEach, reduce, and toArray consumers across Array, String, Map/Set, and RegExp iterators,
Symbol descriptions and its registry, dynamic Function construction with separately validated parameters and body,
hidden AsyncFunction, GeneratorFunction, and AsyncGeneratorFunction constructors with their
QuickJS prototype topology, dynamic grammar goals, per-function generator instance prototypes,
custom `newTarget` inheritance, and Iterator-prototype generator helpers,
catchable Error subtype objects with preserved explicit throw values and Error.isError branding,
revocable Proxy objects with all thirteen fundamental traps, target invariants, lazy for-in enumeration,
and the primary Reflect and Object ownership, descriptor, prototype, extensibility, and key-enumeration paths,
RegExp with literal escaping, UTC Date helpers, Map/Set,
Promise and RegExp species metadata, ArrayBuffer/SharedArrayBuffer/DataView, single-agent Atomics load/store/read-modify-write/notify/pause operations and nonblocking host rejection for wait, the complete Number-backed typed-array family, and BigInt64Array/BigUint64Array, with live integer-indexed exotic properties,
canonical numeric index handling, buffer aliasing, detachment, reflective descriptors, iterable and
array-like construction, static from/of factories, bulk set, and values/keys/entries iteration,
constructor-only invocation, derived-instance prototype preservation, ArrayBuffer.isView,
the shared `%TypedArray%` constructor and prototype inheritance topology, Uint8Array Base64 and Hex
construction, encoding, and partial-write decoding,
at/join/search, in-place fill/copy/reverse/sort, callback/reduce, species-aware map/filter/slice,
and toReversed/toSorted/with copying algorithms,
UTC and TZif-backed host-local Date instances with construction, DST-aware field access and mutation,
fixed QuickJS locale formatting, ISO and JSON conversion, URI and legacy escape functions,
Math.random, performance.now, qjs scriptArgs, print, console.log, JSON.parse, and JSON.stringify.
The qjs host layer accepts virtual imports of `std`, `os`, and the native GoML
`./bjson.so` compatibility module. Its current native file slice
includes open and temporary files, byte and line access, ArrayBuffer reads and writes, seeking,
formatted output, loadFile, popen, file-descriptor pipes, process status execution, terminal
detection, directory and path metadata, links, time updates, removal, and read/write readiness
handlers. Standard input, output, and error streams, process-local environment mutation and
enumeration, global evalScript, strerror values, append and exclusive opens, descriptor duplication,
renaming, curl-backed text and binary URL retrieval, monotonic time, blocking and Promise-returning
sleep, terminal sizing and raw mode, debug printing, process identifiers, realm-local working-directory
changes shared by file and process APIs, and queued signal handlers interoperate across `std` and `os`.
`os.Worker` provides isolated JavaScript realms, queued message events, structured
cloning with cyclic object support, and shared SharedArrayBuffer backing stores through a
deterministic cooperative scheduler. Promise reactions drain
before timer jobs, including when a reaction requests garbage collection. Unhandled Promise
rejections are tracked through chained and adopted promises, retained across host garbage
collection, and reported at job checkpoints before timer and other host events after dynamic
import jobs and Promise reactions drain. The upstream QuickJS
`tests/test_builtin.js`, `tests/test_std.js`, `tests/test_rw_handler.js`, `tests/test_bjson.js`,
and `tests/test_worker.js` scripts pass against this host layer.
Uncaught JavaScript errors retain source stack information in qjs diagnostics.
The command accepts forced module, script, and strict execution modes, preloaded `-I` scripts,
`--std`, `--no-unhandled-rejection`, and runtime-only `-q` initialization. File execution treats `.mjs` inputs as modules,
otherwise detects module syntax, and uses a path-aware ES module loader with named,
default, namespace, combined, string-named, and side-effect imports, declaration, named,
default, star, and namespace exports with source-order function declaration hoisting,
live binding cells, immutable namespace accessors, relative resolution, star ambiguity checks,
QuickJS-compatible working-directory resolution for bare specifiers, cycle-safe graph
instantiation and evaluation, and Promise-based dynamic import with specifier
coercion, import-option validation, cached namespace identity, original evaluation-error
rejections, and asynchronous settlement. Module graphs support top-level `await` across cyclic
components, start independent asynchronous dependencies before draining jobs, evaluate importers
after their dependencies settle, and propagate original rejection values through static and
dynamic imports. Static and dynamic import attributes accept QuickJS
`javascript` and `json` module types; JSON modules expose the parsed value as their default export.
Each source module also receives a stable null-prototype `import.meta` with QuickJS-compatible
file URL and main-module metadata. Module namespace objects are non-extensible null-prototype
objects tagged as `Module`. The upstream cyclic-import fixture executes
without the native implementation's documented resolution crash workaround.

This is an active compatibility port, not yet a complete QuickJS replacement.
General native binary module loading, externally delivered host signals, process-wide directory mutation for module resolution,
dynamic `TZ` changes after realm creation and Intl locale-sensitive Date options,
locale-sensitive formatting and the remaining regular-expression syntax remain under
development.
Nonblocking processes currently execute synchronously behind emulated process identifiers, and
timer delays are not observed. Workers are cooperatively scheduled rather than backed by operating-system
threads, and transfer lists, blocking Atomics waits, and parallel execution are not implemented. `std.parseExtJSON`
currently evaluates its input through the JavaScript compiler and must not be used as a security
boundary. Garbage collection is deferred while opaque Promise reaction closures remain queued.
Settled Promise values and reasons and pending JavaScript reaction handlers are exposed as heap
edges. Suspended async function frames and pending async generator requests are retained until
their awaited chain or request queue settles because their remaining native captures are not yet
all exposed as precise heap edges.
The Test262 runner parses metadata, loads harness files, handles strict/sloppy,
module, and positive/negative variants, executes synchronous and asynchronous module graphs with
harness setup separately in the same realm, and executes script and module async variants
through a realm-local `$DONE` callback after draining Promise and host jobs. Parse, resolution,
and runtime negatives require matching phases and error types; module errors retain structured
runtime types instead of inferring them from diagnostic text. Fixture files, unsupported proposal
features, blocking-host cases, and Intl402 paths are skipped before execution. The runner does not
yet provide a wall-clock timeout for indefinitely pending async tests. Harness `var` and function
declarations persist through the global object, while top-level harness `let`, `const`, and `class`
declarations persist through the realm's global lexical environment and are visible to subsequent
scripts and modules.

Run the module tests with:

```sh
../stage2/bin/goml test --compiler ../stage2/bin/gomlc
```

Build the current qjs-compatible command with:

```sh
../stage2/bin/goml build --compiler ../stage2/bin/gomlc
_artifact/bin/cmd/qjs/qjs --help
```

Build and run selected Test262 files with:

```sh
../stage2/bin/goml build --compiler ../stage2/bin/gomlc
_artifact/bin/cmd/run_test262/run_test262 --harness /path/to/test262/harness test.js
```
