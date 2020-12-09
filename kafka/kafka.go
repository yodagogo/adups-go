package kafka

import (
	"fmt"
	"log"

	"github.com/Shopify/sarama"
)

//Config kafka连接配置信息
type Config struct {
	BrokerList        []string
	accessLogProducer sarama.AsyncProducer
	Topic             string
}

//NewConfig 初始化kakfa client
func (cfg *Config) NewConfig(brokerList []string, topic string) *Config {
	return &Config{BrokerList: brokerList, Topic: topic}
}

//NewProducer 初始化producer client

func (cfg *Config) newSyncProducer() (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true

	// On the broker side, you may want to change the following settings to get
	// stronger consistency guarantees:
	// - For your broker, set `unclean.leader.election.enable` to false
	// - For the topic, you could increase `min.insync.replicas`.
	producer, err := sarama.NewSyncProducer(cfg.BrokerList, config)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
		return nil, err
	}

	return producer, nil
}

//SyncConsumer 订阅
func (cfg *Config) SyncConsumer() error {

	return nil
}

//SyncProducer 异步发送producer
func (cfg *Config) SyncProducer(val string) error {

	producer, err := cfg.newSyncProducer()
	if err != nil {
		return err
	}

	partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: cfg.Topic,
		Value: sarama.StringEncoder(val),
	})
	if err != nil {
		log.Fatalf("Failed to store your data:%s to kafka", val)

	} else {
		fmt.Printf("Your data is stored with unique identifier important %d %d\n", partition, offset)
	}

	return err
}
