package core

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type HandlerFunc func(msg Message) error

type Message struct {
	Src  string          `json:"src,omitempty"`
	Dest string          `json:"dest,omitempty"`
	Body json.RawMessage `json:"body,omitempty"`
}

type MessageBody struct {
	Type       string `json:"type,omitempty"`
	MessageID  int    `json:"msg_id,omitempty"`
	InReplyTo  int    `json:"in_reply_to,omitempty"`
	ErrCode    int    `json:"code,omitempty"`
	ErrMessage string `json:"text,omitempty"`
}

type InitMessageBody struct {
	MessageBody
	NodeID  string   `json:"node_id,omitempty"`
	NodeIDs []string `json:"node_ids,omitempty"`
}

type Node struct {
	// For Reading messages from Maelstrom
	Stdin io.Reader
	// For Writing messages from Maelstrom
	Stdout io.Writer
	mu     sync.Mutex
	wg     sync.WaitGroup

	id            string
	nodeIDs       []string
	nextMessageID string

	handlers  map[string]HandlerFunc
	callbacks map[int]HandlerFunc
}

func NewMaelstromNode() *Node {
	return &Node{
		handlers:  make(map[string]HandlerFunc),
		callbacks: make(map[int]HandlerFunc),

		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	}
}

func (n *Node) Reply(req Message, body any) error {
	var reqBody MessageBody
	if err := json.Unmarshal(req.Body, &reqBody); err != nil {
		return err
	}

	b := make(map[string]any)

	if buf, err := json.Marshal(body); err != nil {
		return err
	} else if err := json.Unmarshal(buf, &b); err != nil {
		return err
	}
	b["in_reply_to"] = reqBody.MessageID

	return n.Send(req.Src, b)
}

func (n *Node) Handle(typ string, fn HandlerFunc) {
	if _, ok := n.handlers[typ]; ok {
		log.Fatalf("dupilicate mesage handler for %q message type", typ)
	}
	n.handlers[typ] = fn
}

func (n *Node) Send(dst string, body any) error {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return err
	}

	buf, err := json.Marshal(Message{
		Src:  n.id,
		Dest: dst,
		Body: bodyJSON,
	})

	if err != nil {
		return err
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	log.Printf("sent %v", string(buf))
	if _, err = n.Stdout.Write(buf); err != nil {
		return err
	}

	_, err = n.Stdout.Write([]byte{'\n'})
	return err
}

func (n *Node) ID() string {
	return n.id
}

func (n *Node) Init(id string, nodeIDs []string) {
	n.id = id
	n.nodeIDs = nodeIDs
}

func (n *Node) NodeIDs() []string {
	return n.nodeIDs
}

func (n *Node) RPC(dst string, body any, handler HandlerFunc) error {
	return nil
}

func (n *Node) Run() error {
	scanner := bufio.NewScanner(n.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		var msg Message
		if err := json.Unmarshal(line, &msg); err != nil {
			return fmt.Errorf("unmarshal message: %v", err)
		}

		var body MessageBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return fmt.Errorf("unmarshal message body: %v", err)
		}

		log.Printf("received: %v", string(msg.Body))

		if body.InReplyTo != 0 {
			n.mu.Lock()
			h := n.callbacks[body.InReplyTo]
			delete(n.callbacks, body.InReplyTo)
			n.mu.Unlock()

			if h == nil {
				log.Printf("ignoring reply to %d with no callback ", body.InReplyTo)
				continue
			}

			n.wg.Add(1)
			go func() {
				defer n.wg.Done()
				n.HandleCallback(h, msg)
			}()
			continue
		}

		var h HandlerFunc
		if body.Type == "init" {
			h = n.HandleInitMessage
		} else if h = n.handlers[body.Type]; h == nil {
			return fmt.Errorf("No handler for %s", line)
		}

		n.wg.Add(1)
		go func() {
			defer n.wg.Done()
			h(msg)
		}()
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	n.wg.Wait()

	return nil

}

func (n *Node) SyncRPC(ctx context.Context, dst string, body any) (Message, error) {
	return Message{}, nil
}

func (n *Node) HandleCallback(h HandlerFunc, msg Message) {
	if err := h(msg); err != nil {
		log.Printf("callback error: %s", err)
	}
}

func (n *Node) HandleInitMessage(msg Message) error {
	var body InitMessageBody
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return fmt.Errorf("unmarshall init message body: %v", err)
	}

	n.Init(body.NodeID, body.NodeIDs)

	if h := n.handlers["init"]; h != nil {
		if err := h(msg); err != nil {
			return err
		}
	}

	// Send back a response that the node has been initialized.
	return n.Reply(msg, MessageBody{Type: "init_ok"})
}
