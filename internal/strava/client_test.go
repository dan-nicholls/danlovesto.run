package strava

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Test Mock Stores
type fakeTokenStore struct {
	tokens map[string]Token
	err    error
}

func (f *fakeTokenStore) Load(key string) (Token, error) {
	if f.err != nil {
		return Token{}, f.err
	}
	t, ok := f.tokens[key]
	if !ok {
		return Token{}, fmt.Errorf("no token for key %q", key)
	}
	return t, nil
}

func (f *fakeTokenStore) Save(key string, token Token) error {
	if f.err != nil {
		return f.err
	}
	if f.tokens == nil {
		f.tokens = make(map[string]Token)
	}
	f.tokens[key] = token
	return nil
}

// Test Mock HTTP client
type fakeHTTPDoer struct {
	responses []*http.Response
	err       error
	calls     []*http.Request
}

func (f *fakeHTTPDoer) Do(req *http.Request) (*http.Response, error) {
	f.calls = append(f.calls, req)
	if f.err != nil {
		return nil, f.err
	}
	if len(f.responses) == 0 {
		return nil, fmt.Errorf("No more responses")
	}
	r := f.responses[0]
	f.responses = f.responses[1:]
	return r, nil
}

type scriptedHTTPDoer struct {
	calls []*http.Request
	reqFn func(req *http.Request) (*http.Response, error)
}

func (s *scriptedHTTPDoer) Do(req *http.Request) (*http.Response, error) {
	s.calls = append(s.calls, req)
	if s.reqFn != nil {
		return s.reqFn(req)
	}
	return nil, nil
}

