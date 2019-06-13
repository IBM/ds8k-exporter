package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/common/log"
	"github.com/tidwall/gjson"
)

type DS8kClient struct {
	UserName   string
	Password   string
	AuthToken  string
	IpAddress  string
	ErrorCount float64
	Location   string
}

func (ds8kClient *DS8kClient) RetriveAuthToken() (authToken string, err error) {
	reqAuthURL := "https://" + ds8kClient.IpAddress + ":8452/api/v1/tokens"
	httpclient := &http.Client{Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	},
		Timeout: 45 * time.Second}

	postValue := []byte(`{ "request" : { "params" : { "username" : "` + ds8kClient.UserName + `" , "password" : "` + ds8kClient.Password + ` "} } }`)
	req, _ := http.NewRequest("POST", reqAuthURL, bytes.NewBuffer(postValue))
	req.Header.Add("Content-Type", "application/json")
	resp, err := httpclient.Do(req)
	if err != nil {
		log.Errorf("Error doing http request URL[%s] Error: %v", reqAuthURL, err)
		err = fmt.Errorf("Error doing http request URL[%s] Error: %v", reqAuthURL, err)
		return
	}
	defer resp.Body.Close()

	log.Debugf("Response Status Code: %v", resp.StatusCode)
	log.Debugf("Response Status: %v", resp.Status)
	// log.Debugf("Response Body: %v", resp.Body)

	if resp.StatusCode != 200 {
		// we didnt get a good response code, so bailing out
		log.Errorln("Got a non 200 response code: ", resp.StatusCode)
		log.Debugln("response was: ", resp)
		ds8kClient.ErrorCount++
		return "", fmt.Errorf("received non 200 error code: %v. the response was: %v", resp.Status, resp)
	}
	respbody, err := ioutil.ReadAll(resp.Body)
	body := string(respbody)
	log.Debugf("Response Body: %v", body)
	tokenInfo := gjson.Get(body, "token").String()
	authToken = gjson.Get(tokenInfo, "token").String()
	log.Debugf("AuthToken is: %v", authToken)
	return authToken, err

}

func (ds8kClient *DS8kClient) CallDS8kAPI(request string) (body string, err error) {
	httpclient := &http.Client{Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	},
		Timeout: 45 * time.Second}

	// New POST request
	req, _ := http.NewRequest("GET", request, nil)
	// header parameters
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Auth-Token", ds8kClient.AuthToken)
	resp, err := httpclient.Do(req)
	if err != nil {
		return "", fmt.Errorf("\n Error connecting to : %v. the error was: %v", request, err)
	}
	defer resp.Body.Close()
	respbody, err := ioutil.ReadAll(resp.Body)
	body = string(respbody)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("\nGot error code: %v when accessing URL: %s\n Body text is: %s", resp.StatusCode, request, respbody)
	}
	return body, nil

}
