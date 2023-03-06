package gt

import (
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"
)

var feed = `<?xml version="1.0" encoding="UTF-8"?>
		<rss>
			<channel>
				<title>Going Go Programming</title>
				<description>Golang : https://github.com/goinggo</description>
				<link>http://www.goinggo.net/</link>
				<item>
					<pubDate>Sun, 15 Mar 2015 15:04:00 +0000</pubDate>
					<title>Object Oriented Programming Mechanics</title>
					<description>Go is an object oriented language.</description>
					<link>http://www.goinggo.net/2015/03/object-oriented</link>
				</item>
		</channel>
	</rss>`

type personExample struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var (
	serverOnce sync.Once
)

func mockServer(t *testing.T) {
	xmlHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set(contentType, xmlContentType)
		_, err := w.Write([]byte(feed))
		if err != nil {
			t.Fatal(err)
		}
	}

	jsonHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set(contentType, jsonContentType)
		seedNum := time.Now().UnixNano()
		rand.Seed(seedNum)
		time.Sleep(time.Duration(int64(rand.Intn(7))) * time.Second)
		_, err := w.Write([]byte(`{"name":"zhangSan", "age": 1}`))
		if err != nil {
			t.Fatal(err)
		}
	}

	headerHandler := func(w http.ResponseWriter, r *http.Request) {
		role := r.Header.Get("role")
		if role == "" {
			t.Fatal("not found header key role")
		}
		_, _ = w.Write([]byte("success"))
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/json-example", jsonHandler)
	mux.HandleFunc("/xml-example", xmlHandler)
	mux.HandleFunc("/header", headerHandler)
	err := http.ListenAndServe(":80", mux)
	if err != nil {
		t.Fatal(err)
	}

}

func startServer(t *testing.T) {
	serverOnce.Do(func() {
		mockServer(t)
	})
}
