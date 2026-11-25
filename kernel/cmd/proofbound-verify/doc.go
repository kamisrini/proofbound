// Command proofbound-verify is the warranty verify page (docs/plans/ui-reviewer-spec.md §2.3).
//
// Built for the browser (GOOS=js GOARCH=wasm) it is the check itself: pbwarranty.PageCheck, the
// warranty package's own Parse and Verify with no repository (W-INV-13). Built natively it assembles
// the page: one self-contained HTML file that carries the compressed WebAssembly, published with each
// release and never inside a bundle, so a bundle cannot vouch for itself.
package main
