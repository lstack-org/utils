package gt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMethod(t *testing.T) {
	go startServer(t)

	result := personExample{}

	url := "http://localhost/json-example"
	err := NewClient().
		GET(url).
		Do().
		InTo(&result, JSON)
	assert.Nil(t, err)
	t.Logf("%+v", result)
}

func TestHttpConcurrency(t *testing.T) {
	go startServer(t)

	for i := 0; i < 100; i++ {
		go func() {
			result := personExample{}
			url := "http://localhost/json-example"
			err := NewDefaultClient().
				GET(url).
				Do().
				InTo(&result, JSON)
			assert.Nil(t, err)
		}()
	}
	time.Sleep(10 * time.Second)

}

func TestGetWithHeader(t *testing.T) {
	go startServer(t)

	header := map[string][]string{
		"role": []string{"hello"},
	}
	url := "http://localhost/header"
	c := NewDefaultClient().
		GET(url).
		AddHeader(header).
		EnableLog(LogDebug).
		Do()
	assert.Nil(t, c.Err)
}

func TestBodyDecode(t *testing.T) {
	go startServer(t)

	header := map[string][]string{
		"role": []string{"hello"},
	}
	var result string
	url := "http://localhost/header"
	err := NewDefaultClient().
		GET(url).
		AddHeader(header).
		EnableLog(LogDebug).
		Do().
		InTo(&result, BODY)
	assert.Nil(t, err)
	t.Log("result: ", result)
}
