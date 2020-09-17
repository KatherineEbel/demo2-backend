package demo2MySql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"

	"go_systems/src/demo2Config"
	"go_systems/src/websockets"
)

var (
	DBConn *sql.DB
)

func init() {
	var err error
	DBConn, err = sql.Open("mysql", demo2Config.MySqlUser+":"+demo2Config.MySqlPass+"@tcp(127.0.0.1:3306)/")
	if err != nil {
		log.Fatalln("Unable to connect to MySql", err)
	}
	err = DBConn.Ping()
	if err != nil {
		log.Fatalf("Can't Ping MySql")
	} else {
		fmt.Println("MySql connected...")
	}
	DBConn.SetMaxOpenConns(20)
}

// Rest API
func GetDBs() ([]byte, error) {
	var names []string
	rows, err := DBConn.Query("SHOW DATABASES;")
	if err != nil {
		return nil, err
	}
	var dbs string
	for rows.Next() {
		_ = rows.Scan(&dbs)
		names = append(names, dbs)
		fmt.Println("DBs", names)
	}
	encoded, err := json.Marshal(names)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func GetTables(database string) ([]byte, error) {
	var tblNames []string
	tx, err := DBConn.Begin()
	if err != nil {
		return nil, err
	}
	if _, err := tx.Query("USE " + database); err != nil {
		return nil, err
	}
	rows, err := tx.Query("SHOW TABLES;")
	if err != nil {
		return nil, err
	}
	var dbTn string
	for rows.Next() {
		_ = rows.Scan(&dbTn)
		tblNames = append(tblNames, dbTn)
	}
	encoded, err := json.Marshal(tblNames)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	} // close transaction after
	return encoded, nil
}

func Select(database string, query string) (interface{}, error) {
	tx, err := DBConn.Begin()
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)

	result := map[int]map[string]string{}
	id := 0
	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		tmp := map[string]string{}
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			tmp[col] = fmt.Sprintf("%s", v)
		}
		result[id] = tmp
		id++
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	fmt.Println(result)
	return result, nil
}

func Exec(database string, query string) error {
	tx, err := DBConn.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Query("USE " + database); err != nil {
		return err
	}
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func AddTodo(td string) {
	// TODO: For Testing
}

type MySqlStoredUser struct {
	UUID              string `json:"uuid"`
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	Alias             string `json:"alias"`
	Role              string `json:"role"`
	Email             string `json:"email"`
	Password          string `json:"password"`
	TimestampCreated  int64  `json:"timestampCreated"`
	TimestampModified int64  `json:"timestampModified"`
}

// used with create user
func CreateUser(database string, u *MySqlStoredUser) error {
	tx, err := DBConn.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Query("USE " + database); err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		"INSERT INTO USERS(uuid, first_name, last_name, alias, role, email, password) VALUES (?,?,?,?,?,?,?)")
	if stmt != nil {
		defer stmt.Close()
	}
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(u.UUID, u.FirstName, u.LastName, u.Alias, u.Role, u.Email, u.Password); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// For WebSocket Tasks :: as example
type GetMysqlDbsTask struct {
	ws *websocket.Conn
}

func NewGetMysqlDbsTask(ws *websocket.Conn) *GetMysqlDbsTask {
	return &GetMysqlDbsTask{ws}
}

func (t *GetMysqlDbsTask) Perform() {
	dbs, err := GetDBs()
	if err != nil {
		fmt.Println("GetDB Task error", err)
		return
	}
	m := &websockets.Message{
		Jwt:  "^vAr^",
		Type: "mysql=dbs-list",
		Data: string(dbs),
	}
	fmt.Println(m.Send(t.ws))
}
