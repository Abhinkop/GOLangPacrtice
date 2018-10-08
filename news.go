package main

// An App that currently filters out only Soccer news fron the "Washington Post"
// Adding somthing still in learning
import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

var wg sync.WaitGroup

type SiteMapIndex struct {
	Locations []string `xml:"sitemap>loc"`
}

type News struct {
	Titles    []string `xml:"url>news>title"`
	Keywords  []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
}

type NewsMap struct {
	Keywords string
	Location string
}
type SingleNews struct {
	Title   string
	Loc     string
	Keyword string
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	//p := SingleNews{"Title","Location"}
	football_news := make(map[string]NewsMap)
	x1 := []string{"SOCCER", "FOOTBALL"}
	GetNews(football_news, x1)
	//fmt.Println(football_news)
	t, err := template.ParseFiles("BasicTemplate.html")
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	//fmt.Println(p)
	err = t.Execute(w, football_news)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func GetNewsRoutine(locs string, c chan SingleNews, Keyword []string, trackingChannel chan string) {
	defer fmt.Println("closed", locs)
	//defer wg.Done()
	fmt.Println("launched", locs)
	var n News
	response, _ := http.Get(locs)
	bytes, _ := ioutil.ReadAll(response.Body)
	xml.Unmarshal(bytes, &n)
	response.Body.Close()
	//i := 0
	for index, _ := range n.Titles {
		//fmt.Println(n.Titles[index])
		upper_title := strings.ToUpper(n.Keywords[index])
		for _, key1 := range Keyword {
			if strings.Contains(upper_title, key1) {
				fmt.Println(n.Titles[index])
				//i = i + 1
				c <- SingleNews{
					n.Titles[index],
					n.Locations[index],
					n.Keywords[index]}
			}
		}
	}
	trackingChannel <- locs
}

func GetLocs(in_siteName string, c chan string) {
	//var Locations []string
	defer wg.Done()
	var url string
	if in_siteName == "Washington Post" {
		url = "https://www.washingtonpost.com/news-sitemap-index.xml"
	}
	if in_siteName == "Daily Mail UK" {
		url = "https://www.dailymail.co.uk/newssitemap.xml"
	}
	var s SiteMapIndex
	response, _ :=
		http.Get(url)
	bytes, _ := ioutil.ReadAll(response.Body)
	xml.Unmarshal(bytes, &s)
	response.Body.Close()
	for _, locs := range s.Locations {
		//Locations = append(Locations, locs)
		c <- locs
	}
	//return Locations
}

func GetNews(out_map map[string]NewsMap, Keyword []string) {
	fmt.Printf("Getting %s News Started\n", Keyword)

	//var s SiteMapIndex
	//news_map := make(map[string]NewsMap)
	xx := make(chan string, 30)
	wg.Add(1)
	//go GetLocs("Washington Post", xx)
	go GetLocs("Daily Mail UK", xx)
	wg.Wait()
	close(xx)
	trackingMap := make(map[string]bool)
	channelForTracking := make(chan string)
	q := make(chan SingleNews, 2)
	for locs := range xx {
		fmt.Println(locs)
		//wg.Add(1)
		trackingMap[locs] = true
		go GetNewsRoutine(locs, q, Keyword, channelForTracking)
	}
	//fmt.Println(ll)
	// response, _ :=
	// 	http.Get("https://www.washingtonpost.com/news-sitemap-index.xml")
	// bytes, _ := ioutil.ReadAll(response.Body)
	// xml.Unmarshal(bytes, &s)
	// for _, locs := range s.Locations {
	// 	wg.Add(1)
	// 	go GetNewsRoutine(locs, q, Keyword)
	// }
	var breakflag bool = false
	for {
		select {
		case news := <-q:
			{
				out_map[news.Title] = NewsMap{
					news.Keyword,
					news.Loc}
			}
		case loc := <-channelForTracking:
			{
				trackingMap[loc] = false
				breakflag = true
				for loc, val := range trackingMap {
					fmt.Println(loc, val)
					if val {
						breakflag = false
						break
					}
				}
			}
		}
		if breakflag == true {
			break
		}
	}
	// for n := range q {

	// 	out_map[n.Title] = NewsMap{
	// 		n.Keyword,
	// 		n.Loc}
	// }
	//fmt.Println(out_map)
	//response.Body.Close()
	fmt.Printf("Getting %s News Ended\n", Keyword)
}

func main() {
	http.HandleFunc("/a", index_handler)
	http.ListenAndServe(":8000", nil)

}
