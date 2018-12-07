package topics

import (
	"github.com/Shopify/sarama"
	"github.com/random-dwi/kafkactl/util/output"
)

type Topic struct {
	Name       string
	Partitions []partition `json:",omitempty" yaml:",omitempty"`
	Configs    []config    `json:",omitempty" yaml:",omitempty"`
}

type partition struct {
	Id           int32
	OldestOffset int64
	NewestOffset int64
	Leader       string  `json:",omitempty" yaml:",omitempty"`
	Replicas     []int32 `json:",omitempty" yaml:",omitempty"`
	ISRs         []int32 `json:",omitempty" yaml:",omitempty"`
}

type config struct {
	Name  string
	Value string
}

func ReadTopic(client *sarama.Client, admin *sarama.ClusterAdmin, name string, readPartitions bool, readLeaders bool, readReplicas bool, readConfigs bool) (Topic, error) {
	var (
		err           error
		ps            []int32
		led           *sarama.Broker
		configEntries []sarama.ConfigEntry
		top           = Topic{Name: name}
	)

	if !readPartitions {
		return top, nil
	}

	if ps, err = (*client).Partitions(name); err != nil {
		return top, err
	}

	for _, p := range ps {
		np := partition{Id: p}

		if np.OldestOffset, err = (*client).GetOffset(name, p, sarama.OffsetOldest); err != nil {
			return top, err
		}

		if np.NewestOffset, err = (*client).GetOffset(name, p, sarama.OffsetNewest); err != nil {
			return top, err
		}

		if readLeaders {
			if led, err = (*client).Leader(name, p); err != nil {
				return top, err
			}
			np.Leader = led.Addr()
		}

		if readReplicas {
			if np.Replicas, err = (*client).Replicas(name, p); err != nil {
				return top, err
			}

			if np.ISRs, err = (*client).InSyncReplicas(name, p); err != nil {
				return top, err
			}
		}

		top.Partitions = append(top.Partitions, np)
	}

	if readConfigs {

		configResource := sarama.ConfigResource{
			Type: sarama.TopicResource,
			Name: name,
		}

		if configEntries, err = (*admin).DescribeConfig(configResource); err != nil {
			output.Failf("failed to describe config: %v", err)
		}

		for _, configEntry := range configEntries {

			if !configEntry.Default {
				entry := config{Name: configEntry.Name, Value: configEntry.Value}
				top.Configs = append(top.Configs, entry)
			}
		}
	}

	return top, nil
}
