package main

import (
	"fmt"

	"github.com/jlaffaye/ftp"
)

func main() {
	fmt.Println("before dial")
	conn, err := ftp.Dial("filez.home.net.pl:21")
	if err != nil {
		fmt.Println("error:", err.Error())
	}
	fmt.Println("before login")
	err = conn.Login("anonymous", "anonymous")
	if err != nil {
		fmt.Println("error:", err.Error())
	}
	fmt.Println("before list")
	list, err := conn.List("./")
	if err != nil {
		fmt.Println("error:", err.Error())
	}
	fmt.Println("before range")
	for _, l := range list {
		fmt.Println(l.Name)
	}
}
