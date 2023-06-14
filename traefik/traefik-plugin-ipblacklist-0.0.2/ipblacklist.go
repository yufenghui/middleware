package traefik_plugin_ipblacklist

import (
	"context"
	"errors"
	"fmt"
	"github.com/LyuHe-uestc/traefik-plugin-ipblacklist/ip"
	"net/http"
)

// IPBlackList holds the ip Black list configuration.
type IPBlackList struct {
	SourceRange []string    `json:"sourceRange,omitempty" toml:"sourceRange,omitempty" yaml:"sourceRange,omitempty"`
	IPStrategy  *IPStrategy `json:"ipStrategy,omitempty" toml:"ipStrategy,omitempty" yaml:"ipStrategy,omitempty"  label:"allowEmpty" file:"allowEmpty" export:"true"`
}

type IPStrategy struct {
	Depth       int      `json:"depth,omitempty" toml:"depth,omitempty" yaml:"depth,omitempty" export:"true"`
	ExcludedIPs []string `json:"excludedIPs,omitempty" toml:"excludedIPs,omitempty" yaml:"excludedIPs,omitempty"`
	// TODO(mpl): I think we should make RemoteAddr an explicit field. For one thing, it would yield better documentation.
}

// Get an IP selection strategy.
// If nil return the RemoteAddr strategy
// else return a strategy based on the configuration using the X-Forwarded-For Header.
// Depth override the ExcludedIPs.
func (s *IPStrategy) Get() (ip.Strategy, error) {
	if s == nil {
		return &ip.RemoteAddrStrategy{}, nil
	}

	if s.Depth > 0 {
		return &ip.DepthStrategy{
			Depth: s.Depth,
		}, nil
	}

	if len(s.ExcludedIPs) > 0 {
		checker, err := ip.NewChecker(s.ExcludedIPs)
		if err != nil {
			return nil, err
		}
		return &ip.PoolStrategy{
			Checker: checker,
		}, nil
	}

	return &ip.RemoteAddrStrategy{}, nil
}

func CreateConfig() *IPBlackList {
	return &IPBlackList{
		SourceRange: nil,
		IPStrategy: &IPStrategy{
			Depth:       0,
			ExcludedIPs: nil,
		},
	}
}

type ipBlackLister struct {
	next        http.Handler
	blackLister *ip.Checker
	strategy    ip.Strategy
	name        string
}

func New(ctx context.Context, next http.Handler, config *IPBlackList, name string) (http.Handler, error) {

	if len(config.SourceRange) == 0 {
		return nil, errors.New("sourceRange is empty, IPBlackLister not created")
	}

	checker, err := ip.NewChecker(config.SourceRange)
	if err != nil {
		return nil, fmt.Errorf("cannot parse CIDR Blacklist %s: %w", config.SourceRange, err)
	}

	strategy, err := config.IPStrategy.Get()
	if err != nil {
		return nil, err
	}

	return &ipBlackLister{
		strategy:    strategy,
		blackLister: checker,
		next:        next,
		name:        name,
	}, nil
}

func (wl *ipBlackLister) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	err := wl.blackLister.IsAuthorized(wl.strategy.GetIP(req))
	if err == nil {
		reject(rw)
		return
	}

	wl.next.ServeHTTP(rw, req)
}

func reject(rw http.ResponseWriter) {
	statusCode := http.StatusForbidden

	rw.WriteHeader(statusCode)
}
