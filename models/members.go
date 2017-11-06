package models

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

type Member struct {
	ID       string     `json:"id" db:"user_id"`
	Name     NullString `json:"name" db:"name"`
	Nickname NullString `json:"nickname" db:"nick"`
	// Cannot parse Date format
	Birthday NullTime   `json:"birthday" db:"birthday"`
	Gender   NullString `json:"gender" db:"gender"`
	Work     NullString `json:"occupation" db:"work"`
	Mail     NullString `json:"mail" db:"mail"`

	RegisterMode NullString `json:"register_mode" db:"register_mode"`
	SocialID     NullString `json:"social_id,omitempty" db:"social_id"`
	CreateTime   NullTime   `json:"created_at" db:"create_time"`
	UpdatedAt    NullTime   `json:"updated_at" db:"updated_at"`
	UpdatedBy    NullString `json:"updated_by" db:"updated_by"`
	Password     NullString `json:"-" db:"password"`
	// Ignore password JSON marshall for now

	Description  NullString `json:"description" db:"description"`
	ProfileImage NullString `json:"profile_image" db:"profile_picture"`
	Identity     NullString `json:"identity" db:"identity"`

	CustomEditor bool `json:"custom_editor" db:"c_editor"`
	HideProfile  bool `json:"hide_profile" db:"hide_profile"`
	ProfilePush  bool `json:"profile_push" db:"profile_push"`
	PostPush     bool `json:"post_push" db:"post_push"`
	CommentPush  bool `json:"comment_push" db:"comment_push"`
	Active       bool `json:"active" db:"active"`
}

