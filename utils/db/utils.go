package dbutils

import (
	utils "bookfinder/utils/common"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

func BulkInsert[T any, M any](tableName string, unsavedRows []T, prepare func(value T) M) error {
	db, err := sql.Open("sqlite3", "db/bookstore.db")
	if err != nil {
		return err
	}
	defer db.Close()

	queryBindings := make([]string, 0, len(unsavedRows))

	typed := reflect.ValueOf(prepare(unsavedRows[0]))
	columnNames := make([]string, 0, typed.NumField())
	utils.LoopFields(typed, func(i int) {
		columnNames = append(columnNames, typed.Type().Field(i).Tag.Get("db"))
	})

	valuesToBind := make([]interface{}, 0, len(unsavedRows)*len(columnNames))
	for _, row := range unsavedRows {
		queryBindings = append(queryBindings, fmt.Sprintf("(%s)", strings.Join(utils.Repeat("?", len(columnNames)), ",")))
		rowValues := reflect.ValueOf(prepare(row))
		values := make([]any, 0, len(columnNames))
		utils.LoopFields(rowValues, func(i int) {
			values = append(values, rowValues.Field(i).Interface())
		})
		valuesToBind = append(valuesToBind, values...)
	}

	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(columnNames, ","),
		strings.Join(queryBindings, ","))
	_, err = db.Exec(stmt, valuesToBind...)

	return err
}
