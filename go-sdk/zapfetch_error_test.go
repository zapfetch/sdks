package zapfetch

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	cases := []struct {
		name string
		err  *APIError
		want string
	}{
		{
			name: "with code and status",
			err:  &APIError{StatusCode: 500, ErrorCode: "INTERNAL", Message: "boom"},
			want: "zapfetch: HTTP 500 [INTERNAL]: boom",
		},
		{
			name: "with status only",
			err:  &APIError{StatusCode: 404, Message: "not found"},
			want: "zapfetch: HTTP 404: not found",
		},
		{
			name: "no status",
			err:  &APIError{Message: "bad config"},
			want: "zapfetch: bad config",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.want {
				t.Fatalf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestJobTimeoutError_Error(t *testing.T) {
	err := &JobTimeoutError{
		APIError:       APIError{Message: "crawl job timed out"},
		JobID:          "job-42",
		TimeoutSeconds: 120,
	}
	want := "zapfetch: job job-42 timed out after 120 seconds"
	if got := err.Error(); got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

// AuthenticationError and RateLimitError embed APIError; `errors.As` should
// unwrap to the underlying APIError so callers can switch on it.
func TestAuthenticationError_UnwrapsToAPIError(t *testing.T) {
	var err error = &AuthenticationError{APIError: APIError{StatusCode: 401, Message: "bad key"}}

	var authErr *AuthenticationError
	if !errors.As(err, &authErr) {
		t.Fatal("expected errors.As to match *AuthenticationError")
	}
	if authErr.StatusCode != 401 {
		t.Fatalf("StatusCode = %d, want 401", authErr.StatusCode)
	}
}

func TestRateLimitError_Matches(t *testing.T) {
	var err error = &RateLimitError{APIError: APIError{StatusCode: 429, Message: "slow down"}}
	var rateErr *RateLimitError
	if !errors.As(err, &rateErr) {
		t.Fatal("expected errors.As to match *RateLimitError")
	}
}
