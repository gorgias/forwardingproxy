// Copyright (C) 2018 Betalo AB - All Rights Reserved

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBasicProxyAuth(t *testing.T) {
	// Arrange

	cases := []struct {
		name         string
		givenAuth    string
		expectedUser string
		expectedPass string
		expectedAuth bool
	}{
		{
			name:         "ValidAuth",
			givenAuth:    "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==",
			expectedUser: "Aladdin",
			expectedPass: "open sesame",
			expectedAuth: true,
		},
		{
			name:         "InvalidAuth",
			givenAuth:    "Basic ####",
			expectedUser: "",
			expectedPass: "",
			expectedAuth: false,
		},
		{
			name:         "InvalidPrefix",
			givenAuth:    "Foo QWxhZGRpbjpvcGVuIHNlc2FtZQ==",
			expectedUser: "",
			expectedPass: "",
			expectedAuth: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Act

			observedUser, observedPass, observedAuth := parseBasicProxyAuth(tc.givenAuth)

			// Assert

			assert.Equal(t, tc.expectedUser, observedUser)
			assert.Equal(t, tc.expectedPass, observedPass)
			assert.Equal(t, tc.expectedAuth, observedAuth)
		})
	}
}

func TestNewForwardingHTTPProxy(t *testing.T) {
	// Arrange

	// Proxy server

	forwardingHTTPProxy := NewForwardingHTTPProxy(nil)
	proxyServer := httptest.NewServer(forwardingHTTPProxy)
	defer proxyServer.Close()

	proxyServerURL, err := url.Parse(proxyServer.URL)
	require.NoError(t, err)

	// Destination server
	destServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hop-by-hop headers are removed when sent to the backend.
		// See: https://golang.org/src/net/http/httputil/reverseproxy.go#L129

		_, found := r.Header["Proxy-Authorization"]
		assert.False(t, found)
		assert.Equal(t, "bar", r.Header.Get("Content-type"))

		fmt.Fprintln(w, "dummy-response")
	}))
	defer destServer.Close()

	// Act

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyServerURL),
		},
	}

	req, err := http.NewRequest("GET", destServer.URL, nil)
	require.NoError(t, err)

	req.Header.Set("Proxy-Authorization", "Basic foo")
	req.Header.Set("Content-type", "bar")

	resp, err := client.Do(req)
	require.NoError(t, err)

	// Assert

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, "dummy-response", strings.TrimSpace(string(b)))
}
