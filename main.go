package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)
func main() {
	content, err := ioutil.ReadFile("./enron1/ham/0002.1999-12-13.farmer.ham.txt")
	if err != nil {
		panic(err)
	}
	tokens := strings.Split(string(content)," ")

	for i:= range tokens {
	fmt.Println(tokens[i]);
	}
}
