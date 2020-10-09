package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:3001")
	fmt.Println("starting example server...")
	if err != nil {
		log.Fatal(err)
	}
	mx := http.DefaultServeMux
	mx.HandleFunc("/getDataSource", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Write([]byte(`{ "data" : {"token" : "someSampleValue"}}`))
	})
	mx.HandleFunc("/login", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Write([]byte(`{ "data" : {"username" : "bob", "password" : "123456"}}`))
	})
	mx.HandleFunc("/getUser", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("user") == "bob" && request.Header.Get("X-Password") == "123456"{
			writer.WriteHeader(200)
			writer.Write([]byte(`{
	'username' : 'bob', 'firstName' : 'bob', "lastName" : "doe", "createdAt" : "2020"
}`))
		} else if request.URL.Query().Get("user") == "bob" && request.Header.Get("X-Password") == "123456" {
			writer.WriteHeader(200)
			writer.Write([]byte(`{
	'username' : 'bob', 'firstName' : 'bob', "lastName" : "doe", "createdAt" : "2020"
}`))
		} else {
			writer.WriteHeader(404)
			writer.Write([]byte(`user not found`))
		}

	})
	mx.HandleFunc("/verifyToken", func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-Sample-Token") == "token-someSampleValue" {
			writer.WriteHeader(200)
			writer.Header().Set("X-Token-Verify", "true")
		} else {
			writer.WriteHeader(404)
			writer.Write([]byte(`token not found`))
		}

	})
	if err := http.Serve(lis, mx); err != nil {
		log.Fatal(err)
	}
}
