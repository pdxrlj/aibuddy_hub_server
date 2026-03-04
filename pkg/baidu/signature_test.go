package baidu

import (
	"net/url"
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		input         string
		expected      string
		encodingSlash bool
	}{
		{"abc", "abc", true},
		{"test/value", "test%2Fvalue", true},
		{"test/value", "test/value", false},
		{"测试", "%E6%B5%8B%E8%AF%95", true},
		{"hello world", "hello%20world", true},
		{"test!value", "test!value", true},
	}

	for _, test := range tests {
		result := normalize(test.input, test.encodingSlash)
		if result != test.expected {
			t.Errorf("normalize(%s, %v) = %s; expected %s", test.input, test.encodingSlash, result, test.expected)
		}
	}
}

func TestGenerateAuth(t *testing.T) {
	signer := NewSignature("test-ak", "test-sk")

	query := url.Values{}
	query.Set("param", "value")

	headers := map[string]string{
		"host": "bj.bcebos.com",
	}

	result := signer.GenerateAuth("GET", "/bucket/object", query, headers)

	if result.Authorization == "" {
		t.Error("Authorization should not be empty")
	}

	if !strings.HasPrefix(result.Authorization, "bce-auth-v1/") {
		t.Errorf("Authorization should start with 'bce-auth-v1/', got %s", result.Authorization[:12])
	}

	if result.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}

	if result.SignedHeaders == "" {
		t.Error("SignedHeaders should not be empty")
	}

	t.Logf("Authorization: %s", result.Authorization)
}
