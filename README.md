# afl-fuzz fixup shim

## General Information

This is a skeleton to fixup tests for afl-fuzz >= 1.52. I've used a Go fixer,
but it should work for any language.

### The way everything works:

afl-fuzz will load a fixup library as a .so if you pass it the
`AFL_POST_LIBRARY` environment variable. It will then call afl_postprocess
once per test. If you want to write your fixup code in C then rejoice - you're
done! Just modify the afl example in experimental/post_library. If not, read
on.

The C code in ext/ builds a shim .so that will write each test received from
afl to a long-running unix socket using DJB netstrings, read a response and
return that to afl. It is configured via the `AFL_FIX_SOCK` environment
variable.

More about netstrings: http://cr.yp.to/proto/netstrings.txt

This architecture should allow you to write fixers in any language that can
talk to unix sockets, which should be more or less any language at all.

### Caveats:

- Start your fix server BEFORE you start afl. The shim will die if it can't
  connect.
- Don't be slow. This is on the critical path.
- Don't screw up. If you send the shim a broken netstring it will exit() and
  take afl with it.
- Don't use this at all. It's a bad approach and will do more harm than good
  in almost all cases.

Benchmark for the Go side of a very simple string substitution fixer:
```
$ go test -tags simple -bench=.
PASS
BenchmarkFixup	    200000	     16212 ns/op
ok  	github.com/bnagy/aflfix	3.411s
```

## Bugs

This was not written in the anticipation of anyone else using it.

## Contributing

Fork & pullreq

## License

BSD Style, See LICENSE file for details
(c) Ben Nagy, 2015
