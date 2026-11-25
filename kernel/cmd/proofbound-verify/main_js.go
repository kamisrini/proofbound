//go:build js && wasm

package main

import (
	"syscall/js"

	pbwarranty "github.com/kamisrini/proofbound/kernel/internal/warranty"
)

// main exposes check to the page as proofboundVerify(bytes, trustKey) and stays alive.
func main() {
	js.Global().Set("proofboundVerify", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 2 {
			return map[string]any{"state": "INVALID", "reason": "internal: wrong call"}
		}
		data := make([]byte, args[0].Get("length").Int())
		js.CopyBytesToGo(data, args[0])
		r := pbwarranty.PageCheck(data, args[1].String())
		return map[string]any{"state": r.State, "reason": r.Reason, "commit": r.Commit, "change": r.Change, "issuer": r.Issuer, "key": r.Key, "obligations": r.Claims}
	}))
	js.Global().Call("proofboundReady")
	select {}
}
