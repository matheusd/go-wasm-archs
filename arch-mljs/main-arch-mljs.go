package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"syscall/js"
	"time"
)

var closeChan chan struct{}
var stopProcesssing chan struct{}
var currentIndex uint32

func goStep01(thisArg js.Value, args []js.Value) interface{} {
	fmt.Println("goStep01 called", args)
	return rand.Int()
}

func goStep02(thisArg js.Value, args []js.Value) interface{} {
	fmt.Println("goStep02 called", args)
	return rand.Int()
}

func goStep03(thisArg js.Value, args []js.Value) interface{} {
	fmt.Println("goStep03 called", args)

	// A potentially long-running processing step should be run on a separate
	// goroutine
	go func() {
		res := performLong()
		fmt.Println("goStep03 finished processing")
		js.Global().Set("step03_done", res)
	}()
	return nil
}

func goStopProcessing(thisArg js.Value, args []js.Value) interface{} {
	close(stopProcesssing)
	return nil
}

func goReportCurrentIndex(thisArg js.Value, args []js.Value) interface{} {
	return currentIndex
}

func goCloseApp(thisArg js.Value, args []js.Value) interface{} {
	close(closeChan)
	return nil
}

func performLong() bool {
	var bts [4]byte
	for i := uint32(0); i < 1000000; i++ {
		binary.LittleEndian.PutUint32(bts[:], i)
		sha256.Sum256(bts[:])
		currentIndex = i

		// possibly cancel or yield to the main js event loop thread.
		select {
		case <-closeChan:
			return false
		case <-stopProcesssing:
			return false
		default:
			if i%100 == 0 {
				// Without some of this, this function locks up the js
				// processing thread in the browser.
				time.Sleep(time.Nanosecond)
			}
		}
	}
	return true
}

type fnT func(this js.Value, args []js.Value) interface{}

func setFunc(name string, fn fnT) js.Func {
	wrapped := js.FuncOf(fn)
	js.Global().Set(name, wrapped)
	return wrapped
}

func main() {
	closeChan = make(chan struct{})
	stopProcesssing = make(chan struct{})

	fns := make([]js.Func, 0)
	fns = append(fns, setFunc("goStopProcessing", fnT(goStopProcessing)))
	fns = append(fns, setFunc("goCloseApp", fnT(goCloseApp)))
	fns = append(fns, setFunc("goReportCurrentIndex", fnT(goReportCurrentIndex)))
	fns = append(fns, setFunc("goStep01", fnT(goStep01)))
	fns = append(fns, setFunc("goStep02", fnT(goStep02)))
	fns = append(fns, setFunc("goStep03", fnT(goStep03)))

	fmt.Fprintln(os.Stdout, "initialized in stdout")

	// some long init operation....
	time.Sleep(time.Second)

	js.Global().Set("lib_ready", true)
	<-closeChan

	fmt.Fprintln(os.Stdout, "ending execution")

	for _, fn := range fns {
		fn.Release()
	}
}
