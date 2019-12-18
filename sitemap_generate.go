package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"encoding/json"
	"net/http"
)

func main() {

	BASE_URL := "https://animenetwork.net"

	resp, err := http.Get("https://animeapi.com/anime/")

	fmt.Printf(resp)

	// Pull all anime 

	// Pull episodes for each anime

	// Pull pages for main site

}