package main

import (
	"testing"
	"time"
)

// TestPortaniaGetPorts tests the getPorts function in Portania.
func TestPortaniaGetPorts(t *testing.T) {

	testSuite := map[string]struct {
		portRange string
		portList  []string
		err       string
		ports     []int
	}{
		"getPorts should throw an error due to nil values": {
			err: "no ports found to parse",
		},
		"getPorts using portList should return the ports 80,443,8080": {
			portList: []string{"80", "443", "8080"},
			ports:    []int{80, 443, 8080},
		},
		"getPorts using portRange should return the ports 80-85": {
			portRange: "80-85",
			ports:     []int{80, 81, 82, 83, 84, 85},
		},
	}
	for testName, testCase := range testSuite {

		t.Logf("Running test %v\n", testName)
		ports, err := getPorts(testCase.portList, testCase.portRange)
		if err != nil && err.Error() != testCase.err {
			t.Errorf("expected getPorts to fail with %v but received %v.", testCase.err, err.Error())
		} else {
			t.Logf("received the expected error result %v", testCase.err)
		}

		if len(testCase.ports) != 0 {

			for _, p := range testCase.ports {

				match := false
				for _, x := range ports {
					if p == x {
						match = true
						break
					}
				}

				if match == false {
					t.Errorf("%v was not found in the returned slice from getPorts", p)
				}
			}
		}
	}
}

// ExampleConnectionBrokerPass tests the output printed when a connection is attempted during
// the connectionBrokers worker go routine
func ExampleConnectionBroker() {

	connectionBroker(time.Second*5, 3, []string{"google.com", "github.com", "jomoLocal.fail"}, []int{443}, false, false, false, false)
	// Output:
	//Port 443 is open on host google.com
	//Port 443 is open on host github.com
	//Port 443 is closed on jomoLocal.fail
}

// TestPortaniaConnection tests the testConnection function in Portania.
func TestPortaniaConnection(t *testing.T) {

	testSuite := map[string]struct {
		addr     string
		duration time.Duration
		err      string
		success  bool
	}{
		"testConnection should throw an due to a closed port on localhost": {
			addr:     "localhost:555",
			success:  false,
			duration: time.Second * 5,
			err:      "dial tcp [::1]:555: connect: connection refused",
		},
		"testConnection should return true indicating that the port is open": {
			addr:     "github.com:443",
			duration: time.Second * 5,
			success:  true,
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("Running test %v\n", testName)

		ok, err := testConnection(testCase.addr, testCase.duration)
		if err != nil && err.Error() != testCase.err {
			t.Errorf("expected testConnection to fail with %v but received %v\n", testCase.err, err.Error())
		} else {
			t.Logf("received the expected error result %v\n", testCase.err)
		}

		if ok != testCase.success {
			t.Errorf("expected testConnection to return %v but received %v\n", testCase.success, ok)
		} else {
			t.Logf("received the expected response from testConnection\n")
		}
	}
}
