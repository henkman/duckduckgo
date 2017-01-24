package duckduckgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

type Session struct {
	cli http.Client
}

type WebResult struct {
	Url string
}

type ImageResult struct {
	Url string
}

type VideoResult struct {
	Id string
}

var (
	reWeb = regexp.MustCompile("\"u\":\"([^\"]+)\"")
	reVqd = regexp.MustCompile("vqd='([^']+)'")
)

func (sess *Session) Init() error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	sess.cli.Jar = jar
	return nil
}

func (sess *Session) IsInitialized() bool {
	return sess.cli.Jar != nil
}

func (sess *Session) Web(query string, offset uint) ([]WebResult, error) {
	fd := url.Values{
		"q":  []string{query},
		"b":  []string{},
		"kl": []string{"us-en"},
	}.Encode()
	req, err := sess.newRequest("POST", "https://duckduckgo.com/html/",
		bytes.NewBufferString(fd))
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	res, err := sess.cli.Do(req)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}
	results := make([]WebResult, 0, 16)
	doc.Find(".result__a").Each(func(i int, s *goquery.Selection) {
		if a, ok := s.Attr("href"); ok {
			results = append(results, WebResult{Url: a})
		}
	})
	return results, nil
}

func (sess *Session) Images(query string, offset uint) ([]ImageResult, error) {
	var vqd string
	{
		u := url.URL{
			Scheme: "https",
			Host:   "duckduckgo.com",
			Path:   "/",
			RawQuery: url.Values{
				"q":   []string{query},
				"iax": []string{"1"},
				"ia":  []string{"images"},
				"t":   []string{"h_"},
			}.Encode(),
		}
		res, err := sess.request("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}
		source, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return nil, err
		}
		m := reVqd.FindStringSubmatch(string(source))
		if m == nil {
			return nil, errors.New("invalid response")
		}
		vqd = m[1]
	}
	var results struct {
		Results []struct {
			Image string `json:"image"`
		} `json:"results"`
	}
	{
		u := url.URL{
			Scheme: "https",
			Host:   "duckduckgo.com",
			Path:   "/i.js",
			RawQuery: url.Values{
				"l":   []string{"us-en"},
				"o":   []string{"json"},
				"q":   []string{query},
				"vqd": []string{vqd},
				"f":   []string{},
				"s":   []string{fmt.Sprint(offset)},
			}.Encode(),
		}
		res, err := sess.request("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}
		if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
			res.Body.Close()
			return nil, err
		}
		res.Body.Close()
	}
	if len(results.Results) == 0 {
		return []ImageResult{}, nil
	}
	images := make([]ImageResult, len(results.Results))
	for i := 0; i < len(results.Results); i++ {
		images[i].Url = results.Results[i].Image
	}
	return images, nil
}

func (sess *Session) Videos(query string, offset uint) ([]VideoResult, error) {
	var vqd string
	{
		u := url.URL{
			Scheme: "https",
			Host:   "duckduckgo.com",
			Path:   "/",
			RawQuery: url.Values{
				"q":   []string{query},
				"iax": []string{"1"},
				"ia":  []string{"videos"},
				"t":   []string{"h_"},
			}.Encode(),
		}
		res, err := sess.request("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}
		source, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return nil, err
		}
		m := reVqd.FindStringSubmatch(string(source))
		if m == nil {
			return nil, errors.New("invalid response")
		}
		vqd = m[1]
	}
	var results struct {
		Results []struct {
			Provider string `json:"provider"`
			ID       string `json:"id"`
		} `json:"results"`
	}
	{
		u := url.URL{
			Scheme: "https",
			Host:   "duckduckgo.com",
			Path:   "/v.js",
			RawQuery: url.Values{
				"o":      []string{"json"},
				"strict": []string{"1"},
				"q":      []string{query},
				"vqd":    []string{vqd},
				"s":      []string{fmt.Sprint(offset)},
			}.Encode(),
		}
		res, err := sess.request("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}
		if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
			res.Body.Close()
			return nil, err
		}
		res.Body.Close()
	}
	ids := make([]VideoResult, 0, len(results.Results))
	for _, v := range results.Results {
		if v.Provider == "YouTube" {
			ids = append(ids, VideoResult{Id: v.ID})
		}
	}
	if len(ids) == 0 {
		return []VideoResult{}, nil
	}
	return ids, nil
}

func (sess *Session) request(method, url string, body io.Reader) (*http.Response, error) {
	req, err := sess.newRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return sess.cli.Do(req)
}

func (sess *Session) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("user-agent",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.94 Safari/537.36")
	return req, nil
}
