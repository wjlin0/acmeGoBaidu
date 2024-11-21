package dns01

import (
	"fmt"
	"time"

	"github.com/miekg/dns"
)

const (
	// DefaultPropagationTimeout default propagation timeout.
	DefaultPropagationTimeout = 60 * time.Second

	// DefaultPollingInterval default polling interval.
	DefaultPollingInterval = 2 * time.Second

	// DefaultTTL default TTL.
	DefaultTTL = 120
)

// ChallengeInfo contains the information use to create the TXT record.
type ChallengeInfo struct {
	// FQDN is the full-qualified challenge domain (i.e. `_acme-challenge.[domain].`)
	FQDN string

	// EffectiveFQDN contains the resulting FQDN after the CNAMEs resolutions.
	EffectiveFQDN string

	// Value contains the value for the TXT record.
	Value string
}

func getChallengeBaiduFQDN(domain string) string {
	return fmt.Sprintf("%s.", domain)
}

func GetChallengeBaiduInfo(domain string) ChallengeInfo {
	return ChallengeInfo{
		Value: fmt.Sprintf("%s.a.bdydns.com.", domain),
		FQDN:  getChallengeBaiduFQDN(domain),
	}
}

func CheckCNAMExistBaidu(domain string) bool {
	r, err := dnsQuery(domain, dns.TypeCNAME, recursiveNameservers, true)
	if err != nil {
		return false
	}
	cname := updateDomainWithCName(r, domain)
	if cname == fmt.Sprintf("%s.a.bdydns.com.", domain) {
		return false
	}
	return true
}
