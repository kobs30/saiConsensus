package main

import (
	"github.com/saiset-co/saiService"
	"github.com/saiset-co/saiVM1/internal"
)

func main() {
	svc := saiService.NewService("sai_VM1")
	svc.RegisterConfig("config.yml")

	is := internal.Service(svc.Context)

	svc.RegisterHandlers(
		is.Handlers(),
	)

	svc.Start()
}
