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
	imgs, err := sess.Web("cat", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(imgs) == 0 {
		t.Logf("nothing found")
	}
	for _, img := range imgs {
		fmt.Println(img.Url)
	}
}

func TestImage(t *testing.T) {
	return
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
