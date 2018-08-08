package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

//RestClient is the restful client to nordnet
type RestClient struct {
	client     http.Client
	sessionKey string
}

var nordnetPublicKey = []byte(`
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA5td/64fAicX2r8sN6RP3
mfHf2bcwvTzmHrLcjJbU85gLROL+IXclrjWsluqyt5xtc/TCwMTfC/NcRVIAvfZd
t+OPdDoO0rJYIY3hOGBwLQJeLRfruM8dhVD+/Kpu8yKzKOcRdne2hBb/mpkVtIl5
avJPFZ6AQbICpOC8kEfI1DHrfgT18fBswt85deILBTxVUIXsXdG1ljFAQ/lJd/62
J74vayQJq6l2DT663QB8nLEILUKEt/hQAJGU3VT4APSfT+5bkClfRb9+kNT7RXT/
pNCctbBTKujr3tmkrdUZiQiJZdl/O7LhI99nCe6uyJ+la9jNPOuK5z6v72cXenmK
ZwIDAQAB
-----END PUBLIC KEY-----
`)

//NewRestClient create a restful client
func NewRestClient() *RestClient {
	return &RestClient{
		sessionKey: "",
		client:     http.Client{},
	}
}

//Login nordnet
func (c *RestClient) Login(user, pass string) map[string]interface{} {

	block, _ := pem.Decode(nordnetPublicKey)
	if block == nil || block.Type != "PUBLIC KEY" {
		log.Fatal("failed to decode PEM block containing public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}

	// Construct the base for the auth parameter
	timestamp := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(int(time.Now().UnixNano() / 1000000))))
	username := base64.StdEncoding.EncodeToString([]byte(user))
	password := base64.StdEncoding.EncodeToString([]byte(pass))
	login := username + ":" + password + ":" + timestamp
	// RSA encrypt it using NNAPI public key
	out, err := rsa.EncryptPKCS1v15(rand.Reader, (pub).(*rsa.PublicKey), []byte(login))
	if err != nil {
		log.Fatalf("encrypt: %s", err)
	}
	// Encode the encrypted data in Base64
	encoded := base64.StdEncoding.EncodeToString(out)
	//loginParams := base64.URLEncoding.EncodeToString([]byte(encoded))
	params := url.Values{}
	params.Set("service", Service)
	params.Add("auth", encoded)
	urlStr := ServerURL + APIVersion + "/login"
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(params.Encode()))
	req.Header.Add("Accept", "application/json")
	fmt.Printf("%v\n", urlStr)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}
	resp, err := c.client.Do(req)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}
	fmt.Printf("%v: %v\n", resp.StatusCode, resp.Status)
	result := make(map[string]interface{})
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("%v\n", result)
	c.sessionKey = result["session_key"].(string)
	log.Infof("%v\n", c.sessionKey)
	return result
}

//Get resouce from the given url from nordnet
func (c *RestClient) Get(url string) []map[string]interface{} {
	log.Infof("Get %v\n", url)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(c.sessionKey, c.sessionKey)
	if resp, err := c.client.Do(req); err == nil && resp != nil {
		log.Infof("Get response %v\n", resp.StatusCode)
		defer resp.Body.Close()
		result := make([]map[string]interface{}, 1)
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Error(err)
			return nil
		}
		return result
	}
	return nil
}

//Post data to nordnet to the given url and read reponse
func (c *RestClient) Post(url string, params *url.Values) []map[string]interface{} {
	log.Infof("Post to %v\n", url)
	req, _ := http.NewRequest("POST", url, strings.NewReader(params.Encode()))
	req.SetBasicAuth(c.sessionKey, c.sessionKey)
	req.Header.Add("Accept", "application/json")
	if resp, err := c.client.Do(req); err == nil && resp != nil {
		defer resp.Body.Close()
		result := make([]map[string]interface{}, 1)
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Error(err)
			return nil
		}
		return result
	}
	return nil
}

//Logout from nordnet
func (c *RestClient) Logout() {
	urlStr := ServerURL + APIVersion + "/login/" + c.sessionKey
	req, _ := http.NewRequest("DELETE", urlStr, nil)
	req.SetBasicAuth(c.sessionKey, c.sessionKey)
	req.Header.Add("Accept", "application/json")
	resp, _ := c.client.Do(req)
	log.Infof("logout status: %v\n", resp.Status)
}
