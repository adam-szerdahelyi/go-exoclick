package exoclick

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"golang.org/x/time/rate"
)

const (
	Version          = "v0.0.4"
	defaultBaseUrl   = "https://api.exoclick.com/v2/"
	defaultUserAgent = "go-exoclick" + "/" + Version

	headerRateLimit     = "X-Ratelimit-Limit"
	headerRateRemaining = "X-Ratelimit-Remaining"
	headerRateReset     = "X-Ratelimit-Reset"

	rateGlobal          = "Global"
	rateLimitGlobal     = rate.Limit(2.5)
	rateStatistics      = "Statistics"
	rateLimitStatistics = rate.Limit(0.5)
)

var errNonNilContext = errors.New("context must be non-nil")

type Client struct {
	client *http.Client

	BaseURL   *url.URL
	apiToken  string
	authToken AuthToken
	UserAgent string

	RateLimiter map[string]*rate.Limiter

	common service

	Campaigns   *CampaignsService
	Category    *CategoryService
	File        *FileService
	Marketplace *MarketplaceService

	Statistics *StatisticsService
}

type service struct {
	client *Client
}

type AuthToken struct {
	Token           string `json:"token"`
	TokenExpiry     int64  `json:"expires_in"`
	TokenExpiryDate time.Time
}

func NewClient(httpClient *http.Client, apiToken string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	httpClient2 := *httpClient
	c := &Client{client: &httpClient2}

	c.apiToken = apiToken

	c.initialize()

	return c
}

func (c *Client) initialize() {
	if c.client == nil {
		c.client = &http.Client{}
	}

	if c.BaseURL == nil {
		c.BaseURL, _ = url.Parse(defaultBaseUrl)
	}

	if c.UserAgent == "" {
		c.UserAgent = defaultUserAgent
	}

	transport := c.client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	c.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if !strings.Contains(req.URL.Path, "login") {
			if c.authToken.TokenExpiryDate.Before(time.Now()) {
				err := c.Login()
				if err != nil {
					return nil, err
				}
			}

			req = req.Clone(req.Context())
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken.Token))
		}

		return transport.RoundTrip(req)
	})

	c.RateLimiter = make(map[string]*rate.Limiter)
	c.RateLimiter[rateGlobal] = rate.NewLimiter(rateLimitGlobal, 1)
	c.RateLimiter[rateStatistics] = rate.NewLimiter(rateLimitStatistics, 1)

	c.common.client = c
	c.Campaigns = (*CampaignsService)(&c.common)
	c.Category = (*CategoryService)(&c.common)
	c.File = (*FileService)(&c.common)
	c.Marketplace = (*MarketplaceService)(&c.common)

	c.Statistics = (*StatisticsService)(&c.common)
}

func (c *Client) bareDo(ctx context.Context, caller *http.Client, req *http.Request) (*http.Response, error) {
	if ctx == nil {
		return nil, errNonNilContext
	}

	req = req.WithContext(ctx)

	cat := GetRateLimitCategory(req.URL.Path)

	err := c.RateLimiter[cat].Wait(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := caller.Do(req)

	if err != nil {
		select {
		case <-ctx.Done():
			return resp, ctx.Err()
		default:
		}

		return resp, err
	}

	err = CheckResponse(resp)

	return resp, err
}

func (c *Client) BareDo(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.bareDo(ctx, c.client, req)
}

func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.BareDo(ctx, req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	switch v := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(v, resp.Body)
	default:
		decErr := json.NewDecoder(resp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil
		}
		if decErr != nil {
			err = decErr
		}
	}

	return resp, err
}

func (c *Client) NewRequest(method, urlStr string, body interface{}, opts ...string) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}

	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	return req, nil
}

func (c *Client) Login() error {
	url, err := c.BaseURL.Parse("login")
	if err != nil {
		return err
	}

	input := map[string]any{
		"api_token": c.apiToken,
	}

	body, err := json.Marshal(input)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bodyBytes, &c.authToken)
	if err != nil {
		return err
	}

	c.authToken.TokenExpiryDate = time.Now().Add(time.Duration(c.authToken.TokenExpiry) * time.Second)

	return nil
}

type ErrorResponse struct {
	Response *http.Response `json:"-"`
	Message  string         `json:"message"`
}

func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := io.ReadAll(r.Body)
	if err == nil && data != nil {
		err = json.Unmarshal(data, errorResponse)
		if err != nil {

			return errors.New("failed to parse error response")
		}
	}

	r.Body = io.NopCloser(bytes.NewBuffer(data))

	return errorResponse
}

func (r *ErrorResponse) Error() string {
	if r.Response != nil && r.Response.Request != nil {
		return fmt.Sprintf("%v %v: %d %v",
			r.Response.Request.Method, r.Response.Request.URL,
			r.Response.StatusCode, r.Message)
	}

	if r.Response != nil {
		return fmt.Sprintf("%d %v", r.Response.StatusCode, r.Message)
	}

	return fmt.Sprintf("%v", r.Message)
}

func GetRateLimitCategory(path string) string {
	if strings.HasPrefix(path, "/statistics/") {
		return rateStatistics
	}

	return rateGlobal
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

type ListOptions struct {
	Limit  int `url:"limit,omitempty" json:"limit,omitempty"`
	Offset int `url:"offset,omitempty" json:"offset,omitempty"`
}

func addOptions(s string, opts interface{}) (string, error) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opts)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

type CustomDate struct {
	time.Time
}

func (d CustomDate) String() string {
	return d.Time.Format("2006-01-02")
}

func (d *CustomDate) UnmarshalJSON(b []byte) error {
	time, err := time.Parse("\"2006-01-02\"", string(b))
	if err != nil {
		return err
	}

	*d = CustomDate{time}

	return nil
}

func (d CustomDate) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.Format("2006-01-02"))), nil
}

func (d CustomDate) Value() (driver.Value, error) {
	return d.Time, nil
}

func (d *CustomDate) Scan(value interface{}) error {
	if value == nil {
		*d = CustomDate{}
		return nil
	}

	dateStr, ok := value.(time.Time)
	if !ok {
		return errors.New("failed to scan Date: expected string")
	}

	*d = CustomDate{Time: dateStr}
	return nil
}

type TimeZone struct {
	*time.Location
}

func (tz *TimeZone) MarshalJSON() ([]byte, error) {
	if tz == nil || tz.Location == nil {
		return []byte("null"), nil
	}
	return json.Marshal(tz.String())
}
