package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {

	host := flag.String("h", "", "The host that will be scanned.")
	timeout := flag.Int64("t", 30, "The timeout duration in seconds.")
	workers := flag.Int("w", 3, "The number of workers (threads) to use when scanning the remote host.")
	portList := flag.String("p", "", "A comma seperated list containing the ports to scan.\n\tE.G. usage :  80,443,3389,1433.")
	portRange := flag.String("pr", "", "A port range as 'port'-'port'.\n\tE.G. usage : 80-443 would scan all ports from 80 through 443")
	splay := flag.Bool("s", false, "Enable splay, this causes a random sleep between each port scanned.")

	flag.Parse()

	ports, err := getPorts(*portList, *portRange)
	if err != nil {
		log.Fatal(err.Error())
	}

	if *workers == 0 {
		*workers = 1
	}
	if *timeout == 0 {
		*timeout = 30
	}

	duration := time.Duration(*timeout) * time.Second
	connectionBroker(duration, *workers, *host, ports, *splay)

}

// getPorts takes two strings and uses either of those to construct a splice of ints that represents a port range.
func getPorts(portList, portRange string) ([]int, error) {

	var ports []int

	if portList != "" {
		for _, p := range strings.Split(portList, ",") {

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
func connectionBroker(duration time.Duration, workers int, host string, ports []int, splay bool) {

	work := make(chan string)
	rand.NewSource(time.Now().UnixNano())

	go func() {

		for _, p := range ports {
			work <- fmt.Sprintf("%v:%v", host, p)
			if splay {
				time.Sleep(time.Second * time.Duration(rand.Intn(17)))
			}
		}

		close(work)

	}()

	wg := sync.WaitGroup{}
	wg.Add(workers)

	for x := 0; x < workers; x++ {

		go func() {

			for w := range work {
				if ok, err := testConnection(w, duration); ok {
					fmt.Printf("Connected to %v\n", w)
				} else {
					fmt.Printf("failed to connect to %v : %v\n", w, err)
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
