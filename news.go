package main

// An App that currently filters out only Soccer news fron the "Washington Post"
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
	GetNews(football_news, "SOCCER")
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

func GetNewsRoutine(locs string, c chan SingleNews, Keyword string) {
	defer fmt.Println("closed")
	defer wg.Done()
	fmt.Println("launched")
	var n News
	response, _ := http.Get(locs)
	bytes, _ := ioutil.ReadAll(response.Body)
	xml.Unmarshal(bytes, &n)
	response.Body.Close()
	for index, _ := range n.Titles {
		fmt.Println(n.Titles[index])
		upper_title := strings.ToUpper(n.Keywords[index])
		if strings.Contains(upper_title, Keyword) {
			c <- SingleNews{
				n.Titles[index],
				n.Locations[index],
				n.Keywords[index]}
		}
	}

}

func GetNews(out_map map[string]NewsMap, Keyword string) {
	fmt.Printf("Getting %s News Started\n", Keyword)

	var s SiteMapIndex
	//news_map := make(map[string]NewsMap)
	response, _ :=
		http.Get("https://www.washingtonpost.com/news-sitemap-index.xml")
	bytes, _ := ioutil.ReadAll(response.Body)
	xml.Unmarshal(bytes, &s)
	q := make(chan SingleNews, 30)
	for _, locs := range s.Locations {
		wg.Add(1)
		go GetNewsRoutine(locs, q, Keyword)
	}
	wg.Wait()
	close(q)
	for n := range q {

		out_map[n.Title] = NewsMap{
			n.Keyword,
			n.Loc}
	}
	fmt.Println(out_map)
	response.Body.Close()
	fmt.Printf("Getting %s News Ended\n", Keyword)
}

func main() {
	http.HandleFunc("/a", index_handler)
	http.ListenAndServe(":8000", nil)

}
