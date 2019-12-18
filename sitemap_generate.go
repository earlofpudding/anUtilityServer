package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {

	CORE_URL := "https://animenetwork.net"
	resp, err := http.Get("https://animeapi.com/anime/")

	if err != nil {
		fmt.Println(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body) //read the bullshit

	cleanBody := string(body)

	b := []byte(string(cleanBody))

	var f interface{}
	err1 := json.Unmarshal(b, &f)

	if err1 != nil {
		fmt.Println(err1)
		return
	}

	m := f.(map[string]interface{})

	for k, v := range m {
		switch vv := v.(type) {
		// case string:
		// 	fmt.Println(k, "is string", vv)
		// case float64:
		// 	fmt.Println(k, "is float64", vv)
		case []interface{}:

			for _, u := range vv {

				switch oi := u.(type) {
				case map[string]interface{}:
					anime_id := oi["id"]
					fmt.Printf("%s%s%s%s\n", CORE_URL, "/anime/", anime_id, "/")

					//Fetch anime episodes
					anime_url := ("https://animeapi.com/anime/" + anime_id.(string) + "/episodes/")

					epReq, _ := http.Get(anime_url)
					epBody, _ := ioutil.ReadAll(epReq.Body)

					var sk interface{}

					fail := json.Unmarshal([]byte(epBody), &sk)
					if fail == nil {
						aEp := sk.(map[string]interface{})
						aEpData := aEp["data"].([]interface{})

						for _, v := range aEpData {
							// Now looping through each episode, aEpData[i] is each entity
							cleanV := v.(map[string]interface{})
							epId := cleanV["id"] // episode id of each component

							fmt.Printf("%s%s%s%s\n", CORE_URL, "/episode/", epId, "/")
						}
					}

				}

			}
		default:
			fmt.Println(k, "is of a type I don't know how to handle")
		}
	}

}
