package sender

import (
	"context"
	"time"

	"41/internal/stype"
	"41/internal/utils"

	kafka "github.com/segmentio/kafka-go"
	"github.com/urfave/cli/v2"
)

var confLogger = utils.GetLogger("sender/kafka")

type KafkaSender struct {
	Topic     string
	Host      string
	Partition int
	Interval  int
	Producer  chan *stype.HTTPRequestResponseRecord
}

func (sender *KafkaSender) Send(item *stype.HTTPRequestResponseRecord) (err error) {
	sender.Producer <- item
	confLogger.Println(item.EncodeToString())
	return nil
}

func (sender *KafkaSender) initConsumer() {
	confLogger.Println(sender.Partition)
	for i := 0; i < int(sender.Partition); i++ {
		go func(index int) {
			confLogger.Println(index)
			bufferLen := 100
			conn, err := kafka.DialLeader(context.Background(), "tcp", sender.Host, sender.Topic, index)
			if err != nil {
				confLogger.Fatal("failed to dial leader:", err)
			}
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			defer conn.Close()

			ticker := time.NewTicker(time.Duration(sender.Interval) * time.Second)
			messages := make([]kafka.Message, 0, bufferLen)
			defer ticker.Stop()

			for {
				select {
				case item := <-sender.Producer:
					messages = append(messages, kafka.Message{Value: item.EncodeToBytes()})
					if len(messages) >= 100 {
						messages = make([]kafka.Message, 0, bufferLen)
						_, err = conn.WriteMessages(messages...)
						if err != nil {
							confLogger.Println("kafka sender failed:", err)
						}
					}
				case <-ticker.C:
					if len(messages) > 0 {
						_, err = conn.WriteMessages(messages...)
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
		Topic:     ctx.String("kafka-topic"),
		Host:      ctx.String("kafka-host"),
		Partition: ctx.Int("kafka-topic-partitions"),
		Interval:  ctx.Int("kafka-send-interval"),
		Producer:  make(chan *stype.HTTPRequestResponseRecord, ctx.Int("kafka-send-queue")),
	}
	sender.initConsumer()
	confLogger.Println("NewKafkaSender done")
	return sender
}
