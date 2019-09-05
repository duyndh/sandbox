package v1

import (
	"context"
	"time"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ngray1747/sandbox/pkg/api/v1"
)

const (
	// Version of the server api
	apiVersion = "v1"
)
// Implement of Todo service proto
type toDoServiceServer struct {
	db *sql.DB
}

// Create a new todoserviceserver 
func NewToDoServiceServer(db *sql.DB) v1.TodoServiceServer {
	return &toDoServiceServer{db: db}
}

// Check if the requested version is supported by the server
func (s *toDoServiceServer) checkAPI(api string) error {
	if len(api) > 0 {
		if apiVersion != api {
			return status.Errorf(codes.Unimplemented, "API version wated '%s' has '%s'", apiVersion, api)
		}
	}
	return nil
}

// Connect to the database and return a database connection from the pool
func (s *toDoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	connection, error := s.db.Conn(ctx)
	
	if error != nil {
		return nil, status.Error(codes.Unknown, "Failed to connect to database, error : " + error.Error())
		
	}
	
	return connection, nil
}

// Create todo task
func (s *toDoServiceServer) Create(ctx context.Context, request *v1.CreateRequest) (*v1.CreateResponse, error) {

	// Check if the requested version is supported
	if error := s.checkAPI(request.Api); error != nil {
		return nil, error
	}

	// Create a database connection
	connection , error := s.connect(ctx)
	if error != nil {
		return nil, error
	}
	// Make sure to close connection
	defer connection.Close()

	// Parse reminder field
	reminder, error := ptypes.Timestamp(request.Todo.Reminder)
	if error != nil {
		return nil, status.Error(codes.InvalidArgument, "The reminder field has incorrect format")
	}


	//Insert into database
	response, error := connection.ExecContext(ctx, "INSERT INTO todo (`title`, `description`, `reminder`) VALUES (?, ?, ?) ", request.Todo.Title, request.Todo.Description, reminder)
	if error != nil {
		return nil, status.Errorf(codes.Unknown, "Failed to insert into database, error: " + error.Error())
	}
	
	// Get ID of the inserted todo
	ID, error := response.LastInsertId()
	if error != nil {
		return nil, status.Errorf(codes.Unknown, "Failed to get the inserted id, erorr: " + error.Error())
	}

	return &v1.CreateResponse{
		Api: apiVersion,
		Id: ID,
	}, nil
}
// Read todo task
func (s *toDoServiceServer) Read(ctx context.Context, request *v1.ReadRequest) (*v1.ReadResponse, error) {
	
	// Check request api
	if error := s.checkAPI(request.Api); error != nil {
		return nil, error
	}
	
	connection, error := s.connect(ctx)
	if error != nil {
		return nil, error
	}
	defer connection.Close()

	// Get the requested to from database
	rows, error := connection.QueryContext(ctx, "SELECT `id`, `title`, `description` , `reminder` FROM `todo` WHERE `id` = ?", request.Id)
	if error != nil {
		return nil, status.Errorf(codes.Unknown, "Failed to get the selected todo, error: " + error.Error())
	}

	defer rows.Close()

	if !rows.Next() {
		if error := rows.Err(); error != nil {
			return nil, status.Errorf(codes.Unknown, "Failed to retrieve data from row, error: " + error.Error())
		}
		return nil, status.Errorf(codes.NotFound, "The requested todo id ('%d') is not found!", request.Id)
	}

	// Map Todo data
	var todo v1.Todo
	var reminder time.Time

	if error := rows.Scan(&todo.Id, &todo.Title, &todo.Description, reminder); error != nil {
		return nil, status.Errorf(codes.Unknown, "Failed to map todo with retrieved data, error: " + error.Error())
	}

	todo.Reminder, error = ptypes.TimestampProto(reminder)
	if error != nil {
		return nil, status.Errorf(codes.Unknown, "Reminder has invalid format, error: " + error.Error())
	}

	if rows.Next() {
		return nil, status.Errorf(codes.Unknown, "Found more multiple data with request id = '%d'", request.Id )
	}

	return &v1.ReadResponse {
		Api: apiVersion,
		Todo: &todo, //? WTF !?
	}, nil
}

