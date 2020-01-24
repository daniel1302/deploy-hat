package main

import (
	"testing"
	"regexp"
)

func TestGetClientIP(t *testing.T) {
	var kindOfValidIP = regexp.MustCompile(`^([1-9][0-9]{0,2})(\.[0-9]{0,3}){3}$`)

	for i := 0; i < 5; i++ {
		ip, err := getClientIP()

		if !kindOfValidIP.MatchString(ip) {
			t.Errorf("Could not get public IP.")
		}

		if err != nil {
			t.Errorf("Unexpected Error: %s", err.Error())
		}
	}
}

func TestGetHttpResponseCode(t *testing.T) {
	pagesTable := []struct{
		url  string
		code int
	}{
		{"http://google.pl/", 301},
		{"https://www.google.com/", 200},
		{"http://google.com/some_non_existing_page", 404},
	}

	for _, page := range pagesTable {
		if resp, _ := getHttpResponseCode(page.url); resp != page.code {
			t.Errorf("Invalid response code for %s. Expected %d. Got %d.", page.url, page.code, resp)
		}
	}
}


func TestIsValidRequest(t *testing.T) {
	pagesTable := []struct{
		url    string
		status bool
	}{
		{"http://google.pl/", true},
		{"https://www.google.com/", true},
		{"http://google.com/some_non_existing_page", false},
	}

	for _, page := range pagesTable {
		if resp, _ := isValidRequest(page.url, 0); resp != false {
			t.Errorf("Check for page %s is invalid. Expected %t. Got %t.", page.url, false, resp)
		}

		if resp, _ := isValidRequest(page.url, 1); resp != page.status {
			t.Errorf("Check for page %s is invalid. Expected %t. Got %t.", page.url, page.status, resp)
		}
	}
}
