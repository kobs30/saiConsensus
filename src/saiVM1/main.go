package main

import (
	"github.com/saiset-co/saiService"
	"github.com/saiset-co/saiVM1/internal"
)

func main() {
	svc := saiService.NewService("sai_VM1")
	is := internal.Service(svc.Context)

	svc.RegisterConfig("config.yml")

	svc.RegisterHandlers(
		is.Handlers(),
	)

	svc.RegisterInitTask(is.Init)

	svc.Start()
}