// Update todo task
func (s *toDoServiceServer) Update(ctx context.Context, request *v1.UpdateRequest) (*v1.UpdateResponse, error) {

	// Check the requested API version
	if error := s.checkAPI(request.Api); error != nil {
		return nil, error
	}

	connection, error := s.connect(ctx)
	if error != nil {
		return nil, error
	}

	defer connection.Close()

	reminder, error := ptypes.Timestamp(request.Todo.Reminder)
	if error != nil {
		return nil, status.Errorf(codes.Unknown, "Reminder field has invalid format, error: " + error.Error())
	}

	response, error := connection.ExecContext(ctx, "UPDATE `todo` SET `title` = ?, `description` = ?, `reminder` = ? WHERE `id` = ?", request.Todo.Title, request.Todo.Description, reminder, request.Todo.Id)
	if error != nil {
		return nil, status.Errorf(codes.Unknown, "Failed to update todo, error: " + error.Error())
	}

	rows, error := response.RowsAffected()
	if error != nil {
		return nil, status.Errorf(codes.Unknown, "Failed to retrieve effected row(s), error: " + error.Error())
	}

	if rows == 0 {
		return nil, status.Errorf(codes.Unknown, "Requested ID ('%d') not found", request.Todo.Id)
	}

	return &v1.UpdateResponse {
		Api: apiVersion,
		Updated: rows,
	}, nil
}

// Delete todo task
func (s *toDoServiceServer) Delete(ctx context.Context, request *v1.DeleteRequest) (*v1.DeleteResponse, error) {
	//Check if the request api is supported by the server
	if error := s.checkAPI(request.Api); error != nil {
		return nil, error
	}

	// Create a connection to the database
	connection, error := s.connect(ctx)
	if error != nil {
		return nil, error
	}

	defer connection.Close()

	response, error := connection.ExecContext(ctx, "DELETE FROM `todo` WHERE `id` = ?", request.Id)
	if error != nil {
		return nil, status.Errorf(codes.Unknown, "Failed to delete todo, error: " + error.Error())
	}

	rows, error := response.RowsAffected()
	if error != nil {
		return nil, status.Errorf(codes.Unknown, "Failed to get afftected row, error: " + error.Error())
	}

	if rows == 0 {
		return nil, status.Errorf(codes.Unknown, "Requested ID ('%d') not found!", request.Id)
	}

	return &v1.DeleteResponse {
		Api: apiVersion,
		Deleted: rows,
	}, nil
}

// ReadAll todo task
func (s *toDoServiceServer) ReadAll(ctx context.Context, request *v1.ReadAllRequest) (*v1.ReadAllResponse, error) {
	// Check requested API version
	if error := s.checkAPI(request.Api); error != nil {
		return nil, error
	}

	// Get Connection to the database
	connection, error := s.connect(ctx)
	if error != nil {
		return nil, error
	}

	defer connection.Close()

	rows, error := connection.QueryContext(ctx, "SELECT `id`, `title`, `description`, `reminder` FROM `todo`")
	if error != nil {
		return nil, status.Errorf(codes.Unknown, "Failed to get data from database, error: " + error.Error())
	}

	defer rows.Close()

	var reminder time.Time
	listTodo := []*v1.Todo{}
	
	for rows.Next() {
		todo := new(v1.Todo)

		if error := rows.Scan(&todo.Id, &todo.Title, &todo.Description, &reminder); error != nil {
			return nil, status.Errorf(codes.Unknown, "Failed mapping data, error: " + error.Error())
		}

		todo.Reminder, error = ptypes.TimestampProto(reminder)
		if error != nil {
			return nil, status.Errorf(codes.Unknown, "Reminder field has invalid format, error: " + error.Error())
		}

		listTodo = append(listTodo, todo)
		
	}

	return &v1.ReadAllResponse {
		Api: apiVersion,
		Todo: listTodo,
	}, nil
}
