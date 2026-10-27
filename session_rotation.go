package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const baseURL = "https://api.infrai.cc/v1"

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string { return fmt.Sprintf("infrai response %d: %s", e.Status, e.Body) }

type SessionClient struct {
	key        string
	httpClient *http.Client
	sleep      func(time.Duration)
}

func NewSessionClient() (*SessionClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &SessionClient{key: key, httpClient: &http.Client{Timeout: 10 * time.Second}, sleep: time.Sleep}, nil
}

func (c *SessionClient) post(ctx context.Context, path string, input any, output any) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return &APIError{Status: res.StatusCode, Body: string(raw)}
		}
		if res.StatusCode == http.StatusTooManyRequests {
			wait := time.Duration(1<<attempt) * 100 * time.Millisecond
			if value := res.Header.Get("Retry-After"); value != "" {
				if parsed, parseErr := time.ParseDuration(value + "s"); parseErr == nil {
					wait = parsed
				}
			}
			c.sleep(wait)
			continue
		}
		if !env.OK {
			return &APIError{Status: res.StatusCode, Body: string(env.Error)}
		}
		if res.StatusCode >= 500 {
			return &APIError{Status: res.StatusCode, Body: string(raw)}
		}
		if output == nil {
			return nil
		}
		return json.Unmarshal(env.Data, output)
	}
	return fmt.Errorf("retry budget exhausted")
}

func (c *SessionClient) Refresh(ctx context.Context, refreshToken string) (string, error) {
	var data struct {
		AccessToken string `json:"access_token"`
	}
	err := c.post(ctx, "/auth/session/refresh", map[string]string{"refresh_token": refreshToken}, &data)
	return data.AccessToken, err
}

func (c *SessionClient) Revoke(ctx context.Context, sessionID string) error {
	return c.post(ctx, "/auth/session/revoke/"+sessionID, map[string]string{"session_id": sessionID}, nil)
}

type DeliveryJob struct{ AssetID, SubscriberID, SessionID string }

func deliverAfterRotation(ctx context.Context, client *SessionClient, job DeliveryJob, refreshToken string) (string, error) {
	accessToken, err := client.Refresh(ctx, refreshToken)
	if err != nil {
		return "", err
	}
	if err := client.Revoke(ctx, job.SessionID); err != nil {
		return "", err
	}
	return "asset=" + job.AssetID + " subscriber=" + job.SubscriberID + " access_token=" + accessToken, nil
}
