package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"net"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

//FeedClient fetch feed from nordnet
type FeedClient struct {
	conn net.Conn
}

//OpenFeedClient to the given addr and session key
func OpenFeedClient(addr, key string) *FeedClient {
	conf := &tls.Config{
		//InsecureSkipVerify: true,
	}
	dialer := &net.Dialer{
		Timeout:   10000000 * time.Millisecond,
		KeepAlive: 10000000 * time.Millisecond,
	}

	con, err := tls.DialWithDialer(dialer, "tcp", addr, conf)
	if err != nil {
		log.Error(err)
		return nil
	}
	args := map[string]interface{}{
		"session_key": key,
		"service":     Service,
	}
	login := map[string]interface{}{
		"cmd":  "login",
		"args": args,
	}

	if err := json.NewEncoder(con).Encode(&login); err != nil {
		log.Error(err)
	}
	return &FeedClient{
		conn: con,
	}
}

//Close FeedClient
func (c *FeedClient) Close() error {
	return c.conn.Close()
}

//GetFeed receive feed from nordnet
func (c *FeedClient) GetFeed(jsonCmd map[string]interface{}) map[string]interface{} {
	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	log.Infof("send json command %v\n", jsonCmd)

	json.NewEncoder(os.Stdout).Encode(jsonCmd)
	if err := json.NewEncoder(c.conn).Encode(jsonCmd); err != nil {
		log.Error(err)
		return nil
	}
	c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	scanner := bufio.NewScanner(c.conn)
	result := make(map[string]interface{})
	for scanner.Scan() {
		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			continue
		}
		if result["type"].(string) != "heartbeat" { //ignore heartbeat msg
			return result
		}
	}
	return nil

}
