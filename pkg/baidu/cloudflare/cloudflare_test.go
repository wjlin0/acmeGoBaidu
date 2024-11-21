package cloudflare

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewDNSProviderConfig(t *testing.T) {
	testCases := []struct {
		desc      string
		authEmail string
		authKey   string
		authToken string
		expected  string
	}{
		{
			desc:      "success with email and api key",
			authEmail: "test@example.com",
			authKey:   "123",
		},
		{
			desc:      "success with api token",
			authToken: "012345abcdef",
		},
		{
			desc:      "prefer api token",
			authToken: "012345abcdef",
			authEmail: "test@example.com",
			authKey:   "123",
		},
		{
			desc:     "missing credentials",
			expected: "cloudflare: invalid credentials: key & email must not be empty",
		},
		{
			desc:     "missing email",
			authKey:  "123",
			expected: "cloudflare: invalid credentials: key & email must not be empty",
		},
		{
			desc:      "missing api key",
			authEmail: "test@example.com",
			expected:  "cloudflare: invalid credentials: key & email must not be empty",
		},
		{
			desc:      "missing api token, fallback to api key/email",
			authToken: "",
			expected:  "cloudflare: invalid credentials: key & email must not be empty",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			config := NewDefaultConfig()
			config.AuthEmail = test.authEmail
			config.AuthKey = test.authKey
			config.AuthToken = test.authToken

			p, err := NewDNSProviderConfig(config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				assert.NotNil(t, p.config)
				assert.NotNil(t, p.client)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}