<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
// May have to re-implement with interface to multiple table implementation
func makeSQL(m *member, mode string) (query string, err error) {
=======
=======
func (m *Member) Get() (interface{}, error) {
=======
func (db *DB) Get(id string) (interface{}, error) {
>>>>>>> models using interface for db instances. new db.go file created. implemented TestGet*Member functions

	member := Member{}
	// err := db.QueryRowx("SELECT work, birthday, description, register_mode, social_id, c_editor, hide_profile, profile_push, post_push, comment_push, user_id, name, nick, create_time, updated_at, gender, mail, updated_by, password, active, profile_picture, identity FROM members where user_id = ?", userID).StructScan(&member)
	err := db.QueryRowx("SELECT * FROM members where user_id = ?", id).StructScan(&member)

	switch {
	case err == sql.ErrNoRows:
		log.Printf("User Not Found")
		err = errors.New("User Not Found")
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Successful get user:%s\n", id)
	}
	return member, err
}

func (db *DB) Create(item interface{}) (interface{}, error) {

	member := item.(Member)
	query, _ := makeSQL(&member, "insert")
	result, err := db.NamedExec(query, member)
	// Cannot handle duplicate insert, crash
	if err != nil {
		log.Fatal(err)
		return Member{}, err
	}
	return result, nil
}

func (db *DB) Update(item interface{}) (interface{}, error) {

	member := item.(Member)
	query, _ := makeSQL(&member, "partial_update")
	result, err := db.NamedExec(query, member)

	if err != nil {
		log.Fatal(err)
		return Member{}, err
	}
	return result, nil
}

func (db *DB) Delete(id string) (interface{}, error) {

	result, err := db.Exec("UPDATE members SET active = 0 WHERE user_id = ?", id)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return result, nil
}

>>>>>>> Used global db instance in models package. Member struct implemented interface.
func makeSQL(m *Member, mode string) (query string, err error) {
>>>>>>> seperate http handler from models

	columns := make([]string, 0)
	u := reflect.ValueOf(m).Elem()

	bytequery := &bytes.Buffer{}

	switch mode {
	case "insert":
		fmt.Println("insert")
		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag.Get("db")
			columns = append(columns, tag)
		}

		bytequery.WriteString("INSERT INTO members ( ")
		bytequery.WriteString(strings.Join(columns, ","))
		bytequery.WriteString(") VALUES ( :")
		bytequery.WriteString(strings.Join(columns, ",:"))
		bytequery.WriteString(")")

		// fmt.Println(query)
		query = bytequery.String()
		err = nil

	case "full_update":

		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag.Get("db")
			columns = append(columns, tag)
		}

		temp := make([]string, len(columns))
		for idx, value := range columns {
			temp[idx] = fmt.Sprintf("%s = :%s", value, value)
		}
		bytequery.WriteString("UPDATE members SET ")
		bytequery.WriteString(strings.Join(temp, ", "))
		bytequery.WriteString(" WHERE user_id = :user_id")

		query = bytequery.String()
		err = nil

	case "partial_update":

		fmt.Println("partial")
		for i := 0; i < u.NumField(); i++ {
			tag := u.Type().Field(i).Tag
			field := u.Field(i).Interface()

			switch field := field.(type) {
			case string:
				if field != "" {
					fmt.Printf("%s field = %s\n", u.Field(i).Type(), field)
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

			case bool:
				fmt.Println("bool type", field)
				columns = append(columns, tag.Get("db"))
			default:
				fmt.Println("unrecognised format: ", u.Field(i).Type())
			}
		}

		temp := make([]string, len(columns))
		for idx, value := range columns {
			temp[idx] = fmt.Sprintf("%s = :%s", value, value)
		}
		bytequery.WriteString("UPDATE members SET ")
		bytequery.WriteString(strings.Join(temp, ", "))
		bytequery.WriteString(" WHERE user_id = :user_id")

		query = bytequery.String()
		err = nil
	}
	fmt.Println(columns)
	return
}
<<<<<<< HEAD

func GetMember(c *gin.Context, userID string) (m *Member, err error) {
	db := c.MustGet("DB").(*sqlx.DB)
	member := Member{}
	// err := db.QueryRowx("SELECT work, birthday, description, register_mode, social_id, c_editor, hide_profile, profile_push, post_push, comment_push, user_id, name, nick, create_time, updated_at, gender, mail, updated_by, password, active, profile_picture, identity FROM members where user_id = ?", userID).StructScan(&member)
	err = db.QueryRowx("SELECT * FROM members where user_id = ?", userID).StructScan(&member)

	switch {
	case err == sql.ErrNoRows:
		log.Printf("No user found.")
		m = nil
	case err != nil:
		log.Fatal(err)
		m = nil
	default:
		fmt.Printf("Successful get user:%s\n", userID)
		m = &member
	}
	return
}

func InsertMember(c *gin.Context, member *Member) (result *Member, err error) {
	// Need to manually parse null before insert
	// Do not insert create_time and updated_by,
	// left them to the mySQL default
	db := c.MustGet("DB").(*sqlx.DB)
	query, _ := makeSQL(member, "insert")
	_, e := db.NamedExec(query, member)
	// Cannot handle duplicate insert, crash
	if e != nil {
		log.Fatal(err)
		result = nil
		err = e
	}
	result = member
	err = nil
	return
}

func UpdateMember(c *gin.Context, member *Member) (*Member, error) {

<<<<<<< HEAD
	// Update time incase it's not in the request body
	if member.CreateTime.Valid {
		// member.CreateTime.Time = nil
		member.CreateTime.Valid = false
	}
	if !member.UpdatedAt.Valid {
		member.UpdatedAt.Time = time.Now()
		member.UpdatedAt.Valid = true
	}
=======
>>>>>>> seperate http handler from models
	db := c.MustGet("DB").(*sqlx.DB)
	query, _ := makeSQL(member, "partial_update")
	_, err := db.NamedExec(query, member)
	// fmt.Print(query)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func DeleteMember(c *gin.Context, userID string) (*Member, error) {

	db := c.MustGet("DB").(*sqlx.DB)
	_, err := db.Exec("UPDATE members SET active = 0 WHERE user_id = ?", userID)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &Member{ID: userID}, nil
}
=======
>>>>>>> Used global db instance in models package. Member struct implemented interface.
