package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func fetchJSON(url string, resp interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(&resp)
}

func main() {
	siteURL := "https://animenetwork.net"

	var raw map[string]interface{} //Set empty interface to handle the "unkown" data types
	err := fetchJSON("https://animeapi.com/anime", &raw)
	if err != nil {
		fmt.Println(err)
		return
	}

	data := raw["data"].([]interface{})

	for _, v := range data {
		animeID := v.(map[string]interface{})["id"]

		animeURL := siteURL + "/anime/" + animeID.(string) + "/"
		fmt.Println(animeURL)

		// Loop through each episode of the anime

	}

}
