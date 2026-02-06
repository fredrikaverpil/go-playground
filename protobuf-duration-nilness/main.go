package main

import "log"

func main() {
	if err := listenAndServe("localhost:8080"); err != nil {
		log.Fatal(err)
	}
}
