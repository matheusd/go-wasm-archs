# Architectural Demos for go+wasm

Some possible architectural patterns for go applications compiled into wasm. Read below for further explanations of each architecture.

| Dir       | Pattern |
| --------- | --------|
| arch-mlgo | Main loop in the go application |
| arch-mljs | Main loop in the JS application |


## Running

As of 2019-01-02, you'll need to use either a (1.12) beta version or install go from source, given I'm testing the bleeding edge wasm compilation.

You'll also need to serve the files in each dir using [localserve](https://github.com/matheusd/localserve), since I use the long-running op (`/__sleep__`) feature to simulate performing some remote operation that takes a while to complete.

Enter each arch dir, then:

```
$ ./build.sh
$ localserve
(open https://localhost:8080/test.html in chrome)
```

## arch-mlgo

This architecture is meant for a go command/main app being ported into running on a wasm platform.

The main application logic runs in go, in the main function, while making eventual calls into javascript (for stuff like network access).

When studying this pattern, start by looking at `func main()`, then following the calls into javascript and back.

Pay attention at how JS-side promises are handled by using a simple channel.

Also note that long-running processes on the go side need to have some sort of _yield-like_ structure, so that the main JS event loop thread doesn't completely block and crash the browser.

## arch-mljs

This architecture is meant for a go library being ported into running on a wasm platform.

The main application logic runs in JS, while the go main func is pretty much only used for setup/teardown and holding down the wasm instance object until it is no longer needed by the JS side.

When studying this pattern, start by looking at `async function jsMain()`, then following the calls into go and back.

Note that the long processing operation is now started in a separate goroutine and signals its completion by changing the value of a specific var (this could be refactored into running a second function or resolving a promise such that processing could continue).
