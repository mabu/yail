package main

import (
	"fmt"
	"github.com/mabu/yail"
	"io/ioutil"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Please pass a file name as an argument.")
		os.Exit(1)
	}
	for i := 1; i < len(os.Args); i++ {
		fmt.Println("Starting program", os.Args[i])
		source, err := ioutil.ReadFile(os.Args[i])
		if err != nil {
			fmt.Println("Could not read file:", err)
			os.Exit(1)
		}
		yail.Interpret(string(source), os.Stdin, os.Stdout)
	}
}
