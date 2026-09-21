package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/stackitcloud/stackit-sdk-go/core/runtime"
	sdkWait "github.com/stackitcloud/stackit-sdk-go/services/iaas/v2api/wait"
)

// WrapError wraps the error with an identifier but only if the error is not nil.
func WrapError(err error, name, id string) error {
	if err == nil {
		return nil
	}
	if id == "" {
		return err
	}
	return fmt.Errorf("[%s:%s]: %w", name, id, err)
}

func execute[T any](ctx context.Context, call func(context.Context) (T, error)) (T, error) {
	var httpResp *http.Response
	ctx = runtime.WithCaptureHTTPResponse(ctx, &httpResp)

	resp, err := call(ctx)
	if err != nil {
		var zero T
		err = WrapError(err, "X-Trace-Id", runtime.GetTraceId(ctx))
		if httpResp != nil {
			reqID := httpResp.Header.Get(sdkWait.XRequestIDHeader)
			err = WrapError(err, sdkWait.XRequestIDHeader, reqID)
		}
		return zero, err
	}
	return resp, nil
}

// convertLabelsToSDK converts map[string]string to *map[string]any for SDK
func convertLabelsToSDK(labels map[string]string) map[string]any {
	if labels == nil {
		return nil
	}

	result := make(map[string]any, len(labels))
	for k, v := range labels {
		result[k] = v
	}
	return result
}

// convertLabelsFromSDK converts map[string]any from SDK to map[string]string
func convertLabelsFromSDK(labels map[string]any) map[string]string {
	if labels == nil {
		return nil
	}

	result := make(map[string]string, len(labels))
	for k, v := range labels {
		if strVal, ok := v.(string); ok {
			result[k] = strVal
		}
	}
	return result
}
