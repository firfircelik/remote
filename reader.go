package remote

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// Option is an option to set on remote reader
type Option func(*Reader)

// Reader is a client to read remote bytes or json
// Should be created via NewRemoteReader to configure
// Defaults 1 retry and 5 seconds timeout
type Reader struct {
	retry         uint
	timeout       time.Duration
	skipTLSVerify bool
	userAgent     string
}

// NewReader creates a new remote reader with defaults
func NewReader(options ...Option) *Reader {
	r := &Reader{
		retry:     1,
		timeout:   5 * time.Second,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.81 Safari/537.36", // nolint: lll
	}
	for _, option := range options {
		option(r)
	}
	return r
}

// Retry option for remote reader
func Retry(retry uint) Option { return func(r *Reader) { r.retry = retry } }

// Timeout option for remote reader
func Timeout(timeout time.Duration) Option {
	return func(r *Reader) {
		r.timeout = timeout
	}
}

// SkipTLSVerify option for remote reader to skip TLS Certificate verification
func SkipTLSVerify() Option { return func(r *Reader) { r.skipTLSVerify = true } }

// UserAgent option for remote reader sets the user agent header string for the request
func UserAgent(userAgent string) Option { return func(r *Reader) { r.userAgent = userAgent } }

// Read returns response from given url with configured reader
func (r *Reader) Read(url string) (*http.Response, error) {
	var resp *http.Response
	var err error
	var i uint
	for i = 0; i < r.retry; i++ {
		if resp, err = r.get(url); err == nil || !isTimeoutErr(err) {
			return resp, errors.Wrap(err, "can't get url")
		}
	}
	return resp, errors.Wrap(err, "can't read url")
}

// Bytes reads bytes from given url with configured reader
func (r *Reader) Bytes(url string) ([]byte, error) {
	resp, err := r.Read(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Got %q: can't read given url %q", resp.Status, url)
	}
	b, err := ioutil.ReadAll(resp.Body)
	return b, errors.Wrap(err, "can't read body of response")
}

// JSON reads bytes from given url with configured reader and decodes body into the destination
func (r *Reader) JSON(url string, dest interface{}) error {
	resp, err := r.Read(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("Got %q: can't read given url %q", resp.Status, url)
	}
	return DecodeAsJSON(resp.Body, dest)
}

func (r *Reader) get(url string) (*http.Response, error) {
	client := &http.Client{Timeout: r.timeout}
	if r.skipTLSVerify {
		client.Transport = &http.Transport{
			/* #nosec */
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", r.userAgent)
	return client.Do(req)
}

// isTimeoutErr checks if given error is a timeout
func isTimeoutErr(err error) bool {
	if err == nil {
		return false
	}
	urlError, ok := err.(*url.Error)
	return ok && urlError.Timeout()
}

// DecodeAsJSON decodes given reader into destination
// assuming content is json
func DecodeAsJSON(r io.Reader, dest interface{}) error {
	err := json.NewDecoder(r).Decode(dest)
	if err == io.EOF {
		return nil
	}
	return errors.Wrap(err, "can't decode json")
}
