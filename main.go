package main

import (
	"cryptofu/bittrex"
	"fmt"
)

func main() {
	fmt.Println("Hello, world")
	api := bittrex.LoadCredentials()
	fmt.Println(api)
}
