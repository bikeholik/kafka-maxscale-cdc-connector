// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cdc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"github.com/Shopify/sarama"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Sender struct {
	KafkaBrokers string
	KafkaTopic   string
}

func (s *Sender) Send(ctx context.Context, ch <-chan map[string]interface{}) error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true

	glog.V(3).Infof("connect to brokers %s", s.KafkaBrokers)

	client, err := sarama.NewClient(strings.Split(s.KafkaBrokers, ","), config)
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
			buf := &bytes.Buffer{}
			err := json.NewEncoder(buf).Encode(data)
			if err != nil {
				return errors.Wrap(err, "encode json failed")
			}
			sequence, ok := data["sequence"]
			if !ok {
				glog.Warning("sequence not found in json")
			}
			partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
				Topic: s.KafkaTopic,
				Key:   sarama.StringEncoder(fmt.Sprintf("%v", sequence)),
				Value: sarama.ByteEncoder(buf.Bytes()),
			})
			if err != nil {
				return errors.Wrap(err, "send message to kafka failed")
			}
			glog.V(3).Infof("send message successful to %s with partition %d offset %d", s.KafkaTopic, partition, offset)
		}
	}
}
