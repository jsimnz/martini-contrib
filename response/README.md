# Response Encoder

A [martini](http://github.com/go-martini/martini) Middleware to encode returned values from handlers (not encode values in the handler itself)

### Why not use martini-contrib/encoder or render?

Both [martini-contrib](http://github.com/martini-contrib)/[encoder](http://github.com/martini-contrib/encoder) and [render](http://github.com/martini-contrib/render) require you to encoder your data **in**  your handler. This approach is great, however if you needed to call a handler from another handler ie FindUser handler needs to call FindUserFollowers, it can be ahallenging because the response is already written to the http.ResponseWriter buffer which is an unexported field.

A possible solution (using the above example) is to just call FindUserFriends from your FindUser handler but the returned value is already encoded so you must deocde before getting access to its contents. This adds A) overhead and B) more code

An ideal solution would be to be able to call FindUserFriends and recieve un encoded data. This is what this response encoder does.

You can return your struct or []struct like you normally would in a function, and it is encoded after all your handlers are finished running.

#### Features
- Encodes final returned values from handlers
- Write handlers without encoding values
- Chain handlers together without needing to be decoded
- Works with popular [martini-contrib/encoder](http://github.com/martini-contrib/encoder)
- Automatic error marshaling based on response code

### Usage

It currently requires you to be using the [encoder](http://github.com/martini-contrib/encoder) middleware which is easy to get running, And works with the generic martini handler structure.

This is a over simplified example, but understand how the return values from handlers work

```
package main

import (
	"net/http"

	"github.com/jsimnz/martini-contrib/response"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/encoder"
)

type User struct {
	Firstname string
	Lastname string
	Friends []User
}

// this isnt important
var myDb SomeDB

func main() {
	m := martini.Classic()

	// Insert encoder middleware
	m.Use(func(c martini.Context, w http.ResponseWriter) {
        c.MapTo(encoder.JsonEncoder{}, (*encoder.Encoder)(nil))
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
    })

    // Register the response.NewEncoder as the main ReturnHandle martini should use
    m.Use(response.NewEncoder())

    m.Get("/user/:name", userHandler)

    m.Get("/user/:name/friends", friendsHandler)
}

func userHandler(params martini.Params) User {
	name := params["name"]
	user := myDb.Find(name)

	// we can call another handler and use its data immediately
	friends := friendsHandler(params)
	user.Friends = friends

	// Return un-encoded data, encoded later 
	return user
}

func friendsHandler(params martini.Params) []User {
	name := params["name"]
	friends := myDb.FindFriends(name)

	// return un-encoded data
	return friends
}
```

As you can see in this example, the `friendsHandler` can be either called directly by martini as the handler for the HTTP request /user/:name/friends, or can be called by the `userHandler`, and the logic is all the same.

### Return Values

The return values can be **any struct or slice of struct ([]struct)**, and it will be encoded using whatever encoder you register from [martini-contrib/encoder](http://github.com/martini-contrib/encoder), meaning it works with any `json` tag or `out` tag you define on your struct.

It also works with martini handlers that return error codes as ints, causing the error code to be set in the ResponseWriter. 

```
func userHandler(params martini.Params) User {
	name := params["name"]
	user := myDb.Find(name)

	// we can call another handler and use its data immediately
	err, friends := friendsHandler(params)
	if err != 200 {
		return err, nil
	}
	user.Friends = friends

	// Return un-encoded data, encoded later 
	return user
}

func friendsHandler(params martini.Params) (int, []User) {
	
	... same as above ...
	if err != nil {
		return 400, nil // bad request
	}

	return 200, friends
}
```

### Error Values

If your handler encounters some error during execution, you can return `nil` and optionally some error code, and it will automatically marshled into an error struct. If no error code is returned just `nil`, then it will pull the error code from another handler that set the response header code via `http.ResponseWriter.WriteHeader(400)`, and marshal the error struct as follows

**If no error code is returned via handlers OR set by the ResponseWriter a panic will occur**

JSON Ex.
```
{
	"error": 400				// Whatever error code is returned or grabbed the ResponseWriter
	"message": "Bad Request"	// The associated status text from http.StatusText(code)
}
```

However a more idomatic method would be to, instead of return `nil`, return your own error struct along with an error code. This gives you full control of what is returned when an error happens.

Ex.
```
type HTTPError struct {
	Code int
	Message string
}

func userHandler(params martini.Params) User {
	name := params["name"]
	user := myDb.Find(name)

	// we can call another handler and use its data immediately
	err, friends := friendsHandler(params)
	if err != 200 {
		return err, nil
	}
	user.Friends = friends.([]User) // cast to appropriate type

	// Return un-encoded data, encoded later 
	return user
}

func friendsHandler(params martini.Params) (int, interface{}) {
	
	... same as original definition ...
	if err != nil {
		return 400, ErrorJSON{400, "Some error message"}
	}

	return 200, friends
}
```

### Notes

If you use `response.NewEncoder()` as a middleware to encode returned values from handlers, then you need to keep note the following things.

	1. It replaces the default `martini.ReturnHandler`, so this package won't work if another middleware also replaces the default `martini.ReturnHandler`

	2. It wraps the http.ResponseWriter, so if another middleware also wraps the http.ResponseWriter be careful.

### TODO
- Write tests

### LICENSE 
```
The MIT License (MIT)

Copyright (c) 2014 John-Alan Simmons

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```