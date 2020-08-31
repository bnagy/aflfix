我是光年实验室高级招聘经理。
我在github上访问了你的开源项目，你的代码超赞。你最近有没有在看工作机会，我们在招软件开发工程师，拉钩和BOSS等招聘网站也发布了相关岗位，有公司和职位的详细信息。
我们公司在杭州，业务主要做流量增长，是很多大型互联网公司的流量顾问。公司弹性工作制，福利齐全，发展潜力大，良好的办公环境和学习氛围。
公司官网是http://www.gnlab.com,公司地址是杭州市西湖区古墩路紫金广场B座，若你感兴趣，欢迎与我联系，
电话是0571-88839161，手机号：18668131388，微信号：echo 'bGhsaGxoMTEyNAo='|base64 -D ,静待佳音。如有打扰，还请见谅，祝生活愉快工作顺利。

# afl-fuzz fixup shim

## General Information

This is a skeleton to fixup tests for afl-fuzz >= 1.52. I've used a Go fixer,
but it should work for any language.

## Quickstart

Start fixer
```
ben$ AFL_FIX_SOCK=/tmp/aflfix.sock ./aflfix
2015/04/02 14:48:25 Listening on /tmp/aflfix.sock...
2015/04/02 14:48:25 Starting up. Running Fixer: Startxref 1.0
```

Start afl
```
AFL_POST_LIBRARY=/path/to/aflfix/ext/post_shim.so AFL_FIX_SOCK=/tmp/aflfix.sock ./afl-fuzz [...]
```

Fixer gets a connection. W00t.
```
2015/04/02 14:50:56 Accepted connection!
```

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
