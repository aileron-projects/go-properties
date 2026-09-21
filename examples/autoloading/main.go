package main

import (
	"fmt"

	"github.com/aileron-projects/go-properties/autoload"
)

func main() {
	props := autoload.Props
	fmt.Println(props.GetAs[string]("server.port"))
	fmt.Println(props.GetAs[string]("server.listen"))
	fmt.Println(props.GetAs[int]("server.read-timeout"))
}
