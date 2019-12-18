package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
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

	//Open sitemap file for writing to
	f, _ := os.Create("sitemap.xml")
	defer f.Close()

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

		f.Write([]byte(`
	<url><loc>` + animeURL + "</loc></url>"))

		fmt.Println(animeURL)

		// Loop through each episode of the anime
		apiEpURL := "https://animeapi.com/anime/" + animeID.(string) + "/episodes"
		var raw map[string]interface{}
		err := fetchJSON(apiEpURL, &raw)
		if err != nil {
			fmt.Println(err)
			return
		}

		if (raw["status"].(string) == "FOUND") || (raw["status"].(string) == "found") {
			data := raw["data"].([]interface{}) //slice the random interface
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
					<video:live>no</video:live>
				</video:video></url>`))

				fmt.Println(epURL)

			}
		}
	}

	//Close Sitemap
	f.Write([]byte(`
</urlset>`))

	f.Sync()

}
