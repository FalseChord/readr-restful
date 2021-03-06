package models

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	// For NewDB() usage
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// var db *sqlx.DB

// ------------------------------  NULLABLE TYPE DEFINITION -----------------------------

type NullTime struct {
	Time  time.Time
	Valid bool
}

func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

func (nt NullTime) MarshalJSON() ([]byte, error) {
	if nt.Valid {
		return json.Marshal(nt.Time)
	}
	return json.Marshal(nil)
}

func (nt *NullTime) UnmarshalJSON(text []byte) error {
	nt.Valid = false
	txt := string(text)
	if txt == "null" || txt == "" {
		return nil
	}

	t := time.Time{}
	err := t.UnmarshalJSON(text)
	if err == nil {
		nt.Time = t
		nt.Valid = true
	}

	return err
}

// Create our own null string type for prettier marshal JSON format
type NullString sql.NullString

// Scan is currently a wrap of sql.NullString.Scan()
func (ns *NullString) Scan(value interface{}) error {
	// ns.String, ns.Valid = value.(string)
	// fmt.Printf("string:%s\n, valid:%s\n", ns.String, ns.Valid)
	// return nil
	x := sql.NullString{}
	err := x.Scan(value)
	ns.String, ns.Valid = x.String, x.Valid
	return err
}

// Value validate the value
func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return json.Marshal(ns.String)
	}
	return json.Marshal(nil)
}

func (ns *NullString) UnmarshalJSON(text []byte) error {
	ns.Valid = false
	if string(text) == "null" {
		return nil
	}
	if err := json.Unmarshal(text, &ns.String); err == nil {
		ns.Valid = true
	}
	return nil
}

// ----------------------------- END OF NULLABLE TYPE DEFINITION -----------------------------

type Datastore interface {
	Get(item TableStruct) (TableStruct, error)
	Create(item TableStruct) (interface{}, error)
	Update(item TableStruct) (interface{}, error)
	Delete(item TableStruct) (interface{}, error)
}

type DB struct {
	*sqlx.DB
}

type TableStruct interface {
	GetFromDatabase(*DB) (TableStruct, error)
	InsertIntoDatabase(*DB) error
	UpdateDatabase(*DB) error
	DeleteFromDatabase(*DB) error
}

// func InitDB(dataURI string) {
// 	var err error
// 	db, err = sqlx.Open("mysql", dataURI)
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	if err = db.Ping(); err != nil {
// 		log.Panic(err)
// 	}
// }

func NewDB(dbURI string) (*DB, error) {
	db, err := sqlx.Open("mysql", dbURI)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

// Get implemented for Datastore interface below
func (db *DB) Get(item TableStruct) (TableStruct, error) {

	// Declaration of return set
	var (
		result TableStruct
		err    error
	)

	switch item := item.(type) {
	case Member:
		result, err = item.GetFromDatabase(db)
		if err != nil {
			result = Member{}
		}
	case Article:
		result, err = item.GetFromDatabase(db)
		if err != nil {
			result = Article{}
		}
	}
	return result, err
}

func (db *DB) Create(item TableStruct) (interface{}, error) {

	var (
		result TableStruct
		err    error
	)
	switch item := item.(type) {
	case Member:
		err = item.InsertIntoDatabase(db)
	case Article:
		err = item.InsertIntoDatabase(db)
	default:
		err = errors.New("Insert fail")
	}
	return result, err
}

func (db *DB) Update(item TableStruct) (interface{}, error) {

	var (
		result TableStruct
		err    error
	)
	switch item := item.(type) {
	case Member:
		err = item.UpdateDatabase(db)
	case Article:
		err = item.UpdateDatabase(db)
	default:
		err = errors.New("Update Fail")
	}
	return result, err
}

func (db *DB) Delete(item TableStruct) (interface{}, error) {

	var (
		result TableStruct
		err    error
	)
	switch item := item.(type) {
	case Member:
		err = item.DeleteFromDatabase(db)
		if err != nil {
			result = Member{}
		} else {
			result = item
		}
	case Article:
		err = item.DeleteFromDatabase(db)
		if err != nil {
			result = Article{}
		} else {
			result = item
		}
	}
	return result, err
}

func generateSQLStmt(input interface{}, mode string, tableName string) (query string, err error) {

	columns := make([]string, 0)
	// u := reflect.ValueOf(input).Elem()
	u := reflect.ValueOf(input)

	bytequery := &bytes.Buffer{}

	switch mode {
	case "insert":
		fmt.Println("insert")
		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag.Get("db")
			columns = append(columns, tag)
		}

		bytequery.WriteString(fmt.Sprintf("INSERT INTO %s (", tableName))
		bytequery.WriteString(strings.Join(columns, ","))
		bytequery.WriteString(") VALUES ( :")
		bytequery.WriteString(strings.Join(columns, ",:"))
		bytequery.WriteString(");")

		query = bytequery.String()
		err = nil

	case "full_update":

		fmt.Println("full_update")
		var idName string
		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag
			columns = append(columns, tag.Get("db"))

			if tag.Get("json") == "id" {
				idName = tag.Get("db")
			}
		}

		temp := make([]string, len(columns))
		for idx, value := range columns {
			temp[idx] = fmt.Sprintf("%s = :%s", value, value)
		}

		bytequery.WriteString(fmt.Sprintf("UPDATE %s SET ", tableName))
		bytequery.WriteString(strings.Join(temp, ", "))
		bytequery.WriteString(fmt.Sprintf(" WHERE %s = :%s", idName, idName))

		query = bytequery.String()
		err = nil

	case "partial_update":

		var idName string
		fmt.Println("partial")
		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag
			field := u.Field(i).Interface()

			switch field := field.(type) {
			case string:
				if field != "" {
					if tag.Get("json") == "id" {
						fmt.Printf("%s field = %s\n", u.Field(i).Type(), field)
						idName = tag.Get("db")
					}
					columns = append(columns, tag.Get("db"))
				}
			case NullString:
				if field.Valid {
					fmt.Println("valid NullString : ", field.String)
					columns = append(columns, tag.Get("db"))
				}
			case NullTime:
				if field.Valid {
					fmt.Println("valid NullTime : ", field.Time)
					columns = append(columns, tag.Get("db"))
				}

			case bool, int:
				columns = append(columns, tag.Get("db"))
			default:
				fmt.Println("unrecognised format: ", u.Field(i).Type())
			}
		}

		temp := make([]string, len(columns))
		for idx, value := range columns {
			temp[idx] = fmt.Sprintf("%s = :%s", value, value)
		}
		bytequery.WriteString(fmt.Sprintf("UPDATE %s SET ", tableName))
		bytequery.WriteString(strings.Join(temp, ", "))
		bytequery.WriteString(fmt.Sprintf(" WHERE %s = :%s;", idName, idName))

		query = bytequery.String()
		err = nil
	}
	return
}
