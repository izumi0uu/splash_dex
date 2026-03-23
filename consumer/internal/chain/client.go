package chain

import (
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

type Client struct {
	conn     *websocket.Conn
	url      string
	logger   logx.Logger
	isClosed bool
}

func NewClient(url string) *Client {
	return &Client{
		url:    url,
		logger: logx.WithContext(nil).WithFields(logx.Field("service", "websocket-client")),
	}
}

func (c *Client) Connect() error {
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = time.Second * 5

	for {
		c.logger.Infof("Connecting to WebSocket: %v", c.url)

		conn, _, err := dialer.Dial(c.url, nil)
		if err != nil {
			c.logger.Errorf("WebSocket Dial error: %v, retrying in 1 second...", err)
			time.Sleep(1 * time.Second)
			continue
		}

		c.conn = conn
		c.isClosed = false
		c.logger.Infof("WebSocket connected successfully")
		return nil
	}
}

// send text message to websocket
func (c *Client) SendMessage(message []byte) error {
	if c.conn == nil {
		return fmt.Errorf("websocket connection is not established")
	}
	if c.isClosed {
		return fmt.Errorf("websocket connection is closed")
	}

	err := c.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		c.logger.Errorf("SendMessage error: %v", err)
		// if send message error, try to reconnect
		if strings.Contains(err.Error(), "close") || strings.Contains(err.Error(), "broken pipe") {
			c.logger.Info("Connection lost, attempting to reconnect...")
			_ = c.Connect()
		}
		return err
	}
	return nil
}

// read message from websocket
func (c *Client) ReadMessage() ([]byte, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("websocket connection is not established")
	}
	if c.isClosed {
		return nil, fmt.Errorf("websocket connection is closed")
	}

	_, message, err := c.conn.ReadMessage()
	if err != nil {
		c.logger.Errorf("ReadMessage error: %v", err)

		// if connection is closed or pipe is broken, try to reconnect
		if strings.Contains(err.Error(), "close") || strings.Contains(err.Error(), "broken pipe") {
			c.logger.Info("Connection lost, attempting to reconnect...")
			_ = c.Connect()
		}
		return nil, err
	}

	return message, nil
}

func (c *Client) Close() error {
	c.isClosed = true
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) IsConnected() bool {
	return c.conn != nil && !c.isClosed
}

func (c *Client) Reconnect() error {
	if c.conn != nil {
		_ = c.conn.Close()
	}
	c.conn = nil
	return c.Connect()
}
