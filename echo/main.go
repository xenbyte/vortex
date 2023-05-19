// Challenge #1: Echo
// In this challenge,
// our node will receive an "echo" message from Maelstrom.
// The messages will look like this:
/*
   {
     "src": "c1",
     "dest": "n1",
     "body": {
       "type": "echo",
       "msg_id": 1,
       "echo": "Please echo 35"
     }
   }
*/
// we will send a message with the same body back to the client
// but with a message type of "echo_ok".
// It will also associate itself with the original message,
// by setting the "in_reply_to" field to the original message
//
// The reply message will look something like this:
/*
  {
    "src": "n1",
    "dest": "c1",
    "body": {
      "type": "echo_ok",
      "msg_id": 1,
      "in_reply_to": 1,
      "echo": "Please echo 35"
    }
  }
*/

package main

import (
	"encoding/json"
	"log"

	"github.com/xenbyte/vortex/core"
)

func main() {
	n := core.NewMaelstromNode()

	n.Handle("echo", func(msg core.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body["type"] = "echo_ok"

		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
