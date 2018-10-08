package main

// An App that currently filters out only Soccer,the real "Football" news fron the "Washington Post","DailyMail UK"
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
	football_news := make(map[string]NewsMap) // A map to holf the output refactor this to use SingleNews Slice
	footbalKeywords := []string{"SOCCER", "FOOTBALL"}
	GetNews(football_news, footbalKeywords)
	template, err := template.ParseFiles("BasicTemplate.html") // template to display the news. Beautify the HTML
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	err = template.Execute(w, football_news)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func GetNewsRoutine(locs string, c chan SingleNews, Keyword []string, trackingChannel chan string) {
	defer fmt.Println("closed", locs)
	fmt.Println("launched", locs)
	var n News // used to extract all the news
	response, _ := http.Get(locs)
	bytes, _ := ioutil.ReadAll(response.Body)
	xml.Unmarshal(bytes, &n)
	response.Body.Close()
	for index, _ := range n.Titles {
		upper_Keyword := strings.ToUpper(n.Keywords[index])
		for _, key := range Keyword {
			if strings.Contains(upper_Keyword, key) {
				// Send the news across the channel
				c <- SingleNews{
					n.Titles[index],
					n.Locations[index],
					n.Keywords[index]}
			}
		}
	}
	// Inform The use of the channel is done from the sender and hence the reciever can close it.
	trackingChannel <- locs
}

func GetLocs(in_sitemapURL string, c chan string, trackingChannel chan string) {
	var s SiteMapIndex // Used to extract the locations of the news
	response, _ :=
		http.Get(in_sitemapURL)
	bytes, _ := ioutil.ReadAll(response.Body)
	xml.Unmarshal(bytes, &s)
	response.Body.Close()
	for _, locs := range s.Locations {
		//Send the Locations across the channel
		c <- locs
	}
	// Inform The use of the channel is done from the sender and hence the reciever can close it.
	trackingChannel <- in_sitemapURL
}

func GetNews(out_map map[string]NewsMap, Keyword []string) {
	fmt.Printf("Getting %s News Started\n", Keyword)
	// Used to store the root URL of the sitemaps
	SiteMapRootURL := []string{
		"https://www.washingtonpost.com/news-sitemap-index.xml",
		"https://www.dailymail.co.uk/newssitemap.xml"}
	locsTrackingMap := make(map[string]bool)       // used to track if the go routine has completed exection.
	channelForTrackingGetLocs := make(chan string) // channel used by the goroutine to
	//communicate back that i has completed execution.
	LocsChannel := make(chan string, 30) // Channel to send Location of the news
	var LocationSlice []string
	for _, rootPath := range SiteMapRootURL {
		locsTrackingMap[rootPath] = true
		go GetLocs(rootPath, LocsChannel, channelForTrackingGetLocs)
	}
	var breakflag bool = false
	for {
		select {
		case locs := <-LocsChannel:
			{
				LocationSlice = append(LocationSlice, locs)
			}
		case rootPath := <-channelForTrackingGetLocs:
			{
				locsTrackingMap[rootPath] = false
				breakflag = true
				// check if all the go routines are completed then set flag to break out of the infinte for loop
				for _, val := range locsTrackingMap {
					if val {
						breakflag = false
						break
					}
				}
			}
		}
		if breakflag == true {
			// break out of the infinte for loop
			break
		}
	}
	close(LocsChannel)
	close(channelForTrackingGetLocs)
	trackingMap := make(map[string]bool)    // used to track if the go routine has completed exection.
	channelForTracking := make(chan string) // channel used by the goroutine to
	//communicate back that i has completed execution.
	NewsChannel := make(chan SingleNews, 2) // Channel to send the news
	for _, locs := range LocationSlice {
		trackingMap[locs] = true
		go GetNewsRoutine(locs, NewsChannel, Keyword, channelForTracking)
	}
	breakflag = false
	for {
		select {
		case news := <-NewsChannel:
			{
				out_map[news.Title] = NewsMap{
					news.Keyword,
					news.Loc}
			}
		case loc := <-channelForTracking:
			{
				trackingMap[loc] = false
				breakflag = true
				// check if all the go routines are completed then set flag to break out of the infinte for loop
				for _, val := range trackingMap {
					if val {
						breakflag = false
						break
					}
				}
			}
		}
		if breakflag == true {
			// break out of the infinte for loop
			break
		}
	}
	close(NewsChannel)
	close(channelForTracking)
	fmt.Printf("Getting %s News Ended\n", Keyword)
}

func main() {
	http.HandleFunc("/News", index_handler)
	http.ListenAndServe(":8000", nil)
}
