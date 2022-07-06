package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

func isEnabled() bool {
	checkArgs := strings.Fields("advfirewall firewall show rule name=stream-deck-destiny-solo")
	checkCmd := exec.Command("netsh", checkArgs...)
	checkOut, _ := checkCmd.CombinedOutput()
	return strings.Contains(string(checkOut), "Ok.")
}

func toggle() string {
	ruleDefinition := "advfirewall firewall add rule name=stream-deck-destiny-solo dir=%s action=block protocol=%s remoteport=27000-27200,3097"
	if isEnabled() {
		args := strings.Fields("advfirewall firewall delete rule name=stream-deck-destiny-solo")
		exec.Command("netsh", args...).Run()
		return "false"
	} else {
		exec.Command("netsh", strings.Fields(fmt.Sprintf(ruleDefinition, "IN", "TCP"))...).Run()
		exec.Command("netsh", strings.Fields(fmt.Sprintf(ruleDefinition, "IN", "UDP"))...).Run()
		exec.Command("netsh", strings.Fields(fmt.Sprintf(ruleDefinition, "OUT", "TCP"))...).Run()
		exec.Command("netsh", strings.Fields(fmt.Sprintf(ruleDefinition, "OUT", "UDP"))...).Run()
		return "true"
	}
}

var upgrader = websocket.Upgrader{}

func ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("recv: %s", message)
		msg := string(message)
		if strings.Contains(msg, "toggle") {
			c.WriteMessage(mt, []byte(toggle()))
		} else if strings.Contains(msg, "exit") {
			os.Exit(0)
		} else {
			c.WriteMessage(mt, []byte(strconv.FormatBool(isEnabled())))
		}
	}
}

func main() {
	http.HandleFunc("/", ws)
	log.Fatal(http.ListenAndServe("localhost:33334", nil))
}
