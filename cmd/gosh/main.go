package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yogin/go-ec2/internal/config"
	"github.com/yogin/go-ec2/internal/service"
)

func main() {
	configPath := flag.String("c", "", "Configuration file path")
	flag.Parse()

	cfg := config.NewConfig(configPath)

	service := service.NewService(cfg)
	if err := service.Run(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}
