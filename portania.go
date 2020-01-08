package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {

	host := flag.String("h", "", "The host that will be scanned.")
	timeout := flag.Int64("t", 30, "The timeout duration in seconds.")
	portList := flag.String("p", "", "A comma seperated list containing the ports to scan.\n\tE.G. usage :  80,443,3389,1433.")
	portRange := flag.String("pr", "", "A port range as 'port'-'port'.\n\tE.G. usage : 80-443 would scan all ports from 80 through 443")
	flag.Parse()

	duration := time.Duration(*timeout) * time.Second
	ports, err := getPorts(*portList, *portRange)
	if err != nil {
		log.Fatal(err.Error())
	}

	connectionBroker(duration, *host, ports)

}

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

		return ports, nil
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
		return ports, nil
	}

	return nil, errors.New("no ports found to parse")

}

func connectionBroker(duration time.Duration, host string, ports []int) {

	work := make(chan string)
	go func() {
		for _, p := range ports {
			work <- fmt.Sprintf("%v:%v", host, p)
		}
		close(work)
	}()

	f := sync.WaitGroup{}
	f.Add(5)

	for x := 0; x < 5; x++ {
		go func() {
			for c := range work {
				if ok := testConnection(c, duration); ok {
					fmt.Printf("Connected to %v\n", c)
				} else {
					fmt.Printf("failed to connect to %v\n", c)
				}
			}
			f.Done()
		}()
	}
	f.Wait()
}

func testConnection(host string, duration time.Duration) bool {
	con, err := net.DialTimeout("tcp", host, duration)
	if err != nil {
		return false
	}
	fmt.Printf("Connction successful %v\n", host)
	con.Close()
	return true
}
