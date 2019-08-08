package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	cont "github.com/sarulabs/di"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"grpoc/modules"
	"grpoc/services/todo"
)

const (
	ConfigKeyAppPort  = "app.port"
)

type App struct {
	server    *grpc.Server
	container cont.Container
}

// NewApp - Creates a new application
func NewApp() (app *App) {
	app = &App{}
	return
}

// GetServer - Get the grpc server instance tied with the application
func (app *App) GetServer() *grpc.Server {
	return app.server
}

// Run - Run prepare the application and start the grpc server.
// It initializes the container with the resource like database, app config , logger etc. and registers the rpc services
func (app *App) Run(ctx context.Context) (err error) {
	var (
		listen net.Listener
	)

	app.server = grpc.NewServer()

	app.container, err = modules.InitContainer()
	if err != nil {
		return
	}

	app.registerServices()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			log.Println("Shutting down grpc server...")
			app.server.GracefulStop()
			<-ctx.Done()
		}
	}()

	port := app.container.Get(modules.InstAppConfig).(*viper.Viper).Get(ConfigKeyAppPort).(int)
	listen, err = net.Listen("tcp", ":"+ fmt.Sprint(port))
	if err != nil {
		return
	}

	log.Printf("Starting gRPC server at port %d", port)

	return app.server.Serve(listen)

}

// registerServices - Register rpc services with the app grpc server
func (app *App) registerServices() {

	todo.RegisterToDoServiceServer(app.server, todo.NewToDoServiceServer(&app.container))
}



