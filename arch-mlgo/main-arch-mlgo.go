package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"syscall/js"
	"time"
)

var closeChan chan struct{}
var currentIndex uint32
var step02Resolved chan struct{}

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

func goReportCurrentIndex(thisArg js.Value, args []js.Value) interface{} {
	return currentIndex
}

func goStopProcessing(thisArg js.Value, args []js.Value) interface{} {
	close(closeChan)
	return nil
}

func goResolveStep02(thisArg js.Value, args []js.Value) interface{} {
	close(step02Resolved)
	return nil
}

type fnT func(this js.Value, args []js.Value) interface{}

func setFunc(name string, fn fnT) js.Func {
	wrapped := js.FuncOf(fn)
	js.Global().Set(name, wrapped)
	return wrapped
}

func main() {
	closeChan = make(chan struct{})
	step02Resolved = make(chan struct{})

	fmt.Fprintln(os.Stdout, "initialized in stdout")

	fns := make([]js.Func, 0)
	fns = append(fns, setFunc("goReportCurrentIndex", fnT(goReportCurrentIndex)))
	fns = append(fns, setFunc("goStopProcessing", fnT(goStopProcessing)))
	fns = append(fns, setFunc("goResolveStep02", fnT(goResolveStep02)))

	v := js.Global().Get("initial_value")

	// step01 is a simple sync call into js
	js.Global().Set("current_step", "step 1")
	args := []interface{}{"step01", v}
	res := js.Global().Call("jsStep01", args)

	// step 02 is a call that generates a promise. It only proceeds
	// after the future is resolved (this should also handle cancellation
	// via the closeChan, but omitted for brevity)
	js.Global().Set("current_step", "step 2")
	args = []interface{}{"step02", v, res}
	js.Global().Call("jsStep02", args)
	<-step02Resolved

	// potentially perform some long processing on the go side
	js.Global().Set("current_step", "step 3")
	args = []interface{}{"step03", v}
	if !performLong() {
		js.Global().Set("current_step", "failed")
		js.Global().Call("jsStepFailed", args)
	} else {
		js.Global().Call("jsStep03", args)
	}

	// done!
	js.Global().Set("current_step", "Done!")
	fmt.Fprintln(os.Stdout, "ending execution")

	for _, fn := range fns {
		fn.Release()
	}
}
