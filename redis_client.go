package wafris_traefik

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type RedisClient struct {
	RedisURI string
	Addr     string
	User     string
	Password string
	Timeout  time.Duration
}

func (rc RedisClient) String() string {
	sb := strings.Builder{}

	redactedURI, _ := url.Parse(rc.RedisURI)
	sb.WriteString("RedisClient:\n")
	sb.WriteString("-- RedisURI: ")
	sb.WriteString(redactedURI.Redacted())
	sb.WriteString("\n-- ")
	sb.WriteString("Addr: ")
	sb.WriteString(rc.Addr)
	sb.WriteString("\n-- ")
	sb.WriteString("User: ")
	sb.WriteString(rc.User)
	sb.WriteString("\n-- ")
	sb.WriteString("Password: ")
	if rc.Password != "" {
		sb.WriteString("REDACTED")
	}
	sb.WriteString("\n-- ")
	sb.WriteString("Timeout: ")
	sb.WriteString(fmt.Sprintf("%v", rc.Timeout))
	sb.WriteString("\n")

	return sb.String()
}

// redis URI example: redis://user:password@host:port/dbnum
func newRedisClient(redisURI string, timeout_duration time.Duration) (*RedisClient, error) {
	// make sure we DO NOT log full url details to prevent leaking of credentials
	parsedURL, err := url.Parse(redisURI)
	if err != nil {
		derr := fmt.Errorf("9149404555 [Wafris] newRedisConnection url.Parse error: %+v", err)
		return nil, derr
	}

	if parsedURL.Scheme != "redis" && parsedURL.Scheme != "rediss" {
		derr := fmt.Errorf("9149404556 [Wafris] newRedisConnection unexpected URI scheme: %+v", parsedURL.Scheme)
		return nil, derr
	}

	if parsedURL.Opaque != "" {
		derr := fmt.Errorf("9149404556 [Wafris] newRedisConnection URI is opaque")
		return nil, derr
	}

	host, port, err := net.SplitHostPort(parsedURL.Host)
	if err != nil {
		// if SplitHostPort fails, most likely, no port is explicitly set and we assume the default redis port
		host = parsedURL.Host
		port = "6379"
	}
	if host == "" {
		host = "localhost"
	}
	addr := net.JoinHostPort(host, port)

	user := parsedURL.User.Username()
	pw, _ := parsedURL.User.Password()

	if timeout_duration <= 0 {
		timeout_duration = time.Second
	}

	return &RedisClient{
		RedisURI: redisURI,
		Addr:     addr,
		User:     user,
		Password: pw,
		Timeout:  timeout_duration,
	}, nil
}

// in the current implementation, functions that call this must defer Close() the connection in order to prevent memory leaks
func (rc *RedisClient) newRedisConnection() (net.Conn, error) {

	redis_conn, err := net.Dial("tcp", rc.Addr)
	if err != nil {
		derr := fmt.Errorf("1457411044 [Wafris] net dial err: %v", err)
		return nil, derr
	}

	// log.Printf("3972581051 [Wafris] DEBUG %+v", rc.String())

	if rc.Password != "" {
		err = rc.auth(redis_conn)
		if err != nil {
			derr := fmt.Errorf("92517536889 [Wafris] auth failure: %+v", err)
			return nil, derr
		}
	}

	err = rc.ping(redis_conn)
	if err != nil {
		derr := fmt.Errorf("1552470915 [Wafris] ping failure: %+v", err)
		return nil, derr
	}

	// log.Printf("1942584125 [Wafris] Sucessfully connected to Redis: %v", rc)

	return redis_conn, nil
}