func jsonResp(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// Tests

// FetchAllActs - Happy Paths
func TestFetchAllActivities_SinglePage_Success(t *testing.T) {
	ts := &fakeTokenStore{
		tokens: map[string]Token{
			"strava": {
				AccessToken:  "ACCESS",
				RefreshToken: "REFRESH",
				ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			},
		},
	}

	cfg := StravaConfig{ClientID: "id", ClientSecret: "secret"}
	http := &fakeHTTPDoer{
		responses: []*http.Response{
			jsonResp(200, `[{"id": 1}, {"id": 2}]`),
			jsonResp(200, `[]`),
		},
	}

	client := NewClientWithHTTP(cfg, ts, http)

	ctx := context.Background()
	acts, err := client.FetchAllActivities(ctx, 0, 0, 0, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := len(acts), 2; got != want {
		t.Fatalf("expected %d acts, got %d", want, got)
	}

	if got, want := len(http.calls), 2; got != want {
		t.Errorf("expected %d http calls, got %d", want, got)
	}

	for _, req := range http.calls {
		if got, want := req.Header.Get("Authorization"), "Bearer ACCESS"; got != want {
			t.Errorf("expected Authorization=%s, got %s", want, got)
		}
	}

}

func TestFetchAllActivities_MultiPagesTillEmpty_Success(t *testing.T) {
	ts := &fakeTokenStore{
		tokens: map[string]Token{
			"strava": {
				AccessToken:  "ACCESS",
				RefreshToken: "REFRESH",
				ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			},
		},
	}
	cfg := StravaConfig{ClientID: "id", ClientSecret: "secret"}
	http := &fakeHTTPDoer{
		responses: []*http.Response{
			jsonResp(200, `[{"id": 1}, {"id": 2}]`),
			jsonResp(200, `[{"id": 3}, {"id": 4}]`),
			jsonResp(200, `[]`),
		},
	}

	client := NewClientWithHTTP(cfg, ts, http)

	ctx := context.Background()
	acts, err := client.FetchAllActivities(ctx, 0, 0, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := len(acts), 4; got != want {
		t.Errorf("expected %d acts, got %d", want, got)
	}

	if got, want := len(http.calls), 3; got != want {
		t.Errorf("expected %d http calls, got %d", want, got)
	}

	for i, req := range http.calls {
		if got, want := req.Header.Get("Authorization"), "Bearer ACCESS"; got != want {
			t.Errorf("expected Authorization=%s in %d request, got %s", want, i, got)
		}
	}
}
func TestFetchAllActivities_RespectsMaxPages(t *testing.T) {
	ts := &fakeTokenStore{
		tokens: map[string]Token{
			"strava": {
				AccessToken:  "ACCESS",
				RefreshToken: "REFRESH",
				ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			},
		},
	}
	cfg := StravaConfig{ClientID: "id", ClientSecret: "secret"}
	http := &fakeHTTPDoer{
		responses: []*http.Response{
			jsonResp(200, `[{"id": 1}, {"id": 2}]`),
			jsonResp(200, `[{"id": 3}, {"id": 4}]`),
			jsonResp(200, `[{"id": 5}, {"id": 6}]`),
			jsonResp(200, `[{"id": 7}, {"id": 8}]`),
			jsonResp(200, `[]`),
		},
	}

	client := NewClientWithHTTP(cfg, ts, http)

	ctx := context.Background()
	acts, err := client.FetchAllActivities(ctx, 0, 0, 2, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := len(acts), 4; got != want {
		t.Errorf("expected %d, acts, got %d", want, got)
	}

	if got, want := len(http.calls), 2; got != want {
		t.Errorf("expected %d http calls, got %d", want, got)
	}

}
func TestFetchAllActivites_SetBeforeAndAfterQueryParams(t *testing.T) {
	ts := &fakeTokenStore{
		tokens: map[string]Token{
			"strava": {
				AccessToken:  "ACCESS",
				RefreshToken: "REFRESH",
				ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			},
		},
	}
	cfg := StravaConfig{ClientID: "id", ClientSecret: "secret"}
	http := &fakeHTTPDoer{
		responses: []*http.Response{
			jsonResp(200, `[{"id": 1}, {"id": 2}]`),
			jsonResp(200, `[{"id": 3}, {"id": 4}]`),
			jsonResp(200, `[]`),
		},
	}

	client := NewClientWithHTTP(cfg, ts, http)

	ctx := context.Background()
	timeAfter := int64(10)
	timeBefore := int64(20)
	_, err := client.FetchAllActivities(ctx, timeAfter, timeBefore, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := len(http.calls), 3; got != want {
		t.Errorf("expected %d http calls, got %d", want, got)
	}

	for i, req := range http.calls {
		if got, want := req.Header.Get("Authorization"), "Bearer ACCESS"; got != want {
			t.Errorf("expected Authorization=%s in %d request, got %s", want, i, got)
		}

		expectedAfter := strconv.Itoa(int(timeAfter))
		if got, want := req.URL.Query().Get("after"), expectedAfter; got != want {
			t.Errorf("expected query param \"after\"=%s in request %d, got %s", want, i, got)
		}

		expectedBefore := strconv.Itoa(int(timeBefore))
		if got, want := req.URL.Query().Get("before"), expectedBefore; got != want {
			t.Errorf("expected query param \"before\"=%s in request %d, got %s", want, i, got)
		}
	}
}

// FetchAllActs - Auth / Refresh / Rate Limiting
func TestFetchAllActivities_401_RefreshThenRetry_Success(t *testing.T) {
	ts := &fakeTokenStore{
		tokens: map[string]Token{
			"strava": {
				AccessToken:  "EXPIRED",
				RefreshToken: "OLD_REFRESH",
				ExpiresAt:    time.Now().Add(time.Hour).Unix(),
			},
		},
	}
	cfg := StravaConfig{ClientID: "id", ClientSecret: "secret"}
	fakeHttp := &scriptedHTTPDoer{}
	fakeHttp.reqFn = func(req *http.Request) (*http.Response, error) {
		url := req.URL.String()

		// First Call
		if strings.Contains(url, "/athlete/activities") && len(fakeHttp.calls) == 1 {
			return jsonResp(401, `{"message": "unauthorized"}`), nil
		}

		// Second Call
		if strings.Contains(url, "/oauth/token") && len(fakeHttp.calls) == 2 {
			return jsonResp(200, `{
					"access_token": "ACCESS",
					"refresh_token": "NEW_REFRESH",
					"token_type": "Bearer",
					"expires_at": 4102444800,
					"expires_in": 21600
				}`), nil
		}

		if strings.Contains(url, "/athlete/activities") && len(fakeHttp.calls) == 3 {
			return jsonResp(200, `[{"id": 1}, {"id": 2}]`), nil
		}

		if strings.Contains(url, "/athlete/activities") && len(fakeHttp.calls) == 4 {
			return jsonResp(200, `[]`), nil
		}
		return nil, fmt.Errorf("unexpected request: %s", url)
	}

	client := NewClientWithHTTP(cfg, ts, fakeHttp)

	ctx := context.Background()
	acts, err := client.FetchAllActivities(ctx, 0, 0, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := len(acts), 2; got != want {
		t.Errorf("expected %d acts, got %d", want, got)
	}

	if got, want := len(fakeHttp.calls), 4; got != want {
		t.Errorf("expected %d http calls, got %d", want, got)
	}
}
func TestFetchAllActivities_401_RefreshFails_ReturnsError(t *testing.T) {}
func TestFetchAllActivities_429_ThenRetry_Success(t *testing.T)         {}

// FetchAllActs - Error Paths
func TestFetchAllActivities_HttpDoError_ReturnsError(t *testing.T)           {}
func TestFetchAllActivities_Non2xxStatus_ReturnsError(t *testing.T)          {}
func TestFetchAllActivities_InvalidJSON_ReturnsError(t *testing.T)           {}
func TestFetchAllActivities_ContextCancelledBeforeFirstRequest(t *testing.T) {}
func TestFetchAllActivities_ContextCancelledInLoop(t *testing.T)             {}

// GetActivityDetails - Happy Paths
func TestGetActivityDetails_Success(t *testing.T) {}

// GetActivityDetails - Error Paths
func TestGetActivityDetails_InvalidID_ReturnsError(t *testing.T)             {}
func TestGetActivityDetails_ContextCancelledBeforeFirstRequest(t *testing.T) {}
func TestGetActivityDetails_ContextCancelledInLoop(t *testing.T)             {}
func TestGetActivityDetails_HttpDoError_ReturnsError(t *testing.T)           {}
func TestGetActivityDetails_Non2xxStatus_ReturnsError(t *testing.T)          {}
func TestGetActivityDetails_InvalidJSON_ReturnsError(t *testing.T)           {}

// GetActivityDetails - Auth/Refresh/RateLimiting
func TestGetActivityDetails_401_RefreshThenRetry_Success(t *testing.T) {}
func TestGetActivityDetails_401_RefreshThenRetry_Fail(t *testing.T)    {}
func TestGetActivityDetails_429_ThenRetry_Success(t *testing.T)        {}

// ensureValidToken
func TestEnsureValidToken_ErrorLoadingTokenFromStore(t *testing.T)     {}
func TestEnsureValidToken_EmptyRefreshToken_RunsOAuth(t *testing.T)    {}
func TestEnsureValidToken_EmptyAccessToken_RunsRefresh(t *testing.T)   {}
func TestEnsureValidToken_ExpiredAccessToken_RunsRefresh(t *testing.T) {}
func TestEnsureValidToken_ValidToken_NoAction(t *testing.T)            {}

// refreshAccessToken
func TestRefreshAccessToken_SuccessStoreToken(t *testing.T)                      {}
func TestRefreshAccessToken_ErrorLoadingTokenFromStore(t *testing.T)             {}
func TestRefreshAccessToken_MissingRefreshToken(t *testing.T)                    {}
func TestRefreshAccessToken_HttpDoError_ReturnsError(t *testing.T)               {}
func TestRefreshAccessToken_ContextCancelled_ReturnsERror(t *testing.T)          {}
func TestRefreshAccessToken_Non2xxStatus_ReturnsErrorWithBody(t *testing.T)      {}
func TestRefreshAccessToken_InvalidJSON_ReturnsError(t *testing.T)               {}
func TestAccessRefreshToken_EmptyAccessOrRefreshToken_ReturnsError(t *testing.T) {}

// isExpired
func TestIsExpired_ReturnsTrueForPastTime(t *testing.T)    {}
func TestIsExpired_ReturnsFalseForFutureTime(t *testing.T) {}

// maybe - runOAuth
// maybe - RecordJSON
