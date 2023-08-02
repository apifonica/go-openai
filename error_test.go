package openai_test

import (
	"errors"
	"net/http"
	"testing"
)

func TestAPIErrorUnmarshalJSON(t *testing.T) {
	type testCase struct {
		name      string
		response  string
		hasError  bool
		checkFunc func(t *testing.T, apiErr APIError)
	}
	testCases := []testCase{
		// testcase for message field
		{
			name:     "parse succeeds when the message is string",
			response: `{"message":"foo","type":"invalid_request_error","param":null,"code":null}`,
			hasError: false,
			checkFunc: func(t *testing.T, apiErr APIError) {
				assertAPIErrorMessage(t, apiErr, "foo")
			},
		},
		{
			name:     "parse succeeds when the message is array with single item",
			response: `{"message":["foo"],"type":"invalid_request_error","param":null,"code":null}`,
			hasError: false,
			checkFunc: func(t *testing.T, apiErr APIError) {
				assertAPIErrorMessage(t, apiErr, "foo")
			},
		},
		{
			name:     "parse succeeds when the message is array with multiple items",
			response: `{"message":["foo", "bar", "baz"],"type":"invalid_request_error","param":null,"code":null}`,
			hasError: false,
			checkFunc: func(t *testing.T, apiErr APIError) {
				assertAPIErrorMessage(t, apiErr, "foo, bar, baz")
			},
		},
		{
			name:     "parse succeeds when the message is empty array",
			response: `{"message":[],"type":"invalid_request_error","param":null,"code":null}`,
			hasError: false,
			checkFunc: func(t *testing.T, apiErr APIError) {
				assertAPIErrorMessage(t, apiErr, "")
			},
		},
		{
			name:     "parse succeeds when the message is null",
			response: `{"message":null,"type":"invalid_request_error","param":null,"code":null}`,
			hasError: false,
			checkFunc: func(t *testing.T, apiErr APIError) {
				assertAPIErrorMessage(t, apiErr, "")
			},
		},
		{
			name:     "parse failed when the message is object",
			response: `{"message":{},"type":"invalid_request_error","param":null,"code":null}`,
			hasError: true,
		},
		{
			name:     "parse failed when the message is int",
			response: `{"message":1,"type":"invalid_request_error","param":null,"code":null}`,
			hasError: true,
		},
		{
			name:     "parse failed when the message is float",
			response: `{"message":0.1,"type":"invalid_request_error","param":null,"code":null}`,
			hasError: true,
		},
		{
			name:     "parse failed when the message is bool",
			response: `{"message":true,"type":"invalid_request_error","param":null,"code":null}`,
			hasError: true,
		},
		{
			name:     "parse failed when the message is not exists",
			response: `{"type":"invalid_request_error","param":null,"code":null}`,
			hasError: true,
		},
		// testcase for code field
		{
			name:     "parse succeeds when the code is int",
			response: `{"code":418,"message":"I'm a teapot","param":"prompt","type":"teapot_error"}`,
			hasError: false,
			checkFunc: func(t *testing.T, apiErr APIError) {
				assertAPIErrorCode(t, apiErr, 418)
			},
		},
		{
			name:     "parse succeeds when the code is string",
			response: `{"code":"teapot","message":"I'm a teapot","param":"prompt","type":"teapot_error"}`,
			hasError: false,
			checkFunc: func(t *testing.T, apiErr APIError) {
				assertAPIErrorCode(t, apiErr, "teapot")
			},
		},
		{
			name:     "parse succeeds when the code is not exists",
			response: `{"message":"I'm a teapot","param":"prompt","type":"teapot_error"}`,
			hasError: false,
			checkFunc: func(t *testing.T, apiErr APIError) {
				assertAPIErrorCode(t, apiErr, nil)
			},
		},
		// testcase for param field
		{
			name:     "parse failed when the param is bool",
			response: `{"code":418,"message":"I'm a teapot","param":true,"type":"teapot_error"}`,
			hasError: true,
		},
		// testcase for type field
		{
			name:     "parse failed when the type is bool",
			response: `{"code":418,"message":"I'm a teapot","param":"prompt","type":true}`,
			hasError: true,
		},
		// testcase for error response
		{
			name:     "parse failed when the response is invalid json",
			response: `--- {"code":418,"message":"I'm a teapot","param":"prompt","type":"teapot_error"}`,
			hasError: true,
			checkFunc: func(t *testing.T, apiErr APIError) {
				assertAPIErrorCode(t, apiErr, nil)
				assertAPIErrorMessage(t, apiErr, "")
				assertAPIErrorParam(t, apiErr, nil)
				assertAPIErrorType(t, apiErr, "")
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var apiErr APIError
			err := apiErr.UnmarshalJSON([]byte(tc.response))
			if (err != nil) != tc.hasError {
				t.Errorf("Unexpected error: %v", err)
			}
			if tc.checkFunc != nil {
				tc.checkFunc(t, apiErr)
			}
		})
	}
}

func assertAPIErrorMessage(t *testing.T, apiErr APIError, expected string) {
	if apiErr.Message != expected {
		t.Errorf("Unexpected APIError message: %v; expected: %s", apiErr, expected)
	}
}

func assertAPIErrorCode(t *testing.T, apiErr APIError, expected interface{}) {
	switch v := apiErr.Code.(type) {
	case int:
		if v != expected {
			t.Errorf("Unexpected APIError code integer: %d; expected %d", v, expected)
		}
	case string:
		if v != expected {
			t.Errorf("Unexpected APIError code string: %s; expected %s", v, expected)
		}
	case nil:
	default:
		t.Errorf("Unexpected APIError error code type: %T", v)
	}
}

func assertAPIErrorParam(t *testing.T, apiErr APIError, expected *string) {
	if apiErr.Param != expected {
		t.Errorf("Unexpected APIError param: %v; expected: %s", apiErr, *expected)
	}
}

func assertAPIErrorType(t *testing.T, apiErr APIError, typ string) {
	if apiErr.Type != typ {
		t.Errorf("Unexpected API type: %v; expected: %s", apiErr, typ)
	}
}

func TestRequestError(t *testing.T) {
	var err error = &RequestError{
		HTTPStatusCode: http.StatusTeapot,
		Err:            errors.New("i am a teapot"),
	}

	var reqErr *RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("Error is not a RequestError: %+v", err)
	}

	if reqErr.HTTPStatusCode != 418 {
		t.Fatalf("Unexpected request error status code: %d", reqErr.HTTPStatusCode)
	}

	if reqErr.Unwrap() == nil {
		t.Fatalf("Empty request error occurred")
	}
}
