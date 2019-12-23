package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v7"
)

/*
	CACHE HASHMAP SHIT

	anime HKey name: `anime`
	anime HKey values: `anime anime:660`
	anime episodes HKey values: `anime anime:660:episodes:`
*/

func fetchJSON(url string, resp interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}

	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(&resp)
}

func cache(hName string, hKey string, data *interface{}, rClient *redis.Client) error {
	encodedJSON, _ := json.Marshal(&data)
	_, e := rClient.HSet(hName, hKey, encodedJSON).Result()
	return e
}

func pullCache(rClient *redis.Client, hName string, hKey string, output *interface{}) error {
	val, _ := rClient.HGet(hName, hKey).Result()
	return json.Unmarshal([]byte(val), &output)
}

func main() {
	urlCount := 1
	fileCount := 1
	siteURL := "https://animenetwork.net"

	//Open sitemap file for writing to
	f, _ := os.Create("sitemap-" + strconv.Itoa(fileCount) + ".xml")
	// defer f.Close()

	//Init redis
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer client.Close()

	//Ping server, wait for pong
	_, cErr := client.Ping().Result()
	if cErr != nil {
		fmt.Println(cErr)
		return
	}

	//Opening sitemap data

	f.Write([]byte(`
	<?xml version="1.0" encoding="UTF-8"?>
	<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"> `))

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
		animeGenres := v.(map[string]interface{})["genres"]

		if urlCount < 50000 {
			f.Write([]byte(`
			<url><loc>` + animeURL + "</loc></url>"))
			urlCount++
		} else {
			urlCount = 1
			fileCount++
			f.Close()
			f, _ = os.Create("sitemap-" + strconv.Itoa(fileCount) + ".xml")

			f.Write([]byte(`
			<url><loc>` + animeURL + "</loc></url>"))
			urlCount++
		}

		fmt.Println(animeURL)

		// Loop through each episode of the anime, but check cache first
		var epRaw interface{}
		cachedEps, _ := client.HExists("anime",
			"anime:"+animeID.(string)+":episodes:").Result()

		if cachedEps {

			pullCache(client, "anime",
				"anime:"+animeID.(string)+":episodes:",
				&epRaw)

		} else {
			//Episodes not cached, send API request and cache it into the shit
			apiEpURL := "https://animeapi.com/anime/" + animeID.(string) + "/episodes"
			err := fetchJSON(apiEpURL, &epRaw)
			if err != nil {
				fmt.Println(err)
				return
			}

			epRaw = epRaw.(map[string]interface{})["data"]
			cache("anime", "anime:"+animeID.(string)+":episodes:", &epRaw, client)

		}

		if epRaw != nil {
			//if anime actuall has episodes
			data := epRaw.([]interface{})
			for _, v := range data {
				//Within each episode
				epID := v.(map[string]interface{})["id"]
				epURL := siteURL + "/episode/" + epID.(string) + "/"
				epP := v.(map[string]interface{})["image"]
				epPicture := "https:" + strings.Replace(epP.(string), "animeapi.com", "animenetwork.net", 1)
				epDateFull := v.(map[string]interface{})["date"]
				epDate := epDateFull.(string)[0:10]

				epTitle := v.(map[string]interface{})["title"]
				if epTitle == "" {
					epTitle1 := v.(map[string]interface{})["name"]
					epTitle = epTitle1.(map[string]interface{})["default"]
				}

				epDesc := v.(map[string]interface{})["description"]
				if epDesc == "" || epDesc == nil {
					epDesc = "Watch " + epTitle.(string) + " on animenetwork.net, the best source for watching anime for free! We offer free streaming of over 100,000 anime and cartoons and are always expanding out collection"
				}

				dubbedanimeURL := strings.Replace(epURL, "animenetwork.net/episode/", "watchdubbed.net/anime/watch/", 1)

				if animeGenres == nil {
					animeGenres = ""
				}

				if urlCount < 50000 {
					f.Write([]byte(`
					<url><loc>` + epURL + `</loc> 
					<video:video>
						<video:thumbnail_loc>` + epPicture + `</video:thumbnail_loc>
						<video:title>` + epTitle.(string) + `</video:title>
						<video:description>` + epDesc.(string) + `</video:description>
						<video:platform relationship="allow">web tv</video:restriction>
						<video:requires_subscription>no</video:requires_subscription>
						<video:category>` + animeGenres.(string) + `</video:category>
						<video:publication_date>` + epDate + `</video:publication_date>
						<video:player_loc>` + dubbedanimeURL + `</video:player_loc>
						<video:live>no</video:live>
					</video:video></url>`))
					urlCount++
				} else {
					urlCount = 1
					fileCount++
					f.Close()
					f, _ = os.Create("sitemap-" + strconv.Itoa(fileCount) + ".xml")

					f.Write([]byte(`
					<url><loc>` + epURL + `</loc> 
					<video:video>
						<video:thumbnail_loc>` + epPicture + `</video:thumbnail_loc>
						<video:title>` + epTitle.(string) + `</video:title>
						<video:description>` + epDesc.(string) + `</video:description>
						<video:platform relationship="allow">web tv</video:restriction>
						<video:requires_subscription>no</video:requires_subscription>
						<video:category>` + animeGenres.(string) + `</video:category>
						<video:publication_date>` + epDate + `</video:publication_date>
						<video:player_loc>` + dubbedanimeURL + `</video:player_loc>
						<video:live>no</video:live>
					</video:video></url>`))
					urlCount++
				}

				fmt.Println(epURL)
				fmt.Println(urlCount)

			}
		}

	}

	//Close Sitemap
	f.Write([]byte(`
</urlset>`))

	f.Sync()
	f.Close()
}
