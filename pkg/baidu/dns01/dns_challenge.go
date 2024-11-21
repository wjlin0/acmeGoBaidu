package dns01

import (
	"github.com/miekg/dns"
	"time"
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

	// Value contains the value for the TXT record.
	Value string
}

func GetChallengeInfo(domain string, value string) ChallengeInfo {
	return ChallengeInfo{
		Value: value,
		FQDN:  ToFqdn(domain),
	}
}

func CheckCNAMExistBaidu(fqdn string) (uint16, string, error) {
	// 首先判断是否存在 CNAME
	r, err := dnsQuery(fqdn, dns.TypeCNAME, recursiveNameservers, true)
	if err != nil {
		return 0, "", err
	}
	cname := updateDomainWithCName(r, fqdn)
	if cname != fqdn {
		return dns.TypeCNAME, cname, nil
	}
	// 判断是否存在 A 记录
	r, err = dnsQuery(fqdn, dns.TypeA, recursiveNameservers, true)
	if err != nil {
		return 0, "", err
	}

	ip := updateDomainWithIp(r, fqdn)
	if ip != "" {
		return dns.TypeA, ip, nil
	}

	return 0, "", nil
}
