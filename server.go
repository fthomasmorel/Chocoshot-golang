package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Post struct {
	ID           bson.ObjectId `bson:"_id,omitempty"`
	User         string
	Position     string
	FileName     string
	IsHorizontal string
	Filter       string
	Separator    string
}

func insertDatabase(p Post) {
	session, _ := mgo.Dial("127.0.0.1")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB("chocoshot").C("post")
	c.Insert(&Post{User: p.User, Position: p.Position, FileName: p.FileName, IsHorizontal: p.IsHorizontal, Filter: p.Filter, Separator: p.Separator})
}

func getFromDataBaseWithUser(u string, deletation bool) Post {
	session, _ := mgo.Dial("127.0.0.1")
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	c := session.DB("chocoshot").C("post")
	var results Post
	c.Find(bson.M{"user": u}).One(&results)
	if deletation == true {
		c.Remove(bson.M{"filename": results.FileName})
	}
	return results
}

func getPost(w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get("user")
	deletation := r.Header.Get("deletation")

	fmt.Println("deletation = " + deletation)

	var res = getFromDataBaseWithUser(user, deletation == "true")
	json.NewEncoder(w).Encode(res)
}

func uploadPost(w http.ResponseWriter, r *http.Request) {
	user := strings.TrimSpace(r.FormValue("user"))
	position := strings.TrimSpace(r.FormValue("position"))
	isHorizontal := strings.TrimSpace(r.FormValue("isHorizontal"))
	filter := strings.TrimSpace(r.FormValue("filter"))
	separator := strings.TrimSpace(r.FormValue("separator"))

	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	fileName := RandomString(20)
	f, err := os.OpenFile("./img/"+fileName+".png", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	insertDatabase(Post{User: user, Position: position, FileName: fileName + ".png", IsHorizontal: isHorizontal, Filter: filter, Separator: separator})

	defer f.Close()
	io.Copy(f, file)
}

func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getPost(w, r)
		} else {
			uploadPost(w, r)
		}
	})

	mux.Handle("/image/", http.StripPrefix("/image/", http.FileServer(http.Dir("./img/"))))
	/*if strings.HasPrefix(r.URL.RequestURI(), "/image/") {
		file := strings.Replace(r.URL.RequestURI(), "/image/", "./img/", 1)
		fmt.Println("deleting " + file)
		defer os.Remove(file)
	}*/

	fmt.Println("listening at :9000")
	http.ListenAndServe(":9000", mux)
}
