#!/bin/sh
GOOS=js GOARCH=wasm go build -o mljs.wasm .
