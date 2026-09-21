package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sdkconfig "github.com/stackitcloud/stackit-sdk-go/core/config"
)

var _ = Describe("WrapError", func() {
	It("wraps the error with the provided identifier", func() {
		err := errors.New("test error")
		expected := fmt.Errorf("[X-Trace-Id:12345]: %w", err)
		Expect(WrapError(err, "X-Trace-Id", "12345")).To(Equal(expected))
	})

	It("returns the original error when the identifier is empty", func() {
		err := errors.New("test error")
		Expect(WrapError(err, "trace-id", "")).To(Equal(err))
	})

	It("returns nil when the error is nil", func() {
		Expect(WrapError(nil, "trace-id", "12345")).To(Succeed())
	})
})

var _ = Describe("execute", func() {
	It("wraps API errors with trace and request IDs", func() {
		_, err := execute(context.Background(), func(ctx context.Context) (int, error) {
			response, ok := ctx.Value(sdkconfig.ContextHTTPResponse).(**http.Response)
			Expect(ok).To(BeTrue())
			*response = &http.Response{Header: http.Header{
				"X-Trace-Id":   {"trace-123"},
				"X-Request-Id": {"request-456"},
			}}
			return 0, errors.New("api error")
		})

		Expect(err).To(MatchError("[X-Request-Id:request-456]: [X-Trace-Id:trace-123]: api error"))
	})

	It("returns the result and nil error on success", func() {
		res, err := execute(context.Background(), func(_ context.Context) (string, error) {
			return "success", nil
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal("success"))
	})

	It("wraps only trace ID if request ID is missing", func() {
		_, err := execute(context.Background(), func(ctx context.Context) (int, error) {
			response, ok := ctx.Value(sdkconfig.ContextHTTPResponse).(**http.Response)
			Expect(ok).To(BeTrue())
			*response = &http.Response{Header: http.Header{
				"X-Trace-Id": {"trace-123"},
			}}
			return 0, errors.New("api error")
		})

		Expect(err).To(MatchError("[X-Trace-Id:trace-123]: api error"))
	})

	It("wraps only request ID if trace ID is missing", func() {
		_, err := execute(context.Background(), func(ctx context.Context) (int, error) {
			response, ok := ctx.Value(sdkconfig.ContextHTTPResponse).(**http.Response)
			Expect(ok).To(BeTrue())
			*response = &http.Response{Header: http.Header{
				"X-Request-Id": {"request-456"},
			}}
			return 0, errors.New("api error")
		})

		Expect(err).To(MatchError("[X-Request-Id:request-456]: api error"))
	})

	It("returns the original error when neither trace ID nor response is present", func() {
		_, err := execute(context.Background(), func(_ context.Context) (int, error) {
			return 0, errors.New("api error")
		})

		Expect(err).To(MatchError("api error"))
	})
})
