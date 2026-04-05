package metric

import (
	"errors"
	"net/http"
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/registry"
	log "github.com/sirupsen/logrus"
)

type Server interface {
	HandleHTTPError(c backends.Context, err error, errorCode int) bool
	SendHTTPReply(c backends.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c backends.Context, payloadType string)
	Validator() security.Validator
	ExecutorDB() database.ExecutorDatabase
	MetricDB() database.MetricDatabase
}

type Handlers struct {
	server Server
}

func NewHandlers(server Server) *Handlers {
	return &Handlers{
		server: server,
	}
}

func (h *Handlers) RegisterHandlers(handlerRegistry *registry.HandlerRegistry) error {
	if err := handlerRegistry.Register(rpc.SetMetricPayloadType, h.HandleSetMetric); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetMetricPayloadType, h.HandleGetMetric); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetMetricsPayloadType, h.HandleGetMetrics); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetAllMetricsPayloadType, h.HandleGetAllMetrics); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.GetMetricHistoryPayloadType, h.HandleGetMetricHistory); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveMetricPayloadType, h.HandleRemoveMetric); err != nil {
		return err
	}
	if err := handlerRegistry.Register(rpc.RemoveAllMetricsPayloadType, h.HandleRemoveAllMetrics); err != nil {
		return err
	}
	return nil
}

// requireExecutorOwnership checks that the recoveredID belongs to an executor
// with the given name in the given colony. Returns error if not.
func (h *Handlers) requireExecutorOwnership(recoveredID string, colonyName string, executorName string) error {
	executor, err := h.server.ExecutorDB().GetExecutorByName(colonyName, executorName)
	if err != nil {
		return errors.New("Failed to find executor with name <" + executorName + ">")
	}
	if executor.ID != recoveredID {
		return errors.New("Only executor <" + executorName + "> is allowed to modify its metrics")
	}
	return nil
}

func (h *Handlers) HandleSetMetric(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateSetMetricMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to set metric, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to set metric, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.Metric.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.requireExecutorOwnership(recoveredID, msg.Metric.ColonyName, msg.Metric.ExecutorName)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	metric := msg.Metric

	// Calculate period start if period is set
	if metric.Period != core.PERIOD_NONE {
		metric.PeriodStart = core.CalculatePeriodStart(metric.Period, time.Now())
	}
	metric.GenerateID()

	if metric.MetricType == core.COUNTER {
		err = h.server.MetricDB().IncrementMetric(metric.ColonyName, metric.ExecutorName, metric.Key, metric.Period, metric.PeriodStart, metric.Value)
	} else {
		err = h.server.MetricDB().SetMetric(metric)
	}
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	result, err := h.server.MetricDB().GetMetric(metric.ColonyName, metric.ExecutorName, metric.Key, metric.Period, metric.PeriodStart)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = result.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": metric.ColonyName, "ExecutorName": metric.ExecutorName, "Key": metric.Key, "Period": metric.Period}).Debug("Setting metric")
	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetMetric(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetMetricMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get metric, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get metric, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	metric, err := h.server.MetricDB().GetMetric(msg.ColonyName, msg.ExecutorName, msg.Key, core.PERIOD_NONE, time.Time{})
	if h.server.HandleHTTPError(c, err, http.StatusBadRequest) {
		return
	}

	jsonString, err = metric.ToJSON()
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "ExecutorName": msg.ExecutorName, "Key": msg.Key}).Debug("Getting metric")
	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetMetrics(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetMetricsMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get metrics, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get metrics, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	metrics, err := h.server.MetricDB().GetMetricsByExecutorName(msg.ColonyName, msg.ExecutorName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = core.ConvertMetricArrayToJSON(metrics)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "ExecutorName": msg.ExecutorName}).Debug("Getting metrics")
	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetAllMetrics(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetAllMetricsMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get all metrics, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get all metrics, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	metrics, err := h.server.MetricDB().GetAllMetricsByExecutorName(msg.ColonyName, msg.ExecutorName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = core.ConvertMetricArrayToJSON(metrics)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "ExecutorName": msg.ExecutorName}).Debug("Getting all metrics")
	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleGetMetricHistory(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateGetMetricHistoryMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to get metric history, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to get metric history, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	metrics, err := h.server.MetricDB().GetMetricHistory(msg.ColonyName, msg.ExecutorName, msg.Key, msg.Period, msg.From, msg.To)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	jsonString, err = core.ConvertMetricArrayToJSON(metrics)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "ExecutorName": msg.ExecutorName, "Key": msg.Key, "Period": msg.Period}).Debug("Getting metric history")
	h.server.SendHTTPReply(c, payloadType, jsonString)
}

func (h *Handlers) HandleRemoveMetric(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveMetricMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to remove metric, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove metric, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.requireExecutorOwnership(recoveredID, msg.ColonyName, msg.ExecutorName)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.MetricDB().RemoveMetric(msg.ColonyName, msg.ExecutorName, msg.Key, core.PERIOD_NONE, time.Time{})
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "ExecutorName": msg.ExecutorName, "Key": msg.Key}).Debug("Removing metric")
	h.server.SendEmptyHTTPReply(c, payloadType)
}

func (h *Handlers) HandleRemoveAllMetrics(c backends.Context, recoveredID string, payloadType string, jsonString string) {
	msg, err := rpc.CreateRemoveAllMetricsMsgFromJSON(jsonString)
	if err != nil {
		h.server.HandleHTTPError(c, errors.New("Failed to remove all metrics, invalid JSON"), http.StatusBadRequest)
		return
	}

	if msg.MsgType != payloadType {
		h.server.HandleHTTPError(c, errors.New("Failed to remove all metrics, msg.MsgType does not match payloadType"), http.StatusBadRequest)
		return
	}

	err = h.server.Validator().RequireMembership(recoveredID, msg.ColonyName, true)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.requireExecutorOwnership(recoveredID, msg.ColonyName, msg.ExecutorName)
	if h.server.HandleHTTPError(c, err, http.StatusForbidden) {
		return
	}

	err = h.server.MetricDB().RemoveAllMetricsByExecutorName(msg.ColonyName, msg.ExecutorName)
	if h.server.HandleHTTPError(c, err, http.StatusInternalServerError) {
		return
	}

	log.WithFields(log.Fields{"ColonyName": msg.ColonyName, "ExecutorName": msg.ExecutorName}).Debug("Removing all metrics")
	h.server.SendEmptyHTTPReply(c, payloadType)
}
