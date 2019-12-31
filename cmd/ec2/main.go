package main

import "github.com/yogin/go-ec2/service"

func main() {
	s := service.NewService()
	s.Run()
}
