package wasm

import (
	"fmt"
)

func GenerateLoader(wasmPath string) string {
	return fmt.Sprintf(`
(async function() {
    const go = new Go();
    const result = await WebAssembly.instantiateStreaming(fetch("%s"), go.importObject);
    go.run(result.instance);
})();
`, wasmPath)
}

func GetWasmExecJS() string {
	return `../misc/wasm/wasm_exec.js`
}
