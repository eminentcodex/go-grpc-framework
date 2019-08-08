package main

import (
	"context"
	"log"
	"os"

	"grpoc/app"
)

func main() {
	application := app.NewApp()

	ctx := context.Background()

	if err := application.Run(ctx); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
