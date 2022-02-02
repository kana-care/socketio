package socketio_test

// THIS FILE DOES NOT CONTAIN TESTS...
// this file contains utilities for tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type mesg struct {
	method, url, data string
}

func connect(t *testing.T, svr *httptest.Server, numSvrs int, ver, tspType string) []string {
	client := svr.Client()

	if numSvrs == 0 {
		numSvrs = 1
	}
	sids := make([]string, numSvrs)

	for i := 0; i < numSvrs; i++ {
		url := fmt.Sprintf("%s/socket.io/?EIO=%s&transport=%s&t=%d&name=%s&con=%d", svr.URL, ver, tspType, time.Now().UnixNano(), strings.TrimPrefix(t.Name(), "TestServerV1Basic"), i)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			t.Fatal(err)
		}

		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(res.Body)

		str := buf.String()
		buf.Reset()

		var (
			m map[string]interface{}
			s = strings.Index(str, "{")
			e = strings.Index(str, "}") + 1
		)

		if err := json.Unmarshal([]byte(str[s:e]), &m); err != nil {
			t.Fatal(err)
		}

		sids[i] = m["sid"].(string)
	}
	return sids
}

func testMessage(t *testing.T, svr *httptest.Server, ver string, sids []string, msg mesg) {
	client := svr.Client()

	rs := []string{"${eio}", fmt.Sprintf("EIO=%s", ver)}
	for i, sid := range sids {
		rs = append(rs, fmt.Sprintf("${sid_%d}", i), fmt.Sprintf("sid=%s", sid))
	}
	rs = append(rs, "${t}", fmt.Sprintf("t=%d", time.Now().UnixNano()))

	rep := strings.NewReplacer(rs...)
	url := fmt.Sprintf("%s%s", svr.URL, rep.Replace(msg.url))

	var body io.Reader
	if strings.ToUpper(msg.method) != "GET" && msg.data != "" {
		body = strings.NewReader(msg.data)
	}

	req, err := http.NewRequest(msg.method, url, body)
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	if err := res.Body.Close(); err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {

		t.Error(buf.String())
		t.Fatal("bad status code:", res.StatusCode)
	}

	if strings.ToUpper(msg.method) == "GET" {
		if buf.String() != msg.data {
			t.Log((msg.url))
			t.Log(rep.Replace(msg.url))
			t.Errorf("[%v] have: %q want: %q", sids, buf.String(), msg.data)
		}
	}
}

func disconnect(t *testing.T, svr *httptest.Server, ver string, sids []string) {
	client := svr.Client()

	for _, sid := range sids {
		url := fmt.Sprintf("%s/socket.io/?EIO=%s&sid=%s&t=%d&disconnect=true", svr.URL, ver, sid, time.Now().UnixNano())

		req, err := http.NewRequest("POST", url, strings.NewReader("2:41"))
		if err != nil {
			t.Fatal(err)
		}

		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != 200 {

			t.Error("disconnect..")
			t.Fatal("bad status code:", res.StatusCode)
		}
	}
}
