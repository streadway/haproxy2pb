package haproxy

import (
	"fmt"
	"regexp"
	"net"
	"time"
	"strings"
	"strconv"
)

// Oct 28 14:30:05 ams-mid006.int.s-cloud.net haproxy[10227]: 10.20.3.41:45572 [28/Oct/2011:14:30:05.615] soundcloud soundcloud/ams-app053-8003 0/0/0/18/18 302 677 - - ---- 157/126/126/0/0 0/0 {med} {7} "GET /stream/pNAFM5DjlcZj?url=http%3A//api.soundcloud.com/users/4641029&referer=http%3A//static.ak.facebook.com/common/referer_frame.php&auto_play=true&height=209&color=3b5998&width=398&show_artwork=false&flash_version=MAC%2011%2C0%2C1%2C152&consumer_key=sc_player HTTP/1.0"

var Format = regexp.MustCompile(`^(\w+ \d+ \S+) (\S+) (\S+)\[(\d+)\]: (\S+):(\d+) \[(\S+)\] (\S+) (\S+)/(\S+) (\S+) (\S+) (\S+) *(\S+) (\S+) (\S+) (\S+) (\S+) (?:\{([^}]*)\} )?(?:\{([^}]*)\} )?"(\S+) ([^"]+) (\S+)" *$`)

var EventTimeLayout = "2/Jan/2006:15:04:05.000"

var Time = regexp.MustCompile(`(\d+)/(\w+)/(\d+):(\d+):(\d+):(\d+).(\d+)`)

var Reason_short = map[string]string{
	"C": "CLIENT_ABORT",
	"S": "SERVER_ABORT",
	"P": "PROXY_ABORT",
	"R": "RESOURCE_LIMIT", 
	"I": "INTERNAL_ERROR", 
	"c": "CLIENT_TIMEOUT", 
	"s": "SERVER_TIMEOUT",
}
var ProxyState_short = map[string]string{
	"R": "REQUEST",
	"Q": "QUEUE",
	"C": "CONNECTION",
	"H": "HEADERS",
	"D": "DATA",
	"L": "LAST",
	"T": "TARPIT",
}

var Cookie_short = map[string]string{
	"N": "NO",
	"I": "INVALID",
	"D": "DOWN",
	"V": "VALID",
	"E": "EXPIRED",
	"O": "OLD",
}

var CookieTransform_short =map[string]string{
	"N": "NONE",
	"I": "INSERTED",
	"U": "UPDATED",
	"P": "PROVIDED",
	"R": "REWRITTEN",
	"D": "DELETED",
}

func _ip(str string) ([]byte) {
	if i := net.ParseIP(str); i != nil {
		return i
	}
	return nil
}

func scanTime(str string, req *Request) {
	if t, err := time.Parse(EventTimeLayout, str); err == nil {
		year := uint32(t.Year())
		month := uint32(t.Month())
		day := uint32(t.Day())
		hour := uint32(t.Hour())
		min := uint32(t.Minute())
		sec := uint32(t.Second())
		nsec := uint32(t.Nanosecond())

		req.Year = &year
		req.Month = &month
		req.Day = &day
		req.Hour = &hour
		req.Minute = &min
		req.Second = &sec
		req.NanoSecond = &nsec
	}
}

func _time(str string) (*float32) {
	if t, err := time.Parse(EventTimeLayout, str); err == nil {
		var seconds = float32(t.Unix()) + float32(t.UnixNano()) / 1e6
		return &seconds
	}
	return nil
}

func _int32(str string) (*int32) {
	if i, err := strconv.ParseInt(str, 10, 32); err == nil {
		i32 := int32(i)
		return &i32
	}
	return nil
}

func _uint32(str string) (*uint32) {
	if i, err := strconv.ParseInt(str, 10, 32); err == nil {
		ui32 := uint32(i)
		return &ui32
	}
	return nil
}

func scanTimings(str string, req *Request) {
	if times := strings.Split(str, "/"); len(times) >= 5 {
		req.TimeQueue                  = count(times[0])
		req.TimeWait                   = count(times[1])
		req.TimeConnect                = count(times[2])
		req.TimeRespond                = count(times[3])
		req.TimeTotal                  = count(times[4])
	}
}

func count(str string) (*int32) {
	if i, err := strconv.ParseInt(str, 10, 32); err == nil {
		if i >= 0 {
			i32 := int32(i)
			return &i32
		}
	}
	return nil
}

func scanTermination(str string, req *Request) {
	if len(str) >= 4 {
		req.TerminationReason = (*Request_Reason)(termination(str[0:1], Reason_short, Request_Reason_value))

		req.TerminationState = (*Request_ProxyState)(termination(str[1:2], ProxyState_short, Request_ProxyState_value))

		if len(str) > 2 {
			req.TerminationCookie = (*Request_Cookie)(termination(str[2:3], Cookie_short, Request_Cookie_value))
			req.TerminationCookieTransform = (*Request_CookieTransform)(termination(str[3:4], CookieTransform_short, Request_CookieTransform_value))
		}
	}
}

func termination(short string, shortMap map[string]string, enumMap map[string]int32) (*int32) {
	enum := shortMap[short]
	if enum != "" {
		tag := enumMap[enum]
		if tag != 0 {
			return &tag
		}
	}
	return nil
}

func scanConnections(str string, req *Request) {
	if counts := strings.Split(str, "/"); len(counts) >= 5 {
		req.ActiveConnections          = count(counts[0])
		req.FrontendConnections        = count(counts[1])
		req.BackendConnections         = count(counts[2])
		req.ServerConnections          = count(counts[3])
		req.Retries                    = count(counts[4])
	}
}

func _string(s string) (*string) {
	if s != "-" {
		return &s
	}
	return nil
}

func _headers(s string) ([]string) {
	if headers := strings.Split(s, "|"); len(headers) > 0 {
		return headers
	}
	return nil
}

func scanQueues(str string, req *Request) {
	if counts := strings.Split(str, "/"); len(counts) >= 2 {
		req.ServerQueue = count(counts[0])
		req.BackendQueue = count(counts[1])
	}
}

func debug(matches []string) {
	for i, v := range matches {
		fmt.Printf("%d: %#v\n", i, v)
	}
}

func Scan(line string, req *Request) (error) {
	var m []string

	if m = Format.FindStringSubmatch(line); m == nil {
		return fmt.Errorf("No match %s", m)
	}

	//debug(m)

	req.Host                       = _string(m[2])
	req.Process                    = _string(m[3])
	req.Pid                        = _int32(m[4])
	req.ClientIp                   = _ip(m[5])
	req.ClientPort                 = _uint32(m[6])
	scanTime(m[7], req)
	req.Frontend                   = _string(m[8])
	req.Backend                    = _string(m[9])
	req.Server                     = _string(m[10])
	scanTimings(m[11], req)
	req.StatusCode                 = _int32(m[12])
	req.BytesRead                  = _int32(m[13])
	req.RequestCookie              = _string(m[14])
	req.ResponseCookie             = _string(m[15])
	scanTermination(m[16], req)
	scanConnections(m[17], req)
	scanQueues(m[18], req)

	req.RequestHeaders						= _headers(m[19])
	req.ResponseHeaders						= _headers(m[20])

	req.HttpMethod								= _string(m[21])
	req.HttpUri										= _string(m[22])
	req.HttpVersion								= _string(m[23])

	return nil
}
