package main

import (
	"io/ioutil"
	"net/http"
	"math/rand"
	"time"
	"regexp"
	"errors"
	"strings"
)

func getClientIP() (string, error) {
	var kindOfValidIP = regexp.MustCompile(`^([1-9][0-9]{0,2})(\.[0-9]{0,3}){3}$`)
	apiUrls := []string {
		"https://api.ipify.org?format=text",
		"http://myexternalip.com/raw",
		"https://ident.me/",
		"http://icanhazip.com",
		"https://ipecho.net/plain",
		"https://ifconfig.co/ip",
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(apiUrls), func(i, j int) { apiUrls[i], apiUrls[j] = apiUrls[j], apiUrls[i] })

	for _, url := range apiUrls {
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		ipBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		ip := strings.Trim(string(ipBytes), "\n \t")
		if  !kindOfValidIP.MatchString(ip) {
			continue
		}

		return ip, nil
	}

	return "", errors.New("Could not get public IP")
}