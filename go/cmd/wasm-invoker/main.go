package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bytecodealliance/wasmtime-go"
)

func main() {
	dir, err := ioutil.TempDir("", "out")
	check(err)
	// defer os.RemoveAll(dir)
	stdoutPath := filepath.Join(dir, "stdout")

	engine := wasmtime.NewEngine()

	wasm, err := LoadWasm("wasm.wasm")
	check(err)
	module, err := wasmtime.NewModule(engine, wasm)
	check(err)

	// Create a linker with WASI functions defined within it
	linker := wasmtime.NewLinker(engine)
	err = linker.DefineWasi()
	check(err)

	// Configure WASI imports to write stdout into a file, and then create
	// a `Store` using this wasi configuration.
	wasiConfig := wasmtime.NewWasiConfig()
	wasiConfig.SetStdoutFile(stdoutPath)
	store := wasmtime.NewStore(engine)
	store.SetWasi(wasiConfig)
	instance, err := linker.Instantiate(store, module)
	check(err)

	// Run the function
	nom := instance.GetFunc(store, "_start")
	_, err = nom.Call(store)
	check(err)

	// Print WASM stdout
	out, err := ioutil.ReadFile(stdoutPath)
	check(err)
	fmt.Print(string(out))
}

func LoadWasm(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
