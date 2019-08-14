package todo

import (
	"context"
	"database/sql"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/sarulabs/di"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"grpoc/models"
	"grpoc/modules"
	mymodel "grpoc/modules/model"
)

const (
	// apiVersion is version of API is provided by server
	apiVersion = "v1"
)

// toDoServiceServer is implementation of v1.ToDoServiceServer proto interface
type toDoServiceServer struct {
	db *sql.DB
}

// NewToDoServiceServer creates ToDo service
func NewToDoServiceServer(cont *di.Container) ToDoServiceServer {
	db := (*cont).Get(modules.InstDatabase).(*sql.DB)
	return &toDoServiceServer{db: db}
}

// checkAPI checks if the API version requested by client is supported by server
func (s *toDoServiceServer) checkAPI(api string) error {
	// API version is "" means use current version of the service
	if len(api) > 0 {
		if apiVersion != api {
			return status.Errorf(codes.Unimplemented,
				"unsupported API version: service implements API version '%s', but asked for '%s'", apiVersion, api)
		}
	}
	return nil
}

// connect returns SQL database connection from the pool
func (s *toDoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to connect to database-> "+err.Error())
	}
	return c, nil
}

// Create new todo task
func (s *toDoServiceServer) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	var (
		err          error
		todoModel, _ = models.NewToDo(ctx, s.db)
		reminder     time.Time
		id           int64
		res          sql.Result
	)
	// check if the API version requested by client is supported by server
	if err = s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	reminder, err = ptypes.Timestamp(req.ToDo.Reminder)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "reminder field has invalid format-> "+err.Error())
	}

	// insert ToDo entity data
	res, err = todoModel.AddTodo(req.ToDo.Title, req.ToDo.Description, reminder)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to insert into ToDo-> "+err.Error())
	}

	// get ID of creates ToDo
	id, err = res.LastInsertId()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve id for created ToDo-> "+err.Error())
	}

	return &CreateResponse{
		Api: apiVersion,
		Id:  id,
	}, nil
}

// Read todo task
func (s *toDoServiceServer) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	var (
		todoModel, _ = models.NewToDo(ctx, s.db)
		td           ToDo
	)
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// query ToDo by ID
	todo, err := todoModel.GetTodoByID(req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from ToDo-> "+err.Error())
	}

	td.Id = todo.ID
	td.Description = todo.Description
	td.Title = todo.Title
	rem, _ := time.Parse(mymodel.SQLDatetime, todo.Reminder)
	td.Reminder, _ = ptypes.TimestampProto(rem)

	return &ReadResponse{
		Api:  apiVersion,
		ToDo: &td,
	}, nil

}
