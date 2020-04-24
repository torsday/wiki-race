package puregorace

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type wikiRaceResponse struct {
	Completed      bool   `json:"completed"`
	Destination    string `json:"destination"`
	ElapsedTimeSec string `json:"elapsedTimeSec"`
	Message        string `json:"message"`
	Path           string `json:"path"`
	Start          string `json:"start"`
}

type node struct {
	parent *node
	pathId int
	url    string
}

const (
	PathIdFromTheFront = 1
	PathIdFromTheRear  = 2
)

/**
containsExcludedPrefix checks a url against slice of prefixes that imply link is not an article.
*/
func containsExcludedPrefix(href string) bool {

	// prefixesToExclude to ignore in order so as to limit children to actual articles
	// I'm being liberal with this as prefixed articles like ISSN and ISBN are far too common, and I want the race to
	// be more fun.
	var prefixesToExclude = [...]string{
		"/wiki/File",
		"/wiki/Geographic_coordinate_system",
		"/wiki/Help",
		"/wiki/ISBN_",
		"/wiki/ISSN_",
		"/wiki/Main_Page",
		"/wiki/Special",
		"/wiki/Wayback_Machine",
		"/wiki/Wikipedia:",
	}

	for _, prefix := range prefixesToExclude {
		if strings.HasPrefix(href, prefix) {
			return true
		}
	}
	return false
}

/**
isWikiArticle returns whether the link is a wikipedia article or not.
*/
func isWikiArticle(href string) bool {
	return !containsExcludedPrefix(href) && strings.HasPrefix(href, "/wiki/")
}

/**
getWikiURL validates and formats a given URL.
*/
func getWikiURL(index int, element *goquery.Selection) string {
	// See if the href attribute exists on the element
	href, exists := element.Attr("href")
	if exists && isWikiArticle(href) {
		return "https://en.m.wikipedia.org" + string(href)
	}
	return ""
}

/**
getChildWikiLinksForWikiPage gets a urls for a given page.

Right now we return all links. This is a method we can swap out with an API call to the Wikimedia API without impacting
additional code.
*/
func getChildWikiLinksForWikiPage(wikiUrl string) ([]string, error) {

	resp, err := http.Get(wikiUrl)
	if err != nil {
		// log.Print("Error getting page", err)
		return nil, err
	}

	document, queryError := goquery.NewDocumentFromReader(resp.Body)
	if queryError != nil {
		log.Fatal("Error loading HTTP response body. ", queryError)
		return nil, queryError
	}

	return document.Find("a").Map(getWikiURL), nil
}

/**
passChildNodesToChan takes in a node, extracts the URL, discovers the URLs on the page, and sends them onto the given channel.
*/
func passChildNodesToChan(a node, setChan chan<- []node, pathNum int) {
	//slice to add new links to
	childNodes := make([]node, 0)

	links, err := getChildWikiLinksForWikiPage(a.url)
	if err != nil {
		// log.Print("Error getting page")
		setChan <- childNodes
		return
	}

	for _, link := range links {
		if link != "" {
			childNodes = append(childNodes, node{url: link, pathId: pathNum, parent: &a})
		}
	}

	setChan <- childNodes
}

/**
getPath concatenates the path from node to the starting node with the path of node2 to the destination node.
*/
func getPath(node1 *node, node2 *node) string {
	path := ""
	for node1 != nil {
		path = node1.url + " -> " + path
		node1 = node1.parent
	}
	path = strings.TrimRight(path, " -> ")
	node2 = node2.parent
	for node2 != nil {
		path = path + " -> " + node2.url
		node2 = node2.parent
	}
	return path
}

/**
processNode concurrently processes nodes to determine if a path has been found, as well as queries Wikipedia for child
links.
*/
func processNode(
	urlToNodeMap map[string]node,
	urlsToVisitQueue []node,
	threadCount int,
	setChan chan<- []node,
	pathId int,
	paths []string,
) (map[string]node, []node, int, []string) {
	if len(urlsToVisitQueue) > 0 {
		var poppedNode node
		var val node
		var found bool

		poppedNode, urlsToVisitQueue = urlsToVisitQueue[0], urlsToVisitQueue[1:]

		// if the url is unknown to this half of the search, but is present in the map, it means we have found a pathId.
		if val, found = urlToNodeMap[poppedNode.url]; found && val.pathId != pathId {
			paths = append(paths, getPath(&poppedNode, &val))
		}

		if !found {
			// add url to map
			urlToNodeMap[poppedNode.url] = poppedNode

			threadCount++

			// launch function concurrently to GET new links on page concurrently
			go passChildNodesToChan(poppedNode, setChan, pathId)
		}
	}
	return urlToNodeMap, urlsToVisitQueue, threadCount, paths
}

/**
Build a node from an article and a given pathId.
*/
func buildNodeFromArticle(article string, pathId int) node {
	return node{
		url:    "https://en.m.wikipedia.org/wiki/" + strings.Replace(article, " ", "_", -1),
		pathId: pathId,
		parent: nil,
	}
}

/**
filterShortestPath returns the shortest path from a given list.
*/
func filterShortestPath(paths []string) string {
	shortestLength := len(paths[0])
	index := 0
	for i, path := range paths {
		if length := strings.Count(path, "->"); length < shortestLength {
			shortestLength = length
			index = i
		}
	}
	return paths[index]
}

/**
addThreadToNextQueue adds thread to the next queuee
*/
func addThreadToNextQueue(threadCount int, ch <-chan []node, nextQueue *[]node) {
	for i := 0; i < threadCount; i++ {
		*nextQueue = append(*nextQueue, <-ch...)
	}
}

