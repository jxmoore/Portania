package main

import (
	"testing"
)

// TestPortaniaGetPorts tests the getPorts function in Portania.
func TestPortaniaGetPorts(t *testing.T) {

	testSuite := map[string]struct {
		portRange string
		portList  string
		err       string
		ports     []int
	}{
		"getPorts should throw an error due to nil values": {
			err: "no ports found to parse",
		},
		"getPorts using portList should return the ports 80,443,8080": {
			portList: "80,443,8080",
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
