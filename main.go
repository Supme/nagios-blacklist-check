package main

import (
	"net"
	"fmt"
	"strings"
	"sync"
	"log"
	"flag"
	"os"
	"bufio"
)

func BlackList(ip string, bl []string, v string) []string {

	result := []string{}

	var wg sync.WaitGroup

	// ToDo check ip version and reverse true method
	ipArr := strings.Split(ip, ".")

	reverseIp := ""
	if len(ipArr) == 4 {
		reverseIp = ipArr[3] + "." + ipArr[2] + "." +ipArr[1] + "." +ipArr[0]
	}

	for _, b := range bl {
		wg.Add(1)
		go func(b string) {
			i, e := net.LookupHost(reverseIp + "." + b)
			if e == nil {
				result = append(result, b)
			}
			if v != "no" {log.Println("Checked: ", b, i)}
			defer wg.Done()
		}(b)

	}
	wg.Wait()

	return result
}

func main()  {

	ip := flag.String("i", "", "IP address")
	listfile := flag.String("f", "blacklist.txt", "Black list file")
	flag.Parse()

	if *ip != "" {

		/*
		 Reply for nagios status
		"STATUS_OK", 0
		"STATUS_WARNING", 1
		"STATUS_CRITICAL", 2
		"STATUS_UNKNOWN", 3
		 */

		file, err := os.Open(*listfile)
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			//ToDo add check comment line
			lines = append(lines, scanner.Text())
		}

		if scanner.Err() != nil {
			fmt.Println(scanner.Err())
			os.Exit(3)
		}

		fmt.Println(lines)

		result := BlackList(*ip, lines, "n")

		if len(result) != 0 {
			fmt.Printf("IP %s in blacklist: ", *ip)
			fmt.Println(result)
			if len(result) > 1 {
				os.Exit(2)
			} else {
				os.Exit(1)
			}

		} else {
			fmt.Printf("OK ip %s not in blacklist\n", *ip)
			os.Exit(0)
		}


	} else {
		fmt.Println("Use '-h' parameter for help text")
	}

	_ = listfile

}
