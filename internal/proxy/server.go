package proxy

import (
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/chramiq/workaround/internal/config"
	"github.com/chramiq/workaround/internal/ui"
)

type Server struct {
	workers         []config.WorkerInfo
	userAgents      []string
	randomizeUA     bool
	counter         uint64
	listener        net.Listener
	httpClient      *http.Client
	TargetScheme    string
	ForceNewCircuit bool
}

func NewServer(workers []config.WorkerInfo, targetScheme string, userAgents []string, randomizeUA bool, upstreamProxy string, forceNewCircuit bool) *Server {
	if targetScheme == "" {
		targetScheme = "https"
	}

	transport := &http.Transport{
		Proxy: nil,
	}

	if upstreamProxy != "" {
		baseProxyURL, err := url.Parse(upstreamProxy)
		if err != nil {
			ui.Error("Invalid upstream proxy URL: %v. Ignoring.", err)
		} else {
			transport.Proxy = func(req *http.Request) (*url.URL, error) {
				if !forceNewCircuit {
					return baseProxyURL, nil
				}

				if baseProxyURL.Scheme == "socks5" {
					u := *baseProxyURL
					id := fmt.Sprintf("%x", rand.Uint64())
					u.User = url.UserPassword("wa-"+id, id)
					return &u, nil
				}
				return baseProxyURL, nil
			}
		}
	}

	if forceNewCircuit {
		transport.DisableKeepAlives = true
	}

	return &Server{
		workers:         workers,
		TargetScheme:    targetScheme,
		userAgents:      userAgents,
		randomizeUA:     randomizeUA,
		ForceNewCircuit: forceNewCircuit,
		httpClient: &http.Client{
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: 30 * time.Second,
		},
	}
}

func (s *Server) Start(addr string) (string, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", err
	}
	s.listener = ln
	go http.Serve(ln, s)
	return ln.Addr().String(), nil
}

func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		ui.Debug("Rejected CONNECT request to %s", r.Host)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Workaround Error: HTTPS Tunneling (CONNECT) is not supported.\n")
		return
	}

	worker := s.nextWorker()

	targetURL := r.URL
	targetURL.Scheme = s.TargetScheme
	finalTarget := targetURL.String()

	req, err := http.NewRequest(r.Method, worker.URL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", 500)
		return
	}

	for k, vv := range r.Header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	if s.randomizeUA && len(s.userAgents) > 0 {
		randIndex := rand.N(len(s.userAgents))
		req.Header.Set("User-Agent", s.userAgents[randIndex])
	}

	req.Header.Set("X-Target-URL", finalTarget)
	req.Host = req.URL.Host

	resp, err := s.httpClient.Do(req)
	if err != nil {
		ui.Debug("Worker request failed: %v", err)
		http.Error(w, fmt.Sprintf("Proxy Error: %v", err), 502)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (s *Server) nextWorker() config.WorkerInfo {
	idx := atomic.AddUint64(&s.counter, 1)
	return s.workers[(idx-1)%uint64(len(s.workers))]
}
