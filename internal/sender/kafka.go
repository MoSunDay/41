package sender

import (
	"context"
	"time"

	"41/internal/utils"

	kafka "github.com/segmentio/kafka-go"
	"github.com/urfave/cli/v2"
)

var confLogger = utils.GetLogger("sender/kafka")

type KafkaSender struct {
	Topic    string
	Host     string
	Worker   int
	Interval int
	Producer chan []byte
}

func (sender *KafkaSender) Send(message []byte) (err error) {
	sender.Producer <- message
	return nil
}

func (sender *KafkaSender) initConsumer() {
	for i := 0; i < int(sender.Worker); i++ {
		go func(index int) {
			bufferLen := 10
			conn := &kafka.Writer{
				Addr:         kafka.TCP(sender.Host),
				Topic:        sender.Topic,
				Async:        true,
				RequiredAcks: 1,
				Balancer:     &kafka.LeastBytes{},
			}

			ticker := time.NewTicker(time.Duration(sender.Interval) * time.Second)
			messages := make([]kafka.Message, 0, bufferLen)
			defer ticker.Stop()

			for {
				select {
				case message := <-sender.Producer:
					messages = append(messages, kafka.Message{Value: message})
					if len(messages) >= 10 {
						err := conn.WriteMessages(context.Background(), messages...)
						if err != nil {
							confLogger.Println("kafka sender failed:", err)
						}
						messages = make([]kafka.Message, 0, bufferLen)
					}
				case <-ticker.C:
					if len(messages) > 0 {
						err := conn.WriteMessages(context.Background(), messages...)
						if err != nil {
							confLogger.Println("kafka sender failed:", err)
						}
						messages = make([]kafka.Message, 0, bufferLen)
					}
				}
			}
		}(i)
	}
}

func NewKafkaSender(ctx *cli.Context) Sender {
	sender := &KafkaSender{
		Topic:    ctx.String("kafka-topic"),
		Host:     ctx.String("kafka-host"),
		Worker:   ctx.Int("kafka-worker"),
		Interval: ctx.Int("kafka-send-interval"),
		Producer: make(chan []byte, ctx.Int("kafka-send-queue")),
	}
	sender.initConsumer()
	confLogger.Println("NewKafkaSender done")
	return sender
}
