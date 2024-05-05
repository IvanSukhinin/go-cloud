package main

import "cloud/internal/app/client"

func main() {
	c := client.New()
	c.Run()
}
