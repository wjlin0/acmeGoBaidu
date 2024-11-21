package baidu

import (
	"fmt"
	"github.com/wjlin0/acmeGoBaidu/pkg/baidu/cloudflare"
)

// NewDNSChallengeProviderByName Factory for DNS providers.
func NewDNSChallengeProviderByName(name string) (Provider, error) {
	switch name {
	case "cloudflare":
		return cloudflare.NewDNSProvider()
	default:
		return nil, fmt.Errorf("unrecognized DNS provider: %s", name)
	}
}
