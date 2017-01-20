package duckduckgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
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

func (sess *Session) Web(query string, offset int) ([]WebResult, error) {
	var vqd string
	{
		u := url.URL{
			Scheme: "https",
			Host:   "duckduckgo.com",
			Path:   "/",
			RawQuery: url.Values{
				"q":  []string{query},
				"ia": []string{"web"},
				"t":  []string{"h_"},
			}.Encode(),
		}
		r, err := sess.request("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}
		source, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			return nil, err
		}
		m := reVqd.FindStringSubmatch(string(source))
		if m == nil {
			return nil, errors.New("invalid response")
		}
		vqd = m[1]
	}
	results := make([]WebResult, 0, 16)
	{
		u := url.URL{
			Scheme: "https",
			Host:   "duckduckgo.com",
			Path:   "/d.js",
			RawQuery: url.Values{
				"q":      []string{query},
				"t":      []string{"D"},
				"l":      []string{"us-en"},
				"s":      []string{fmt.Sprint(offset)},
				"a":      []string{"hs"},
				"dl":     []string{"en"},
				"ct":     []string{"DE"},
				"ss_mkt": []string{"us"},
				"vqd":    []string{vqd},
				"p":      []string{"1"},
				"sp":     []string{"0"},
				"yhs":    []string{"1"},
			}.Encode(),
		}
		fmt.Println(u.String())
		r, err := sess.request("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}
		source, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			return nil, err
		}
		fmt.Println(string(source))
		m := reWeb.FindAllStringSubmatch(string(source), -1)
		if m == nil {
			return nil, errors.New("invalid response")
		}
		for _, el := range m {
			results = append(results, WebResult{Url: el[1]})
		}
	}
	return results, nil
}

func (sess *Session) Images(query string, offset int) ([]ImageResult, error) {
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
		r, err := sess.request("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}
		source, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
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
		r, err := sess.request("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}
		if err := json.NewDecoder(r.Body).Decode(&results); err != nil {
			r.Body.Close()
			return nil, err
		}
		r.Body.Close()
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

func (sess *Session) request(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.94 Safari/537.36")
	return sess.cli.Do(req)
}
