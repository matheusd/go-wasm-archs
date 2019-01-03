#!/bin/sh
GOOS=js GOARCH=wasm go build -o mlgo.wasm .