func (rc *RedisClient) auth(redis_conn net.Conn) error {
	var redis_strings []string

	if rc.User == "" {
		redis_strings = []string{
			"AUTH",
			rc.Password,
		}
	} else {
		redis_strings = []string{
			"AUTH",
			rc.User,
			rc.Password,
		}
	}
	auth_bytes := redisArrayOfBulkStrings(redis_strings)

	redis_conn.SetWriteDeadline(time.Now().Add(rc.Timeout))
	_, err := redis_conn.Write(auth_bytes)
	if err != nil {
		derr := fmt.Errorf("754593925 [Wafris] net write err: %v", err)
		return derr
	}

	redis_conn.SetReadDeadline(time.Now().Add(rc.Timeout))
	bufreader := bufio.NewReader(redis_conn)

	response_bytes, err := bufreader.ReadSlice('\n')
	if err != nil {
		derr := fmt.Errorf("9747736359 [Wafris] read redis response failure: %+v", err)
		log.Println(derr)
		return derr
	} else {
		resp := string(response_bytes)
		// log.Printf("82067161085 [Wafris] DEBUG: parseRedisResponse %v", string(response_bytes))

		if resp == "+OK\r\n" {
			return nil
		} else {
			derr := fmt.Errorf("91081136888 [Wafris] AUTH response: %+v, %+v", resp, err)
			return derr
		}
	}
}

func (rc *RedisClient) ping(redis_conn net.Conn) error {

	redis_conn.SetWriteDeadline(time.Now().Add(rc.Timeout))
	_, err := redis_conn.Write([]byte("PING\r\n"))
	if err != nil {
		derr := fmt.Errorf("1457411045 [Wafris] net write err: %v", err)
		return derr
	}

	redis_conn.SetReadDeadline(time.Now().Add(rc.Timeout))
	bufreader := bufio.NewReader(redis_conn)
	response_bytes, err := bufreader.ReadSlice('\n')
	if err != nil {
		derr := fmt.Errorf("2754780871 [Wafris] read redis response failure: %+v", err)
		log.Println(derr)
		return derr
	} else {

		resp := string(response_bytes)

		if resp == "+PONG\r\n" {
			return nil
		} else {
			derr := fmt.Errorf("91378563400 [Wafris] unexpected response to PING: %+v", err)
			return derr
		}
	}

}
func (rc *RedisClient) scriptLoad(redis_conn net.Conn, lua_script string) (string, error) {
	redis_strings := []string{
		"SCRIPT",
		"LOAD",
		lua_script,
	}
	script_load_bytes := redisArrayOfBulkStrings(redis_strings)

	redis_conn.SetWriteDeadline(time.Now().Add(rc.Timeout))
	_, err := redis_conn.Write(script_load_bytes)
	if err != nil {
		derr := fmt.Errorf("96355322237 [Wafris] net write err: %v", err)
		return "", derr
	}

	redis_conn.SetReadDeadline(time.Now().Add(rc.Timeout))
	bufreader := bufio.NewReader(redis_conn)

	resp, err := parseRedisResponse(bufreader)
	if err != nil {
		derr := fmt.Errorf("93432226744 [Wafris] parseRedisResponse: %+v", err)
		return "", derr
	}

	return string(resp), nil
}

func (rc *RedisClient) evalShaCounter(redis_conn net.Conn, sha string, keys []string, args []string, count int) (string, error) {
	redis_strings := []string{
		"EVALSHA",
		sha,
		strconv.FormatInt(int64(len(keys)), 10),
	}
	redis_strings = append(redis_strings, keys...)
	redis_strings = append(redis_strings, args...)

	evalsha_bytes := redisArrayOfBulkStrings(redis_strings)

	redis_conn.SetWriteDeadline(time.Now().Add(rc.Timeout))
	_, err := redis_conn.Write(evalsha_bytes)
	if err != nil {
		derr := fmt.Errorf("9653235077 [Wafris] net write err: %v", err)
		return "", derr
	}

	redis_conn.SetReadDeadline(time.Now().Add(rc.Timeout))
	bufreader := bufio.NewReader(redis_conn)

	resp, err := parseRedisResponse(bufreader)
	if err != nil {
		// log.Println(9913953570, "[Wafris] DEBUG", string(resp))

		// we have a special exception to error handing here
		if strings.HasPrefix(string(resp), "-NOSCRIPT") {
			_, err := rc.scriptLoad(redis_conn, wafris_core_lua)
			if err != nil {
				derr := fmt.Errorf("9413098426 SCRIPT LOAD failure: %+v", err)
				return "", derr
			}
			// try again
			if count > 2 {
				derr := fmt.Errorf("9413098477 EVALSHA too many retries.v")
				return "", derr
			} else {
				return rc.evalShaCounter(redis_conn, sha, keys, args, count+1)
			}
		} else {

			derr := fmt.Errorf("91091953537 [Wafris] parseRedisResponse: %+v", err)
			return "", derr

		}
	}

	return string(resp), nil
}

