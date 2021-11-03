package main

import (
	"fmt"

	"devtools/icmp"
)

func main() {

	v := icmp.Ping(
		icmp.WithDomain("www.google.com"),
		icmp.WithCount(4),
		icmp.WithPs(32),
		icmp.WithEcho(false),
	)

	if v >= 80 {
		fmt.Printf("{{{%f}}}", v)
	}

}
