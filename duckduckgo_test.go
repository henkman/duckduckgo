package duckduckgo

import (
	"fmt"
	"testing"
)

func TestWeb(t *testing.T) {
	var sess Session
	if err := sess.Init(); err != nil {
		t.Fatal(err)
	}
	webs, err := sess.Web("cat", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(webs) == 0 {
		t.Logf("nothing found")
	}
	for _, web := range webs {
		fmt.Println(web.Url)
	}
}

func TestImage(t *testing.T) {
	var sess Session
	if err := sess.Init(); err != nil {
		t.Fatal(err)
	}
	imgs, err := sess.Images("cat", 50)
	if err != nil {
		t.Fatal(err)
	}
	for _, img := range imgs {
		fmt.Println(img.Url)
	}
}
