package main

import (
    "io/ioutil"
    "net/http"
    "math/rand"
    "time"
    "regexp"
    "errors"
    "strings"
    "fmt"
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


func getHttpResponseCode(url string) (int, error) {
    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }

    resp, err := client.Get(url)

    if err != nil {
        return -1, err
    }

    return resp.StatusCode, nil
}

func isValidRequest(url string, retries uint) (bool, error) {

    statusCodeProp := 0
    for retries >  0 {
        retries--
        statusCode, err := getHttpResponseCode(url)
        statusCodeProp = statusCode

        if err != nil {
            if retries == 0 {
                return false, err
            }

            continue;
        }

        if statusCode >= 200 && statusCode <= 399{
            return true, nil
        }

        if retries > 0 {
            time.Sleep(15 * time.Second)
        }
    }

    return false, errors.New(fmt.Sprintf("Response code %d. Expected in <200; 399>", statusCodeProp))
}
