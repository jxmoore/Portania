package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type address struct {
	port             string
	host             string
	formattedAddress string
}

func main() {

	host := flag.String("hosts", "", "A list of the hosts that will be scanned.\n\tE.G. usage 'google.com localhost github.com'."+
		"\n\tThis list should be space delimited and qutoes are neccesarry if specifying more than one host.")
	timeout := flag.Int64("timeout", 10, "The timeout duration in seconds.")
	workers := flag.Int("workers", 3, "The number of workers (threads) to use when scanning the remote host.")
	portList := flag.String("ports", "", "A delimited seperated list containing the ports to scan.\n\tE.G. usage :  '80 443 3389 1433'.\n\t"+
		"If specifying a list of ports rathen than a single port the quotes are required.")
	portRange := flag.String("portrange", "", "A port range as 'port'-'port'.\n\tE.G. usage : 80-443 would scan all ports from 80 through 443")
	splay := flag.Bool("splay", false, "Enable splay, this causes a random sleep between each port scanned.")
	hideFailures := flag.Bool("hidefailures", false, "Enabling this hides the output regarding closed ports and or connection failures.")
	debug := flag.Bool("debug", false, "Enables debug output, this will include the connection failure information.")

	flag.Parse()

	pList := strings.Fields(*portList)
	ports, err := getPorts(pList, *portRange)
	if err != nil {
		log.Fatal(err.Error())
	}

	if *workers == 0 {
		*workers = 1
	}
	if *timeout == 0 {
		*timeout = 30
	}

	var useColor bool
	hostnames := strings.Fields(*host)
	duration := time.Duration(*timeout) * time.Second

	if runtime.GOOS != "windows" {
		useColor = true
	}

	connectionBroker(duration, *workers, hostnames, ports, *splay, *hideFailures, *debug, useColor)

}

// getPorts takes two strings and uses either of those to construct a splice of ints that represents a port range.
func getPorts(portList []string, portRange string) ([]int, error) {

	var ports []int

	if len(portList) != 0 {
		for _, p := range portList {

			port, err := strconv.Atoi(p)
			if err != nil {
				fmt.Println("unable to parse port ", p)
				continue
			}

			ports = append(ports, port)
		}
	}

	if portRange != "" {

		pr := strings.Split(portRange, "-")
		lower, err := strconv.Atoi(pr[0])
		if err != nil {
			return nil, fmt.Errorf("unable to parse port %v", pr[0])
		}

		upper, err := strconv.Atoi(pr[1])
		if err != nil {
			return nil, fmt.Errorf("unable to parse port %v", pr[1])
		}

		if upper < lower {
			return nil, fmt.Errorf("the upper port range must be larger than the lower end, %v is less than %v", upper, lower)
		}

		for i := lower; i < upper+1; i++ {
			ports = append(ports, i)
		}
	}

	if len(ports) != 0 {
		return ports, nil
	}

	return nil, errors.New("no ports found to parse")

}

// connectionBroker creates a channel and pumps in all of the addresses that need to be tested - 'host+:+p'
// 'x' worker go routines are created that pull from this channel, calling testConnection and printing the result.
func connectionBroker(duration time.Duration, workers int, hostnames []string, ports []int, splay, hideFailures, debug, useColor bool) {

	work := make(chan address)
	rand.NewSource(time.Now().UnixNano())

	go func() {
		for _, h := range hostnames {
			for _, p := range ports {
				port := strconv.Itoa(p)
				work <- address{host: h, port: port, formattedAddress: h + ":" + port}
				if splay {
					time.Sleep(time.Second * time.Duration(rand.Intn(17)))
				}
			}
		}

		close(work)

	}()

	wg := sync.WaitGroup{}
	wg.Add(workers)

	for x := 0; x < workers; x++ {

		go func() {

			for w := range work {
				if ok, err := testConnection(w.formattedAddress, duration); ok {
					colorPrinter(true, useColor, fmt.Sprintf("Port %v is open on host %v\n", w.port, w.host))
				} else {
					if !hideFailures {
						if debug {
							colorPrinter(false, useColor, fmt.Sprintf("Port %v is closed on %v : %v\n", w.port, w.host, err))
						} else {
							colorPrinter(false, useColor, fmt.Sprintf("Port %v is closed on %v\n", w.port, w.host))
						}
					}
				}
				if splay {
					time.Sleep(time.Second * time.Duration(rand.Intn(8)))
				}
			}

			wg.Done()

		}()
	}

	wg.Wait()

}

// testConnection tests the connection provided in address over TCP, if successful it returns true along with a nil error,
// in the event that the timeout is reached, the connection fails or there is an error it returns false along with the error value.
func testConnection(address string, duration time.Duration) (bool, error) {

	con, err := net.DialTimeout("tcp", address, duration)
	if err != nil {
		return false, err
	}

	con.Close()

	return true, nil

}

// colorPrinter is a function that prints out the message passed in using the corresponding color code to indicate success code when the var color == true
func colorPrinter(success, useColor bool, message string) {

	if useColor {
		errorColor := "\033[1;31m%s\033[0m"
		successColor := "\033[0;36m%s\033[0m"
		if success {
			fmt.Printf(successColor, message)
		} else {
			fmt.Printf(errorColor, message)
		}
	} else {
		fmt.Printf(message)
	}

}
