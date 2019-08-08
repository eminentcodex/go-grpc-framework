package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	mymodel "grpoc/modules/model"
)

const (
	// ToDoTableName
	ToDoTableName = "ToDo"
)

// ToDo ...
type ToDo struct {
	mymodel.Model `db:"-"`
	ID            int64  `db:"id"`
	Title         string `db:"title"`
	Description   string `db:"description"`
	Reminder      string `db:"reminder"`
}

// NewToDo ...
func NewToDo(ctx context.Context, db *sql.DB) (*ToDo, error) {
	if db == nil {
		return nil, errors.New(mymodel.SQLNoDatabaseConnection)
	}
	toDo := ToDo{Model: mymodel.Model{
		DB:        sqlx.NewDb(db, "mysql"),
		L:         0,
		O:         0,
		SortOrder: nil,
		TableName: ToDoTableName,
	}}

	return &toDo, nil
}

// AddTodo ...
func (t *ToDo) AddTodo(title string, desc string, reminder time.Time) (res sql.Result, err error) {
	insertSet := []map[string]interface{}{
		{
			"Title":       title,
			"Description": desc,
			"Reminder":    reminder,
		},
	}

	return t.Insert(insertSet)
}

func (t *ToDo) GetTodoByID(id int64) (todo ToDo, err error) {
	var (
		todos []ToDo
	)




	err = t.Select(&todos, conditions)
	if err != nil {
		return
	}



	return
}
