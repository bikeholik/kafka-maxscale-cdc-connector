// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cdc

import (
	"context"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// KafkaSender takes a channel of []byte and send them to the given topic
type KafkaSender struct {
	KafkaBrokers string
	KafkaTopic   string
	GTIDStore    interface {
		Write(gtid *GTID) error
	}
	GTIDExtractor interface {
		Parse(line []byte) (*GTID, error)
	}
}

// Send the given messages to a topic in Kafka
func (k *KafkaSender) Send(ctx context.Context, ch <-chan []byte) error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true

	glog.V(3).Infof("connect to brokers %s", k.KafkaBrokers)

	client, err := sarama.NewClient(strings.Split(k.KafkaBrokers, ","), config)
	if err != nil {
		return errors.Wrap(err, "create client failed")
	}
	defer client.Close()

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return errors.Wrap(err, "create sync producer failed")
	}
	defer producer.Close()

	glog.V(3).Infof("wait for lines")
	for {
		select {
		case <-ctx.Done():
			return nil
		case data, ok := <-ch:
			if !ok {
				return nil
			}
			gtid, err := k.GTIDExtractor.Parse(data)
			if err != nil {
				return errors.Wrap(err, "extract gtid failed")
			}
			partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
				Topic: k.KafkaTopic,
				Key:   sarama.StringEncoder(gtid.String()),
				Value: sarama.ByteEncoder(data),
			})
			if err != nil {
				return errors.Wrap(err, "send message to kafka failed")
			}
			glog.V(3).Infof("send message successful to %s with partition %d offset %d", k.KafkaTopic, partition, offset)
			if err := k.GTIDStore.Write(gtid); err != nil {
				return errors.Wrap(err, "save gtid failed")
			}
		}
	}
}
