package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_desanitize64(t *testing.T) {
	examples := map[string]string{
		"test":        "test",
		"test+test+":  "test-test-",
		"test/test/":  "test_test_",
		"test=test==": "test.test..",
	}

	for expected, example := range examples {
		assert.Equal(t, expected, desanitize64(example))
	}
}

func Test_cleanQuery(t *testing.T) {
	assert.Equal(t, "a\nb\nc", cleanQuery("a\nb\nc"))
	assert.Equal(t, "", cleanQuery("--something"))
	assert.Equal(t, "test", cleanQuery("--test\ntest\n   -- test\n"))
}

func Test_sanitizeFilename(t *testing.T) {
	examples := map[string]string{
		"foo":              "foo",
		"fooBar":           "fooBar",
		"foo.bar":          "foo_bar",
		`"foo"."bar"`:      "foo_bar",
		"!@#$foo.&&*(&bar": "foo_bar",
	}

	for given, expected := range examples {
		t.Run(given, func(t *testing.T) {
			assert.Equal(t, expected, sanitizeFilename(given))
		})
	}
}

func Test_getSessionId(t *testing.T) {
	req := &http.Request{Header: http.Header{}}
	req.Header.Add("x-session-id", "token")
	assert.Equal(t, "token", getSessionId(req))

	req = &http.Request{}
	req.URL, _ = url.Parse("http://foobar/?_session_id=token")
	assert.Equal(t, "token", getSessionId(req))
}

func Test_serveResult(t *testing.T) {
	server := gin.Default()
	server.GET("/good", func(c *gin.Context) {
		serveResult(c, gin.H{"foo": "bar"}, nil)
	})
	server.GET("/bad", func(c *gin.Context) {
		serveResult(c, nil, errors.New("message"))
	})
	server.GET("/nodata", func(c *gin.Context) {
		serveResult(c, nil, nil)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/good", nil)
	server.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `{"foo":"bar"}`, w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/bad", nil)
	server.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
	assert.Equal(t, `{"error":"message","status":400}`, w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/nodata", nil)
	server.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `null`, w.Body.String())
}

func TestExtractURLParamsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req, _ := http.NewRequest("GET", "/api/query", nil)
	c := &gin.Context{Request: req}

	params := extractURLParams(c)

	assert.Empty(t, params)
}

func TestExtractURLParamsGSRPattern(t *testing.T) {
	gin.SetMode(gin.TestMode)

	values := url.Values{}
	values.Set("gsr_client", "test-client")
	values.Set("gsr_inst", "test-instance")
	values.Set("gsr_environment", "production")

	req, _ := http.NewRequest("GET", "/api/query?"+values.Encode(), nil)
	c := &gin.Context{Request: req}

	params := extractURLParams(c)

	assert.Len(t, params, 3)
	assert.Equal(t, "test-client", params["gsr_client"])
	assert.Equal(t, "test-instance", params["gsr_inst"])
	assert.Equal(t, "production", params["gsr_environment"])
}

func TestExtractURLParamsMixedPatterns(t *testing.T) {
	gin.SetMode(gin.TestMode)

	values := url.Values{}
	values.Set("gsr_client", "test-client")
	values.Set("tenant_id", "123")
	values.Set("user_role", "admin")
	values.Set("ignored_param", "should-not-appear")
	values.Set("primaryColor", "#007bff") // UI parameter, should be ignored

	req, _ := http.NewRequest("GET", "/api/query?"+values.Encode(), nil)
	c := &gin.Context{Request: req}

	params := extractURLParams(c)

	assert.Len(t, params, 3)
	assert.Equal(t, "test-client", params["gsr_client"])
	assert.Equal(t, "123", params["tenant_id"])
	assert.Equal(t, "admin", params["user_role"])
	assert.NotContains(t, params, "ignored_param")
	assert.NotContains(t, params, "primaryColor")
}

func TestExtractURLParamsInvalidPatterns(t *testing.T) {
	gin.SetMode(gin.TestMode)

	values := url.Values{}
	values.Set("gsr", "should-not-match")     // No underscore
	values.Set("gsr_", "should-not-match")    // No word after underscore
	values.Set("_client", "should-not-match") // Starts with underscore
	values.Set("tenant", "should-not-match")  // No underscore

	req, _ := http.NewRequest("GET", "/api/query?"+values.Encode(), nil)
	c := &gin.Context{Request: req}

	params := extractURLParams(c)

	assert.Empty(t, params)
}
