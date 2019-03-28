package remote

import (
	"log"
	"net/http"
	"testing"
)

func TestRetry(t *testing.T) {
	r := NewReader(Retry(3))
	if r.retry != 3 {
		t.Error("failed to set reader's retry")
	}
}

func TestSkipTLSVerify(t *testing.T) {
	r := NewReader(SkipTLSVerify())
	if !r.skipTLSVerify {
		t.Error("failed to set reader's skipTLSVerify to true")
	}
}

func TestTimeout(t *testing.T) {
	r := NewReader(Timeout(3))
	if r.timeout != 3 {
		t.Error("failed to set reader's timeout")
	}
}

func TestUserAgent(t *testing.T) {
	newAgent:= "Mozilla/5.0 (Windows NT 6.1; Win64; x64) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/63.0.3239.132 Safari/537.36"
	r := NewReader(UserAgent(newAgent))
	if r.userAgent != newAgent {
		t.Error("failed to set user agent")
	}
}

func TestReader_Read(t *testing.T) {
	_, err := NewReader().Read("https://google.com")
	if err != nil {
		t.Error(err)
	}
}

func TestReader_Bytes(t *testing.T) {
	content, err := NewReader().Bytes("https://google.com")
	if err != nil {
		t.Error(err)
	}
	if len(content) == 0 {
		t.Error("Bytes return empty content")
	}
}

func TestReader_JSON(t *testing.T) {
	// start a small http server
	http.HandleFunc("/json/valid", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("{\"content\": 1}"))
		if err != nil {
			log.Fatal("failed to write back")
		}
	})
	http.HandleFunc("/json/invalid", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("boom"))
		if err != nil {
			log.Fatal("failed to write back")
		}
	})
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// try to read valid and invalid json
	url := "http://localhost:8080/json"
	type testData struct {Content int `json:"content"`}
	if err := NewReader().JSON(url + "/invalid",  &testData{}); err == nil {
		t.Error("failed to read invalid json response")
	}
	result := &testData{}
	if err := NewReader().JSON(url + "/valid", result); err != nil {
		t.Error("failed to read json response")
	}
	if result.Content != 1 {
		t.Error("invalid result Json", result.Content)
	}
}
