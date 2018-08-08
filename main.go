package main

import (
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func main() {

	user := flag.String("user", "your username", "username for login")
	pass := flag.String("passwd", "replace with your pass", "login passwd")
	flag.Parse()

	baseURL := ServerURL + APIVersion

	rest := NewRestClient()
	result := rest.Login(*user, *pass)
	defer rest.Logout()
	//Get account info
	log.Info(rest.Get(baseURL + "/accounts"))
	//Get markets info
	log.Info(rest.Get(baseURL + "/markets"))
	//Get list info
	log.Info(rest.Get(baseURL + "/lists"))

	pmap := result["private_feed"].(map[string]interface{})
	addr := fmt.Sprintf("%s:%d", pmap["hostname"], int(pmap["port"].(float64)))
	log.Infof("connecting to feed server: %s\n", addr)
	privFeed := OpenFeedClient(addr, result["session_key"].(string))
	defer privFeed.Close()

	pmap = result["public_feed"].(map[string]interface{})
	addr = fmt.Sprintf("%s:%d", pmap["hostname"], int(pmap["port"].(float64)))
	log.Infof("connecting to feed server: %s\n", addr)
	publFeed := OpenFeedClient(addr, result["session_key"].(string))
	defer publFeed.Close()

	//Get feed about Ericsson
	args := map[string]interface{}{
		"t": "price",
		"m": 11,
		"i": "101",
	}
	subscribe := map[string]interface{}{
		"cmd":  "subscribe",
		"args": args,
	}
	log.Info(publFeed.GetFeed(subscribe))
}