/**
GetShortestPath returns a string of the shortest path using Breadth First Search.

The Main take away here, and is pretty clear in the pairs of lines below, is that we're concurrently searching from both
start and destination to more efficiently determine our path.
*/
func GetShortestPath(startArticle string, destinationArticle string) (string, bool) {

	startNode := buildNodeFromArticle(startArticle, PathIdFromTheFront)
	destinationNode := buildNodeFromArticle(destinationArticle, PathIdFromTheRear)

	urlToNodeMap := make(map[string]node)

	// initialize queues
	nextQ1 := make([]node, 1)
	nextQ1[0] = startNode
	nextQ2 := make([]node, 1)
	nextQ2[0] = destinationNode

	q1 := make([]node, 0)
	q2 := make([]node, 0)

	// iterate through the queues.
	for len(nextQ1) > 0 || len(nextQ2) > 0 {

		// bring current queue up to bat and initialize empty queue for next round
		q1, nextQ1 = nextQ1, make([]node, 0)
		q2, nextQ2 = nextQ2, make([]node, 0)

		// init variables
		paths := make([]string, 0)
		chan1 := make(chan []node) // we use channels to pass child links
		chan2 := make(chan []node)
		chan1ThreadCount := 0
		chan2ThreadCount := 0

		// process each element in this round's queues
		for len(q1) > 0 || len(q2) > 0 {
			urlToNodeMap, q1, chan1ThreadCount, paths = processNode(urlToNodeMap, q1, chan1ThreadCount, chan1, PathIdFromTheFront, paths)
			urlToNodeMap, q2, chan2ThreadCount, paths = processNode(urlToNodeMap, q2, chan2ThreadCount, chan2, PathIdFromTheRear, paths)
		}

		// this block is executed if a path is found.
		if len(paths) != 0 {
			return filterShortestPath(paths), true
		}

		// if no path is found...
		// this is the definition of side effect; something I would want to refactor normally, but a nice opportunity
		// to experiment with passing pointers.
		addThreadToNextQueue(chan1ThreadCount, chan1, &nextQ1)
		addThreadToNextQueue(chan2ThreadCount, chan2, &nextQ2)
	}
	return "No Path Found.", false
}

/**
translateStringToWikiUrl takes a string and creates a (theoretical) wikipedia url for it.
*/
func translateStringToWikiUrl(articleName string) string {
	return "https://en.wikipedia.org/wiki/" + strings.Replace(articleName, " ", "_", -1)
}

/**
wikiArticleExists takes in a string, translates it into an article url if it exists, and returns whether it exists.
*/
func wikiArticleExists(article string) (bool, string) {
	resp, err := http.Get(translateStringToWikiUrl(article))
	status := resp.StatusCode
	if err != nil || status == 404 {
		return false, "The article " + article + " does not exist."
	}
	return true, ""
}

/**
queryIsValid checks to see if the query is valid.

Notably, are the proper params present.
*/
func queryIsValid(query url.Values) (bool, string) {
	errorMsg := ""
	paramErrors := false

	start, ok := query["start"]
	startingArticle := start[0]
	if !ok || len(startingArticle) < 1 {
		errorMsg += "Url param 'start' is missing. "
		paramErrors = true
	}

	destination, ok := query["destination"]
	destinationArticle := destination[0]
	if !ok || len(destinationArticle) < 1 {
		errorMsg += "Url param 'destination' is missing. "
		paramErrors = true
	}

	if paramErrors == true {
		return false, errorMsg
	}

	return true, ""
}

/**
getArticlesFromQuery
*/
func getArticlesFromQuery(query url.Values) (string, string) {
	start, _ := query["start"]
	destination, _ := query["destination"]

	return start[0], destination[0]
}

/**
getErrorResponse constructs an error response.
*/
func getErrorResponse(errorMsg string) wikiRaceResponse {
	return wikiRaceResponse{
		Completed:      false,
		Destination:    "",
		ElapsedTimeSec: "",
		Message:        errorMsg,
		Path:           "",
		Start:          "",
	}
}

/**
WikiRacePureGoHandler is the handler that powers our API.
*/
func WikiRacePureGoHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	query := r.URL.Query()

	queryIsGood, errMsg := queryIsValid(query)
	if !queryIsGood {
		fmt.Println(errMsg)
		_ = json.NewEncoder(w).Encode(getErrorResponse(errMsg))
		return
	}

	startingArticle, destinationArticle := getArticlesFromQuery(query)

	startingPageExists, errMsg := wikiArticleExists(startingArticle)
	if !startingPageExists {
		fmt.Println(errMsg)
		_ = json.NewEncoder(w).Encode(getErrorResponse(errMsg))
		return
	}

	destinationPageExists, errMsg := wikiArticleExists(destinationArticle)
	if !destinationPageExists {
		fmt.Println(errMsg)
		_ = json.NewEncoder(w).Encode(getErrorResponse(errMsg))
		return
	}

	if startingPageExists && destinationPageExists {
		path, exists := GetShortestPath(startingArticle, destinationArticle)

		elapsedTime := time.Since(startTime)

		myResponse := wikiRaceResponse{
			Completed:      exists,
			Destination:    destinationArticle,
			ElapsedTimeSec: fmt.Sprintf("%f", elapsedTime.Seconds()),
			Message:        "",
			Path:           path,
			Start:          startingArticle,
		}

		json.NewEncoder(w).Encode(myResponse)
		return
	}

}