func (rc *RedisClient) evalSha(redis_conn net.Conn, sha string, keys []string, args []string) (string, error) {
	return rc.evalShaCounter(redis_conn, sha, keys, args, 0)
}

// format array  of bulk strings accorgind to redis protocol
func redisArrayOfBulkStrings(redis_strings []string) []byte {
	count := len(redis_strings)

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("*%v\r\n", count))

	for _, redis_string := range redis_strings {

		sb.WriteString(fmt.Sprintf("$%d\r\n", len(redis_string)))
		sb.WriteString(redis_string)
		sb.WriteString("\r\n")
	}
	final_string := sb.String()

	// debug
	// log.Printf("8401731020 [Wafris] DEBUG: redisArrayOfBulkStrings final_string %v", final_string)
	return []byte(final_string)
}

// only supports bulk strings ($) and errors (-)
func parseRedisResponse(bufreader *bufio.Reader) ([]byte, error) {
	first_char, err := bufreader.Peek(1)
	if err != nil {
		derr := fmt.Errorf("882366713 [Wafris] redis response parse failure: %+v", err)
		return nil, derr
	}

	switch first_char[0] {
	case '$':
		return parseBulkString(bufreader)
	case '-':
		resp, err := bufreader.ReadSlice('\n')
		derr := fmt.Errorf("91378563411 [Wafris] parseRedisResponse returned error: %+v, %+v", string(resp), err)
		return resp, derr
	default:
		resp, err := bufreader.ReadSlice('\n')
		derr := fmt.Errorf("91378563413 [Wafris] parseRedisResponse unexpected response: %+v, %+v", string(resp), err)
		return nil, derr
	}
}

// $<length>\r\n<data>\r\n
func parseBulkString(bufreader *bufio.Reader) ([]byte, error) {

	first_line, err := bufreader.ReadSlice('\n')
	if err != nil {
		derr := fmt.Errorf("91383384310 [Wafris] parseBulkString read second line failure: %+v", err)
		return nil, derr
	}

	switch first_line[0] {
	case '$':
		resp := strings.TrimSpace(string(first_line))
		length, err := strconv.ParseInt(resp[1:], 10, 32)
		if err != nil {
			derr := fmt.Errorf("91383384302 [Wafris] ParseInt failure: %+v", err)
			return nil, derr
		}

		// debug
		// log.Printf("8158334087 [Wafris] DEBUG: parseBulkString response expected length %d", length)
		// log.Printf("8158334088 [Wafris] DEBUG: parseBulkString response %d, %+v", len(resp), resp)

		second_line, err := bufreader.ReadSlice('\n')
		if err != nil {
			derr := fmt.Errorf("91383384303 [Wafris] parseBulkString read second line failure: %+v", err)
			return nil, derr
		}

		bulk_string_value := second_line[:length]

		// debug
		// log.Printf("8158334089 [Wafris] DEBUG: bulk_string_value response expected length %d", length)
		// log.Printf("8158334090 [Wafris] DEBUG: bulk_string_value response %d, %+v", len(bulk_string_value), string(bulk_string_value))

		return bulk_string_value, nil

	case '-':
		resp, err := bufreader.ReadSlice('\n')
		derr := fmt.Errorf("91454862630 [Wafris] parseBulkString returned error: %+v, %+v", resp, err)
		return nil, derr
	default:
		resp, err := bufreader.ReadSlice('\n')
		derr := fmt.Errorf("91454862631 [Wafris] parseBulkString unexpected response: %+v, %+v", resp, err)
		return nil, derr
	}
}
