package main

import (
	"fmt"
)

// @StatusErr[InternalServerError][500100001][InternalServerError]
func call() {
	fmt.Println(InternalServerError)
}

func main() {
	call()
	fmt.Println(Unauthorized)
}
