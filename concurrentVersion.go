package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
)

/*
	Cache Hashmap layout
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

func fetchEpisodes(animeListChannel chan interface{}, client *redis.Client, f *os.File, siteURL string, urlCount *int, fileCount *int, animeGenres interface{}) {
	for {
		// Loop through each episode of the anime, but check cache first
		animeID := <-animeListChannel
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
				epSlug := v.(map[string]interface{})["slug"]
				epURL := siteURL + "/episode/" + epID.(string) + "-" + epSlug.(string) + "/"
				// epP := v.(map[string]interface{})["image"]
				// epPicture := "https:" + strings.Replace(epP.(string), "animeapi.com", "animenetwork.net", 1)
				// epDateFull := v.(map[string]interface{})["date"]
				// epDate := epDateFull.(string)[0:10]

				epTitle := v.(map[string]interface{})["title"]
				if epTitle == "" {
					epTitle1 := v.(map[string]interface{})["name"]
					epTitle = epTitle1.(map[string]interface{})["default"]
				}

				epDesc := v.(map[string]interface{})["description"]
				if epDesc == "" || epDesc == nil {
					epDesc = "Watch " + epTitle.(string) + " on animenetwork.net, the best source for watching anime for free! We offer free streaming of over 100,000 anime and cartoons and are always expanding out collection"
				}

				// dubbedanimeURL := strings.Replace(epURL, "animenetwork.net/episode/", "watchdubbed.net/anime/watch/", 1)

				if *urlCount < 50000 {
					f.Write([]byte(`
					<url><loc>` + epURL + `</loc></url>`))
					*urlCount++
				} else {
					*urlCount = 1
					*fileCount++
					f.Sync()
					f.Close()
					f, _ = os.Create("sitemap-" + strconv.Itoa(*fileCount) + ".xml")

					f.Write([]byte(`
					<url><loc>` + epURL + `</loc></url>`))
					*urlCount++
				}

				//fmt.Println(epURL)
			}
		}
	}
}

func main() {
	urlCount := 1
	fileCount := 1
	siteURL := "https://animenetwork.net"

	tStart := time.Now()

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

	//Flush current cache
	_, flushed := client.Del("anime").Result()

	if flushed != nil {
		fmt.Println("error database couldn't be flushed.")
	}

	//Opening sitemap data

	f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
	<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
		<url><loc>https://animenetwork.net/</loc></url>
		<url><loc>https://animenetwork.net/browse/</loc></url>
	`))

	//Inital Fetch of Anime List
	var raw map[string]interface{} //Set empty interface to handle the "unkown" data types
	err := fetchJSON("https://animeapi.com/anime", &raw)
	if err != nil {
		fmt.Println(err)
		return
	}

	data := raw["data"].([]interface{})
	animeListChannel := make(chan interface{})
	//Open up go routine to begin listening to animeListChannel and begin outputting shit right away if it recieves any data

	go fetchEpisodes(animeListChannel, client, f, siteURL, &urlCount, &fileCount, "")

	for _, v := range data {
		animeID := v.(map[string]interface{})["id"]
		animeSlug := v.(map[string]interface{})["slug"]
		animeURL := siteURL + "/anime/" + animeID.(string) + "-" + animeSlug.(string) + "/"
		animeGenres := v.(map[string]interface{})["genres"]

		//Check of cache if animeID exists
		animeCache, _ := client.HExists(
			"anime",
			"anime:"+animeID.(string)).Result()

		if !animeCache {
			cache("anime",
				"anime:"+animeID.(string),
				&v,
				client)
		}

		if animeGenres == nil {
			animeGenres = ""
		}

		if urlCount < 25000 {
			f.Write([]byte(`
				<url><loc>` + animeURL + "</loc></url>"))
			urlCount++
		} else {

			urlCount = 1
			fileCount++
			f.Sync()
			f.Write([]byte(`
			</urlset>`))
			f.Close()
			f, _ = os.Create("sitemap-" + strconv.Itoa(fileCount) + ".xml")

			f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
			<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
				<url><loc>` + animeURL + "</loc></url>"))
			urlCount++
		}

		// fmt.Println(animeURL)
		animeListChannel <- animeID
	}

	//Close Sitemap
	f.Write([]byte(`
</urlset>`))
	f.Sync()
	f.Close()

	//Make sitemap index
	f, _ = os.Create("sitemap.xml")

	f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
	<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	`))

	for i := 0; i < fileCount; i++ {
		f.Write([]byte(`
		<sitemap>
			<loc>` + siteURL + `/static/sitemap/sitemap-` + strconv.Itoa(i+1) + `.xml</loc>
			<lastmod>` + tStart.Format("2006-01-02") + `</lastmod>
		</sitemap>
		`))
	}

	f.Write([]byte(`
	</sitemapindex>
	`))
	f.Sync()
	f.Close()

	elapsed := time.Since(tStart)
	log.Printf("Binomial took %s", elapsed)
}
