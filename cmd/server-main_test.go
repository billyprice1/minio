/*
 * Minio Cloud Storage, (C) 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"flag"
	"net/http"
	"os"
	"runtime"
	"testing"

	"github.com/minio/cli"
)

func TestGetListenIPs(t *testing.T) {
	testCases := []struct {
		addr       string
		port       string
		shouldPass bool
	}{
		{"localhost", "9000", true},
		{"", "9000", true},
		{"", "", false},
	}
	for _, test := range testCases {
		var addr string
		// Please keep this we need to do this because
		// of odd https://play.golang.org/p/4dMPtM6Wdd
		// implementation issue.
		if test.port != "" {
			addr = test.addr + ":" + test.port
		}
		hosts, port, err := getListenIPs(&http.Server{
			Addr: addr,
		})
		if !test.shouldPass && err == nil {
			t.Fatalf("Test should fail but succeeded %s", err)
		}
		if test.shouldPass && err != nil {
			t.Fatalf("Test should succeed but failed %s", err)
		}
		if test.shouldPass {
			if port != test.port {
				t.Errorf("Test expected %s, got %s", test.port, port)
			}
			if len(hosts) == 0 {
				t.Errorf("Test unexpected value hosts cannot be empty %#v", test)
			}
		}
	}
}

func TestFinalizeEndpoints(t *testing.T) {
	testCases := []struct {
		tls  bool
		addr string
	}{
		{false, ":80"},
		{true, ":80"},
		{false, "localhost:80"},
		{true, "localhost:80"},
	}

	for i, test := range testCases {
		endPoints := finalizeEndpoints(test.tls, &http.Server{Addr: test.addr})
		if len(endPoints) <= 0 {
			t.Errorf("Test case %d returned with no API end points for %s",
				i+1, test.addr)
		}
	}
}

// Tests all the expected input disks for function checkSufficientDisks.
func TestCheckSufficientDisks(t *testing.T) {
	var xlDisks []string
	if runtime.GOOS == "windows" {
		xlDisks = []string{
			"C:\\mnt\\backend1",
			"C:\\mnt\\backend2",
			"C:\\mnt\\backend3",
			"C:\\mnt\\backend4",
			"C:\\mnt\\backend5",
			"C:\\mnt\\backend6",
			"C:\\mnt\\backend7",
			"C:\\mnt\\backend8",
			"C:\\mnt\\backend9",
			"C:\\mnt\\backend10",
			"C:\\mnt\\backend11",
			"C:\\mnt\\backend12",
			"C:\\mnt\\backend13",
			"C:\\mnt\\backend14",
			"C:\\mnt\\backend15",
			"C:\\mnt\\backend16",
			"C:\\mnt\\backend17",
		}
	} else {
		xlDisks = []string{
			"/mnt/backend1",
			"/mnt/backend2",
			"/mnt/backend3",
			"/mnt/backend4",
			"/mnt/backend5",
			"/mnt/backend6",
			"/mnt/backend7",
			"/mnt/backend8",
			"/mnt/backend9",
			"/mnt/backend10",
			"/mnt/backend11",
			"/mnt/backend12",
			"/mnt/backend13",
			"/mnt/backend14",
			"/mnt/backend15",
			"/mnt/backend16",
			"/mnt/backend17",
		}
	}
	// List of test cases fo sufficient disk verification.
	testCases := []struct {
		disks       []string
		expectedErr error
	}{
		// Even number of disks '6'.
		{
			xlDisks[0:6],
			nil,
		},
		// Even number of disks '12'.
		{
			xlDisks[0:12],
			nil,
		},
		// Even number of disks '16'.
		{
			xlDisks[0:16],
			nil,
		},
		// Larger than maximum number of disks > 16.
		{
			xlDisks,
			errXLMaxDisks,
		},
		// Lesser than minimum number of disks < 6.
		{
			xlDisks[0:3],
			errXLMinDisks,
		},
		// Odd number of disks, not divisible by '2'.
		{
			append(xlDisks[0:10], xlDisks[11]),
			errXLNumDisks,
		},
	}

	// Validates different variations of input disks.
	for i, testCase := range testCases {
		endpoints, err := parseStorageEndpoints(testCase.disks)
		if err != nil {
			t.Fatalf("Unexpected error %s", err)
		}
		if checkSufficientDisks(endpoints) != testCase.expectedErr {
			t.Errorf("Test %d expected to pass for disks %s", i+1, testCase.disks)
		}
	}
}

func TestCheckServerSyntax(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []cli.Command{serverCmd}
	serverFlagSet := flag.NewFlagSet("server", 0)
	cli.NewContext(app, serverFlagSet, nil)
	disksGen := func(n int) []string {
		disks, err := getRandomDisks(n)
		if err != nil {
			t.Fatalf("Unable to initialie disks %s", err)
		}
		return disks
	}
	testCases := [][]string{
		disksGen(1),
		disksGen(4),
		disksGen(8),
		disksGen(16),
	}
	for i, disks := range testCases {
		err := serverFlagSet.Parse(disks)
		if err != nil {
			t.Errorf("Test %d failed to parse arguments %s", i+1, disks)
		}
		defer removeRoots(disks)
		endpoints, err := parseStorageEndpoints(disks)
		if err != nil {
			t.Fatalf("Unexpected error %s", err)
		}
		_ = validateDisks(endpoints, nil)
	}
}

func TestIsDistributedSetup(t *testing.T) {
	var testCases []struct {
		disks  []string
		result bool
	}
	if runtime.GOOS == "windows" {
		testCases = []struct {
			disks  []string
			result bool
		}{
			{[]string{`http://4.4.4.4:80/c:\mnt\disk1`, `http://4.4.4.4:80/c:\mnt\disk2`}, true},
			{[]string{`http://4.4.4.4:9000/c:\mnt\disk1`, `http://127.0.0.1:9000/c:\mnt\disk2`}, true},
			{[]string{`http://127.0.0.1:9000/c:\mnt\disk1`, `http://127.0.0.1:9001/c:\mnt\disk2`}, true},
			{[]string{`c:\mnt\disk1`, `c:\mnt\disk2`}, false},
		}
	} else {
		testCases = []struct {
			disks  []string
			result bool
		}{
			{[]string{"http://4.4.4.4:9000/mnt/disk1", "http://4.4.4.4:9000/mnt/disk2"}, true},
			{[]string{"http://4.4.4.4:9000/mnt/disk1", "http://127.0.0.1:9000/mnt/disk2"}, true},
			{[]string{"http://127.0.0.1:9000/mnt/disk1", "http://127.0.0.1:9000/mnt/disk2"}, true},
			{[]string{"/mnt/disk1", "/mnt/disk2"}, false},
		}

	}
	for i, test := range testCases {
		endpoints, err := parseStorageEndpoints(test.disks)
		if err != nil {
			t.Fatalf("Unexpected error %s", err)
		}
		res := isDistributedSetup(endpoints)
		if res != test.result {
			t.Errorf("Test %d: expected result %t but received %t", i+1, test.result, res)
		}
	}
}

func TestInitServerConfig(t *testing.T) {
	ctx := &cli.Context{}
	root, err := newTestConfig("us-east-1")
	if err != nil {
		t.Fatal("Failed to set up test config")
	}
	defer removeAll(root)

	testCases := []struct {
		envVar string
		val    string
	}{
		{"MINIO_MAXCONN", "10"},
		{"MINIO_CACHE_SIZE", "42MB"},
		{"MINIO_CACHE_EXPIRY", "2h45m"},
		{"MINIO_ACCESS_KEY", "abcd1"},
		{"MINIO_SECRET_KEY", "abcd12345"},
	}
	for i, test := range testCases {
		tErr := os.Setenv(test.envVar, test.val)
		if tErr != nil {
			t.Fatalf("Test %d failed with %v", i+1, tErr)
		}
		initServerConfig(ctx)
	}
}
