package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
	"honnef.co/go/netdb"
)

func ErrorMessage() {
	fmt.Println("Usage: go run ./port-scanner.go <host> <OPTION>")
	os.Exit(1)
}

func getServiceName(port int, protocol string) string {
	proto := netdb.GetProtoByName(protocol)
	if proto == nil {
		return "Unknown Protocol"
	}
	service := netdb.GetServByPort(port, proto)
	if service != nil {
		return service.Name
	}
	return "Unknown Service"
}

func scanPort(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func scanHost(host string, portStart int, portEnd int) {
	var openNum int
	var wg sync.WaitGroup
	startTime := time.Now()

	maxGoroutines := 100
	guard := make(chan struct{}, maxGoroutines)

	fmt.Printf("Scanning host: %s\nPort range: %d-%d\n========================\n", host, portStart, portEnd)
	for port := portStart; port <= portEnd; port++ {
		guard <- struct{}{}
		wg.Add(1)

		go func(p int) {
			defer wg.Done()
			if scanPort(host, p) {
				serviceName := getServiceName(p, "tcp")
				fmt.Printf("Port %d is open | Service: %s\n", p, serviceName)
				openNum++
			}
			<-guard
		}(port)
	}
	wg.Wait()
	elapsed := time.Since(startTime)
	fmt.Println("========================")
	fmt.Printf("Scan finished: %d open ports found\nTotal time usage: %s\n", openNum, elapsed)
}

func main() {
	var portStart, portEnd int = 1, 1024

	argc := len(os.Args)
	switch {
	case argc == 2:
		scanHost(os.Args[1], portStart, portEnd)
	case argc == 3:
		if os.Args[2] == "-p-" {
			scanHost(os.Args[1], 1, 65535)
		} else {
			ErrorMessage()
		}
	case argc == 4:
		// var option string = os.Args[2]
		var portRange string = os.Args[3]
		_, err1 := fmt.Sscanf(portRange, "%d ", &portStart)
		if err1 != nil {
			fmt.Println("Case range")
			_, err := fmt.Sscanf(portRange, "%d-%d", &portStart, &portEnd)
			if portStart > portEnd || err != nil {
				fmt.Println("Invalid port range")
				os.Exit(1)
			}
		} else {
			fmt.Println("Case Solo")
			portEnd = portStart
		}
		scanHost(os.Args[1], portStart, portEnd)

	default:
		ErrorMessage()
	}

}