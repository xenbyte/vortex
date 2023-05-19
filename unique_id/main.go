package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/xenbyte/vortex/core"
)

func main() {

	n := core.NewMaelstromNode()

	n.Handle("generate", func(msg core.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body["type"] = "generate_ok"

		var randomID int64
		err := binary.Read(rand.Reader, binary.BigEndian, &randomID)
		if err != nil {
			log.Fatal(err)
		}

		body["id"] = fmt.Sprintf("%v%v", time.Now().UnixNano(), randomID)

		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
