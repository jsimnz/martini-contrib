package main

import (
	"net/http"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/jsimnz/martini-contrib/response"
	"github.com/martini-contrib/encoder"
)

type User struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Friends   []User `json:"friends"`
}

var myDb map[int]User
var friendsDb map[int][]User

func init() {
	myDb = make(map[int]User)
	friendsDb = make(map[int][]User)

	myDb[1] = User{FirstName: "Joe", LastName: "Smith"}
	myDb[2] = User{FirstName: "Heather", LastName: "Erica"}

	friendsDb[1] = []User{
		User{FirstName: "Heather", LastName: "Erica"},
	}
}

func main() {
	m := martini.Classic()

	m.Use(func(c martini.Context, w http.ResponseWriter) {
		c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	})

	m.Use(response.NewEncoder())

	m.Get("/users/:id", FindUserHandler)
	m.Get("/users/:id/friends", FindUserFriendsHandler)

	m.Run()
}

func FindUserHandler(params martini.Params) (int, interface{}) {
	idstr, exists := params["id"]
	if !exists {
		return 400, nil
	}

	id, err := strconv.Atoi(idstr)
	if err != nil {
		return 400, nil
	}

	user, exists := myDb[id]
	if !exists {
		return 404, nil
	}
	code, friends := FindUserFriendsHandler(params)
	if code != 200 {
		return 500, nil
	}
	user.Friends = friends.([]User)

	return 200, user
}

func FindUserFriendsHandler(params martini.Params) (int, interface{}) {
	idstr, exists := params["id"]
	if !exists {
		return 400, nil
	}

	id, err := strconv.Atoi(idstr)
	if err != nil {
		return 400, nil
	}
	friends := friendsDb[id]

	return 200, friends

}
