package temu

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/hiscaler/temu-go/normal"
)

// InvokeResponse 通用 API 调用响应（用于调试/测试）
type InvokeResponse struct {
	StatusCode int             `json:"statusCode"`
	DurationMs int64           `json:"durationMs"`
	Body       json.RawMessage `json:"body"`
}

// Invoke 通用 API 调用
func (c *Client) Invoke(ctx context.Context, typ string, body map[string]any) (*InvokeResponse, error) {
	if body == nil {
		body = map[string]any{}
	}

	var result struct {
		normal.Response
		Result any `json:"result"`
	}
	resp, err := c.Services.Mall.httpClient.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(&result).
		Post(typ)

	invokeResp := &InvokeResponse{}
	if resp != nil {
		invokeResp.StatusCode = resp.StatusCode()
		invokeResp.DurationMs = resp.Time().Milliseconds()
		invokeResp.Body = resp.Body()
	}

	if err = recheckError(resp, result.Response, err); err != nil {
		return invokeResp, err
	}
	return invokeResp, nil
}
