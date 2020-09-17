package demo2Rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"go_systems/src/demo2Config"
	"go_systems/src/demo2Jwt"
	"go_systems/src/demo2Mongo"
	"go_systems/src/demo2MySql"
	"go_systems/src/demo2Users"
	"go_systems/src/demo2Utils"
	"go_systems/src/demo2fs"
	"go_systems/src/websockets"
)

func addHeaders(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Content-Type", "application/json")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func sendRestMsg(w *http.ResponseWriter, jwt string, msgType string, data string) error {
	m := websockets.RestMessage{
		Jwt:  jwt,
		Type: msgType,
		Data: data,
	}
	encoded, err := json.Marshal(m)
	if err != nil {
		return err
	}
	_, err = (*w).Write(encoded)
	return err
}

func isAuthBearerValid(w http.ResponseWriter, r *http.Request, checkFor string) (bool, error) {
	_, ok := r.Header["Authorization"]
	if !ok {
		return ok, errors.New("no Authorization Header")
	}
	header := r.Header.Get("Authorization")
	bearer := strings.Split(header, " ")
	valid, err := demo2Jwt.ValidateJwt(demo2Config.PubKeyFile, bearer[1])

	switch checkFor {
	case "filepond":
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("500 - Sorry, something went wrong..."))
			break
		}
	case "rest-test":
		if err != nil {
			fmt.Println(sendRestMsg(&w, "^vAr^", "rest-jwt-token-invalid", err.Error()))
			break
		} else if valid {
			fmt.Println(sendRestMsg(&w, "^vAr^", "rest-jwt-token-valid", "/rest/jwt/test"))
			break
		}
	case "noop":
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println(sendRestMsg(&w, "vAr^", "rest-jwt-token-invalid", "noop"))
		} else if valid {
			// do nothing
		}
	default:
		break
	}
	return valid, err
}

/** JWT Token Protection Tests **/
func HandleProtectedGetRequestTest(w http.ResponseWriter, r *http.Request) {
	addHeaders(&w)
	if (*r).Method == "OPTIONS" {
		return
	}
	ok, err := isAuthBearerValid(w, r, "rest-test")
	if !ok {
		return
	}
	fmt.Println("GOLDEN!!!")
	fmt.Println(ok, err)

}

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handling Upload")
	addHeaders(&w)
	if (*r).Method == "OPTIONS" {
		return
	}
	valid, err := isAuthBearerValid(w, r, "filepond")
	if !valid || err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Getting ready to parse form")
	if (*r).Method == "POST" {
		err := r.ParseMultipartForm(0)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("MultipartForm %v", r.MultipartForm.File)
		for key, files := range r.MultipartForm.File {
			fmt.Println(key)
			fmt.Println(files)
			for _, file := range files {
				name := strings.Split(file.Filename, ".")
				ext := name[len(name)-1]
				fmt.Println(name, ext)
				fmt.Println(demo2fs.CreateFile(demo2Config.FileStoragePath, file.Filename))
				f, err := file.Open()
				if f == nil {
					continue
				}
				if err != nil {
					fmt.Print("Error opening multipart file...")
				} else {
					buf := bytes.NewBuffer(nil)
					if _, err := io.Copy(buf, f); err != nil {
						fmt.Println(err)
					} else {
						fmt.Println(demo2fs.WriteFile(demo2Config.FileStoragePath+file.Filename, buf.Bytes()))
					}
				}
				fmt.Println(f.Close())
			}
		}
	}
}

func HandleCreateUserMongo(w http.ResponseWriter, r *http.Request) {
	addHeaders(&w)
	if (*r).Method == "OPTIONS" {
		return
	}
	if _, err := isAuthBearerValid(w, r, "noop"); err != nil {
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading request body")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("500 - Sorry, something went wrong..."))
		return
	}
	fmt.Println("Request ", string(body))
	user := demo2Users.AuthUser{}
	if err := json.Unmarshal([]byte(body), &user); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("500 - Sorry, something went wrong..."))
	}
	fmt.Print("Encoded user: ", user)
	pwHash, err := demo2Utils.GenerateUserPassword(user.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("500 - Sorry, something went wrong..."))
	}
	user.Password = pwHash
	fmt.Print("Checking for dock with Email then Alias...")
	userExists := demo2Mongo.CheckDocumentExists("api", "users", "email", user.Email)
	aliasExists := demo2Mongo.CheckDocumentExists("api", "users", "alias", user.Alias)
	if userExists || aliasExists {
		fmt.Println("user exists")
		msgAry := "["
		if userExists {
			msgAry += "'user email in system'"
		}
		if aliasExists {
			msgAry += "'user alias exists in system'"
		}
		msgAry += "]"
		fmt.Println("Error: ", sendRestMsg(&w, "^vAr^", "rest-create-user-fail", msgAry))
	}
	mDoc, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("500 - Sorry, something went wrong..."))
		return
	}
	docId, err := demo2Mongo.InsertDocument("api", "users", mDoc)
	if err != nil {
		fmt.Println("Error inserting user", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("500 - Sorry, something went wrong..."))
		return
	}
	data, err := json.Marshal(struct {
		docId string
	}{docId})
	fmt.Println(sendRestMsg(&w, "^vAr^", "rest-create-user-success", string(data)))
}

func HandleCreateUserMySql(w http.ResponseWriter, r *http.Request) {
	addHeaders(&w)
	if (*r).Method == "OPTIONS" {
		return
	}
	if ok, err := isAuthBearerValid(w, r, "noop"); err != nil || !ok {
		fmt.Println(sendRestMsg(&w, "^vAr^", "rest-test-create-mysql-user-failure", "Auth token error"))
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(sendRestMsg(&w, "^vAr^", "rest-test-create-mysql-user-failure", "parse-body error"))
		return
	}
	user := demo2MySql.MySqlStoredUser{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		fmt.Println(sendRestMsg(&w, "^vAr^", "rest-test-create-mysql-user-failure", "unmarshal error"))
		return
	}
	uuid, err := demo2Utils.GenerateUUID()
	if err != nil {
		fmt.Println(sendRestMsg(&w, "^vAr^", "rest-test-create-mysql-user-failure", "uuid-error"))
		return
	}
	pass := demo2Utils.SHA256OfString(user.Password)
	user.UUID = uuid
	user.Password = pass
	err = demo2MySql.CreateUser("demo2", &user)
	if err != nil {
		fmt.Println(sendRestMsg(&w, "^vAr^", "rest-test-create-mysql-user-failure", "my-sql-error"))
		return
	}
	fmt.Println(sendRestMsg(&w, "^vAr^", "rest-test-create-mysql-user-success", "TAHDAH!"))
}
