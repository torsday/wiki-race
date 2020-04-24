package puregorace

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_WikiRacePureGoHandler(t *testing.T) {
	req, err := http.NewRequest(
		"GET",
		"/wiki-race/goLang?start=St.%20Olaf%20College&destination=pantheon%20(religion)",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(WikiRacePureGoHandler)

	handler.ServeHTTP(rr, req)

	//Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func Test_isWikiArticle(t *testing.T) {
	type args struct {
		href string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Should recognize when href isn't a wikipedia page", args{href: "/yowzer/wiki/yowza"}, false},
		{"Should recognize a wikipedia page", args{href: "/wiki/Madison,_Wisconsin"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isWikiArticle(tt.args.href); got != tt.want {
				t.Errorf("isWikiArticle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_containsExcludedPrefix(t *testing.T) {
	type args struct {
		href string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"should recognize excluded prefixes", args{"/wiki/Special"}, true},
		{"should return false when excluded prefix not found", args{"yabba dabba do"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsExcludedPrefix(tt.args.href); got != tt.want {
				t.Errorf("containsExcludedPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildNodeFromArticle(t *testing.T) {
	type args struct {
		article string
		pathId  int
	}
	tests := []struct {
		name string
		args args
		want node
	}{
		{
			"should build node from article",
			args{article: "Chris Torstenson", pathId: PathIdFromTheRear},
			node{url: "https://en.m.wikipedia.org/wiki/Chris_Torstenson", pathId: PathIdFromTheRear, parent: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildNodeFromArticle(tt.args.article, tt.args.pathId); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildNodeFromArticle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_translateStringToWikiUrl(t *testing.T) {
	type args struct {
		articleName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"should create a wiki url from a given string",
			args{articleName: "Peter Jenkins"},
			"https://en.wikipedia.org/wiki/Peter_Jenkins",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := translateStringToWikiUrl(tt.args.articleName); got != tt.want {
				t.Errorf("translateStringToWikiUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
