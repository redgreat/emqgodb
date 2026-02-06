package mqtt

import (
	"crypto/tls"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/redgreat/emqgodb/src/config"
	"go.uber.org/zap"
)

// Client MQTT客户端封装
type Client struct {
	client  mqtt.Client
	config  *config.EMQXConfig
	logger  *zap.Logger
	handler MessageHandler
}

// MessageHandler 消息处理接口
type MessageHandler interface {
	HandleMessage(topic string, payload []byte) error
}

// NewClient 创建MQTT客户端
func NewClient(cfg *config.EMQXConfig, logger *zap.Logger, handler MessageHandler) (*Client, error) {
	c := &Client{
		config:  cfg,
		logger:  logger,
		handler: handler,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(cfg.ClientID)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetCleanSession(true)

	// 连接回调
	opts.SetOnConnectHandler(c.onConnect)
	opts.SetConnectionLostHandler(c.onConnectionLost)
	opts.SetReconnectingHandler(c.onReconnecting)

	// SSL配置
	if cfg.SSL.Enabled {
		opts.SetTLSConfig(c.newTLSConfig())
	}

	c.client = mqtt.NewClient(opts)
	return c, nil
}

// newTLSConfig 创建TLS配置
func (c *Client) newTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}

// Connect 连接到MQTT服务器
func (c *Client) Connect() error {
	c.logger.Info("正在连接MQTT服务器",
		zap.String("broker", c.config.Broker),
		zap.String("client_id", c.config.ClientID))

	token := c.client.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("连接MQTT服务器失败: %w", token.Error())
	}

	return nil
}

// Disconnect 断开连接
func (c *Client) Disconnect() {
	c.logger.Info("正在断开MQTT连接")
	c.client.Disconnect(1000)
}

// onConnect 连接成功回调
func (c *Client) onConnect(client mqtt.Client) {
	c.logger.Info("MQTT连接成功，正在订阅主题",
		zap.String("topic", c.config.Topic))

	token := client.Subscribe(c.config.Topic, byte(c.config.QoS), c.messageCallback)
	if token.Wait() && token.Error() != nil {
		c.logger.Error("订阅主题失败",
			zap.String("topic", c.config.Topic),
			zap.Error(token.Error()))
		return
	}

	c.logger.Info("订阅主题成功",
		zap.String("topic", c.config.Topic),
		zap.Int("qos", c.config.QoS))
}

// onConnectionLost 连接断开回调
func (c *Client) onConnectionLost(client mqtt.Client, err error) {
	c.logger.Warn("MQTT连接断开",
		zap.Error(err))
}

// onReconnecting 重连回调
func (c *Client) onReconnecting(client mqtt.Client, opts *mqtt.ClientOptions) {
	c.logger.Info("正在重新连接MQTT服务器")
}

// messageCallback 消息回调
func (c *Client) messageCallback(client mqtt.Client, msg mqtt.Message) {
	c.logger.Debug("收到消息",
		zap.String("topic", msg.Topic()),
		zap.Int("payload_size", len(msg.Payload())))

	if err := c.handler.HandleMessage(msg.Topic(), msg.Payload()); err != nil {
		c.logger.Error("处理消息失败",
			zap.String("topic", msg.Topic()),
			zap.Error(err))
	}
}

// IsConnected 检查是否已连接
func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}
