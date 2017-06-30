package main

import (
	"runtime"

	"inspector"
)

var fileDir, port string

func init() {
	fileDir = "../file/"
	port = ":8000"
    runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	inspector.Start(fileDir, port)
}

