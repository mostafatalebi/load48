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
	mx.HandleFunc("/login", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Write([]byte(`{
	'username' : 'bob', 'password' : '123456'
}`))
	})
	mx.HandleFunc("/getUser", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		writer.Write([]byte(`{
	'username' : 'bob', 'firstName' : 'bob', "lastName" : "doe", "createdAt" : "2020"
}`))
	})
	if err := http.Serve(lis, mx); err != nil {
		log.Fatal(err)
	}
}
