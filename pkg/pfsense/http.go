package pfsense

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

func shouldRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if err != nil {
		return true, nil //lint:ignore nilerr httpDoErr handled elsewhere
	}

	if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented) {
		return true, fmt.Errorf("%w %s", ErrHTTPStatus, resp.Status)
	}

	return false, nil
}

func linearJitter(minJitter, maxJitter time.Duration, attempt int) *time.Timer {
	rand := rand.New(rand.NewSource(int64(time.Now().Nanosecond()))) // #nosec G404
	jitter := int64(rand.Float64()*float64(maxJitter-minJitter)) + int64(minJitter)
	duration := time.Duration(jitter * int64(attempt))

	return time.NewTimer(duration)
}

func (pf *Client) retryableDo(req *http.Request, reqBody *[]byte) (*http.Response, error) {
	var resp *http.Response
	var attempt int
	var retry bool
	var httpDoErr, shouldRetryErr error

	for attempt = 1; ; attempt++ {
		if reqBody != nil {
			req.Body = io.NopCloser(bytes.NewReader(*reqBody))
		}

		resp, httpDoErr = pf.httpClient.Do(req)
		retry, shouldRetryErr = shouldRetry(req.Context(), resp, httpDoErr)

		if !retry || (*pf.Options.MaxAttempts-attempt) <= 0 {
			break
		}

		if httpDoErr == nil {
			resp.Body.Close()
			_, _ = io.Copy(io.Discard, resp.Body)
		}

		timer := linearJitter(*pf.Options.RetryMinWait, *pf.Options.RetryMaxWait, attempt)
		select {
		case <-req.Context().Done():
			timer.Stop()

			return nil, req.Context().Err()
		case <-timer.C:
		}

		httpReq := *req
		req = &httpReq
	}

	if httpDoErr == nil && shouldRetryErr == nil && !retry {
		return resp, nil
	}

	if resp != nil {
		resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
	}

	if httpDoErr != nil {
		return nil, fmt.Errorf("%w after %d attempt(s), %s %s, %w", ErrFailedRequest, attempt, req.Method, req.URL.Path, httpDoErr)
	}

	if shouldRetryErr != nil {
		return nil, fmt.Errorf("%w after %d attempt(s), %s %s, %w", ErrFailedRequest, attempt, req.Method, req.URL.Path, shouldRetryErr)
	}

	return nil, fmt.Errorf("%w after %d attempt(s), %s %s", ErrFailedRequest, attempt, req.Method, req.URL.Path)
}

func (pf *Client) call(ctx context.Context, method string, relativeURL url.URL, values *url.Values) (*http.Response, error) {
	var reqBody *[]byte
	var reqBodyContentLength int64
	if values != nil {
		if pf.tokenKey != "" && pf.token != "" {
			values.Set(pf.tokenKey, pf.token)
		}
		reqBytes := []byte(values.Encode())
		reqBody = &reqBytes
		reqBodyContentLength = int64(len(reqBytes))
	}

	url := pf.Options.URL.ResolveReference(&relativeURL).String()
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request, %s %s %w", method, relativeURL.Path, err)
	}

	req.ContentLength = reqBodyContentLength
	req.Header.Set("User-Agent", "go-pfsense")
	if values != nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := pf.retryableDo(req, reqBody)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w, %s %s %s", ErrFailedRequest, resp.Status, req.Method, req.URL.Path)
	}

	return resp, nil
}
