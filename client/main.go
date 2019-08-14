package main

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	mymodel "grpoc/modules/model"
	"grpoc/services/todo"
)

func main() {
	conn, err := grpc.Dial("localhost:3000", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	//call to ToDo
	c := todo.NewToDoServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t := time.Now().In(time.UTC)
	reminder, _ := ptypes.TimestampProto(t)
	pfx := t.Format(mymodel.SQLDatetime)

	reqCreate := todo.CreateRequest{
		Api: "v1",
		ToDo: &todo.ToDo{
			Title:       "Title" + pfx,
			Description: "description" + pfx,
			Reminder:    reminder,
		},
	}

	resCreate, err := c.Create(ctx, &reqCreate)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Response:", resCreate)

	reqAdd := todo.ReadRequest{
		Api: "v1",
		Id:  2,
	}

	resRead, err := c.Read(ctx, &reqAdd)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Response:", resRead)
}
