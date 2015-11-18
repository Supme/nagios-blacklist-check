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
	"errors"
)

func BlackList(ip string, bl []string, v string) ([]string, error) {
	result := []string{}

	rIp, ipv := reverseIP(ip)
	if ipv != "ipv4" { return result, errors.New(ipv + " not supported")}

	if v != "n" { log.Println("Reverse:", rIp) }

	var wg sync.WaitGroup

	//ToDo add ipv6 blacklist support
	for _, b := range bl {
		wg.Add(1)
		go func(b string) {
			i, err := net.LookupHost(rIp + "." + b)
			if err == nil {
				result = append(result, b)
			}
			if v != "n" { log.Println("Checked: ", b, i) }
			defer wg.Done()
		}(b)

	}
	wg.Wait()

	return result, nil
}

/* Reverse the IP given to us.
 * 127.0.0.1 -> 1.0.0.127
 * 2001:4130:8:67d2::3363 -> 3.6.3.3.0.0.0.0.0.0.0.0.0.0.0.0.2.d.7.6.8.0.0.0.0.d.1.4.1.0.0.2
 *
 * return reverseIp, versionIp
 */
func reverseIP(ip string) (string, string) {
	ipVersion := "not valid ip"
	var stringSplitIP []string

	if net.ParseIP(ip) == nil { return "", ipVersion}
	if net.ParseIP(ip).To4() != nil { // Check for an IPv4 address
		ipVersion = "ipv4"
		stringSplitIP = strings.Split(ip, ".") // Split into 4 groups
		for x, y := 0, len(stringSplitIP)-1; x < y; x, y = x+1, y-1 {
			stringSplitIP[x], stringSplitIP[y] = stringSplitIP[y], stringSplitIP[x] // Reverse the groups
		}
	} else {
		ipVersion = "ipv6"

		stringSplitIP = strings.Split(ip, ":") // Split into however many groups

		/* Due to IPv6 lookups being different than IPv4 we have an extra check here
		We have to expand the :: and do 0-padding if there are less than 4 digits */
		for key := range stringSplitIP {
			if len(stringSplitIP[key]) == 0 { // Found the ::
				stringSplitIP[key] = strings.Repeat("0000", 8-strings.Count(ip, ":"))
			} else if len(stringSplitIP[key]) < 4 { // 0-padding needed
				stringSplitIP[key] = strings.Repeat("0", 4-len(stringSplitIP[key])) + stringSplitIP[key]
			}
		}

		// We have to join what we have and split it again to get all the letters split individually
		stringSplitIP = strings.Split(strings.Join(stringSplitIP, ""), "")

		for x, y := 0, len(stringSplitIP)-1; x < y; x, y = x+1, y-1 {
			stringSplitIP[x], stringSplitIP[y] = stringSplitIP[y], stringSplitIP[x]
		}
	}

	return strings.Join(stringSplitIP, "."), ipVersion // Return the IP.
}

func main()  {

	ip := flag.String("i", "", "IP address")
	listfile := flag.String("f", "blacklist.txt", "Black list file")
	verbose := flag.String("v", "n", "Verbose check log (y or n)")
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
		var line string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line = scanner.Text()
			//check comment line begin '#'
			if line[0] != 35 {
				lines = append(lines, line)
			}
		}

		if scanner.Err() != nil {
			fmt.Println(scanner.Err())
			os.Exit(3)
		}

		result, err := BlackList(*ip, lines, *verbose)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(3)
		}

		if len(result) != 0 {
			fmt.Printf("IP %s in blacklists:", *ip)
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

}
