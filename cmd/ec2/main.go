package main

import (
	"flag"

	"github.com/yogin/gosh/service"
)

func main() {
	c := service.Config{
		Profile: flag.String("p", "default", "AWS Profile"),
	}
	flag.Parse()
	c.Args = flag.Args()

	s := service.NewService(&c)
	s.Run()
}
