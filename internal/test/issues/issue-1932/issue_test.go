package issue1932

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func doGet(t *testing.T, app *fiber.App, rawURL string) (*http.Response, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Invalid url: %s", rawURL)
	}

	req := httptest.NewRequest("GET", u.RequestURI(), nil)
	req.Header.Add("Accept", "application/json")
	req.Host = u.Host

	return app.Test(req)
}

// TestIssue1932
func TestIssue1932(t *testing.T) {

	r := fiber.New()
	svr := NewServer()
	RegisterHandlers(r, svr)

	iters := 100000

	wg := sync.WaitGroup{}
	wg.Add(iters)

	for i := 0; i < iters; i++ {
		go func(idx int) {
			defer wg.Done()
			strIdx := strconv.Itoa(idx)
			rr, _ := doGet(t, r, fmt.Sprintf("/param/%s", strIdx))
			if rr == nil {
				return
			}

			assert.Equal(t, http.StatusOK, rr.StatusCode)
			var resultMsg SimpleParam
			err := json.NewDecoder(rr.Body).Decode(&resultMsg)
			assert.NoError(t, err, "error unmarshaling response")
			assert.NotEmpty(t, resultMsg)
			assert.Equal(t, strIdx, resultMsg.Message)
		}(i)
	}

	wg.Wait()
}
