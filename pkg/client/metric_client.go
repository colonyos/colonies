package client

import (
	"context"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) SetMetric(metric core.Metric, prvKey string) (core.Metric, error) {
	msg := rpc.CreateSetMetricMsg(metric)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return core.Metric{}, err
	}

	respBodyString, err := client.sendMessage(rpc.SetMetricPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return core.Metric{}, err
	}

	return core.ConvertJSONToMetric(respBodyString)
}

func (client *ColoniesClient) GetMetric(colonyName string, executorName string, key string, prvKey string) (core.Metric, error) {
	msg := rpc.CreateGetMetricMsg(colonyName, executorName, key)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return core.Metric{}, err
	}

	respBodyString, err := client.sendMessage(rpc.GetMetricPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return core.Metric{}, err
	}

	return core.ConvertJSONToMetric(respBodyString)
}

func (client *ColoniesClient) GetMetrics(colonyName string, executorName string, prvKey string) ([]core.Metric, error) {
	msg := rpc.CreateGetMetricsMsg(colonyName, executorName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetMetricsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToMetricArray(respBodyString)
}

func (client *ColoniesClient) GetAllMetrics(colonyName string, executorName string, prvKey string) ([]core.Metric, error) {
	msg := rpc.CreateGetAllMetricsMsg(colonyName, executorName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetAllMetricsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToMetricArray(respBodyString)
}

func (client *ColoniesClient) GetMetricHistory(colonyName string, executorName string, key string, period int, from time.Time, to time.Time, prvKey string) ([]core.Metric, error) {
	msg := rpc.CreateGetMetricHistoryMsg(colonyName, executorName, key, period, from, to)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetMetricHistoryPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToMetricArray(respBodyString)
}

func (client *ColoniesClient) RemoveMetric(colonyName string, executorName string, key string, prvKey string) error {
	msg := rpc.CreateRemoveMetricMsg(colonyName, executorName, key)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveMetricPayloadType, jsonString, prvKey, false, context.TODO())
	return err
}

func (client *ColoniesClient) RemoveAllMetrics(colonyName string, executorName string, prvKey string) error {
	msg := rpc.CreateRemoveAllMetricsMsg(colonyName, executorName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveAllMetricsPayloadType, jsonString, prvKey, false, context.TODO())
	return err
}
