package wafris_traefik

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Config the plugin configuration.
type Config struct {
	URL           string  `json:"url,omitempty"`
	WafrisTimeout float64 `json:"wafris_timeout,omitempty"` //in seconds
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		URL: "redis://localhost:6379?protocol=3",
	}
}

// the main plugin struct.
type WafrisTraefikPlugin struct {
	next http.Handler
	rc   *RedisClient
	sha  string
}

// New created a new WafrisTraefikPlugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.URL) == 0 {
		return nil, fmt.Errorf("2135914077 [Wafris] url cannot be empty")
	}
	timeout_duration := 750 * time.Millisecond

	// log.Printf("2175505556 %+v", config)

	if config.WafrisTimeout > 0 {
		timeout_duration = time.Duration(config.WafrisTimeout * float64(time.Second))
	}

	rc, err := newRedisClient(config.URL, timeout_duration)
	if err != nil {
		derr := fmt.Errorf("91457411044 [Wafris] newRedisClient err: %v", err)
		return nil, derr
	}

	redis_conn, err := rc.newRedisConnection()
	if err != nil {
		derr := fmt.Errorf("93339313776 [Wafris] newRedisConnection failure: %+v", err)
		return nil, derr
	}

	defer redis_conn.Close()

	sha, err := rc.scriptLoad(redis_conn, wafris_core_lua)
	if err != nil {
		derr := fmt.Errorf("4135098426 [Wafris] SCRIPT LOAD failure: %+v", err)
		log.Println(derr)
	}

	// log.Printf("4135098427 [Wafris] INFO Sucessfully loaded lua script %+v", sha)

	return &WafrisTraefikPlugin{
		next: next,
		rc:   rc,
		sha:  sha,
	}, nil
}

func (wtp *WafrisTraefikPlugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// we only set a header for debugging purposes
	req.Header.Set("X-WafrisPlugin", "Passthru")

	ip := getRealIp(req)

	parsed_ip := net.ParseIP(ip)

	args := []string{
		//ip
		ip,
		// ip integer string
		Ip2IntString(parsed_ip),
		// time
		strconv.FormatInt(time.Now().UnixMilli(), 10),
		// request user agent
		req.UserAgent(),
		// request path
		req.URL.RawPath,
		// request query string
		req.URL.RawQuery,
		// request host
		req.Host,
		// request method
		req.Method,
	}

	redis_conn, err := wtp.rc.newRedisConnection()
	if err != nil {
		derr := fmt.Errorf("1457411044 [Wafris] net dial err: %v", err)
		log.Println(derr)
	}
	defer redis_conn.Close()

	sha := wtp.sha
	resp, err := wtp.rc.evalSha(redis_conn, sha, []string{}, args)

	if err != nil {
		derr := fmt.Errorf("92855330633 failure: %+v", err)
		log.Println(derr)
	} else {
		// log.Printf("1226800038 [Wafris] Sucessful evalsha script %+v", sha)
		// log.Printf("1226800039 [Wafris] Sucessful evalsha script %+v", resp)

		if resp == "Blocked" {
			log.Println("2548097413 [Wafris] Blocked:", ip, req.Method, req.Host, req.URL.String())
			writeBlockedResponse(rw)
			return
		}
	}

	// passthru anything else for now
	wtp.next.ServeHTTP(rw, req)
}

// best effort based on x-forwarded-for and RemoteAddr
func getRealIp(req *http.Request) string {
	// var err error
	xff_values := req.Header.Values("x-forwarded-for")

	if len(xff_values) != 0 {
		// reverse the slice
		for i, j := 0, len(xff_values)-1; i < j; i, j = i+1, j-1 {
			xff_values[i], xff_values[j] = xff_values[j], xff_values[i]
		}

		// log.Debugf("req.xff_values: %v", xff_values)

		for _, ip := range xff_values {
			// log.Debugf("3450939168 req.IsTrustedProxy: %v, %v", ip, IsTrustedProxy(ip))
			if !IsTrustedProxy(ip) {
				// ues this one
				return ip
			}
		}
	}

	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		log.Printf("3529707130 [Wafris] error: req.RemoteAddr: %q is not IP:port", req.RemoteAddr)
		ip = req.RemoteAddr
	}
	return ip
}

// https://andrew.red/posts/golang-ipv4-ipv6-to-decimal-integer
func Ip2IntString(ip net.IP) string {
	if ip == nil {
		return "0"
	}

	big_int := big.NewInt(0)
	big_int.SetBytes(ip)
	return big_int.String()
}

func writeBlockedResponse(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusForbidden)
	io.WriteString(w, "Blocked")
	return nil
}
