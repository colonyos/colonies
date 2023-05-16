package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type controllerMock struct {
	returnError bool
}

func (v *controllerMock) getCronPeriod() int {
	return -1
}

func (v *controllerMock) getGeneratorPeriod() int {
	return -1
}

func (v *controllerMock) getEtcdServer() *cluster.EtcdServer {
	return nil
}

func (v *controllerMock) getEventHandler() *eventHandler {
	return nil
}

func (v *controllerMock) getThisNode() cluster.Node {
	return cluster.Node{}
}

func (v *controllerMock) subscribeProcesses(executorID string, subscription *subscription) error {
	return nil
}

func (v *controllerMock) subscribeProcess(executorID string, subscription *subscription) error {
	return nil
}

func (v *controllerMock) getColonies() ([]*core.Colony, error) {
	return nil, nil
}

func (v *controllerMock) getColony(colonyID string) (*core.Colony, error) {
	return nil, nil
}

func (v *controllerMock) addColony(colony *core.Colony) (*core.Colony, error) {
	return nil, nil
}

func (v *controllerMock) deleteColony(colonyID string) error {
	return nil
}

func (v *controllerMock) renameColony(colonyID string, name string) error {
	return nil
}

func (v *controllerMock) addExecutor(executor *core.Executor, allowExecutorReregister bool) (*core.Executor, error) {
	return nil, nil
}

func (v *controllerMock) getExecutor(executorID string) (*core.Executor, error) {
	return nil, nil
}

func (v *controllerMock) getExecutorByColonyID(colonyID string) ([]*core.Executor, error) {
	return nil, nil
}

func (v *controllerMock) approveExecutor(executorID string) error {
	return nil
}

func (v *controllerMock) rejectExecutor(executorID string) error {
	return nil
}

func (v *controllerMock) deleteExecutor(executorID string) error {
	return nil
}

func (v *controllerMock) addProcessToDB(process *core.Process) (*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) addProcess(process *core.Process) (*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) addChild(processGraphID string, parentProcessID string, childProcessID string, process *core.Process, executorID string, insert bool) (*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) getProcess(processID string) (*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) findProcessHistory(colonyID string, executorID string, seconds int, state int) ([]*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) findPrioritizedProcesses(executorID string, colonyID string, count int) ([]*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) findWaitingProcesses(colonyID string, count int) ([]*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) findRunningProcesses(colonyID string, count int) ([]*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) findSuccessfulProcesses(colonyID string, count int) ([]*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) findFailedProcesses(colonyID string, count int) ([]*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) updateProcessGraph(graph *core.ProcessGraph) error {
	return nil
}

func (v *controllerMock) createProcessGraph(workflowSpec *core.WorkflowSpec, args []interface{}, rootInput []interface{}) (*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) submitWorkflowSpec(workflowSpec *core.WorkflowSpec) (*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) getProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) findWaitingProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) findRunningProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) findSuccessfulProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) findFailedProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) deleteProcess(processID string) error {
	return nil
}

func (v *controllerMock) deleteAllProcesses(colonyID string, state int) error {
	return nil
}

func (v *controllerMock) deleteProcessGraph(processID string) error {
	return nil
}

func (v *controllerMock) deleteAllProcessGraphs(colonyID string, state int) error {
	return nil
}

func (v *controllerMock) closeSuccessful(processID string, executorID string, output []interface{}) error {
	return nil
}

func (v *controllerMock) notifyChildren(process *core.Process) error {
	return nil
}

func (v *controllerMock) closeFailed(processID string, errs []string) error {
	return nil
}

func (v *controllerMock) handleDefunctProcessgraph(processGraphID string, processID string, err error) error {
	return nil
}

func (v *controllerMock) assign(executorID string, colonyID string) (*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) unassignExecutor(processID string) error {
	return nil
}

func (v *controllerMock) resetProcess(processID string) error {
	return nil
}

func (v *controllerMock) getColonyStatistics(colonyID string) (*core.Statistics, error) {
	return nil, nil
}

func (v *controllerMock) getStatistics() (*core.Statistics, error) {
	return nil, nil
}

func (v *controllerMock) addAttribute(attribute *core.Attribute) (*core.Attribute, error) {
	return nil, nil
}

func (v *controllerMock) getAttribute(attributeID string) (*core.Attribute, error) {
	return nil, nil
}

func (v *controllerMock) addFunction(function *core.Function) (*core.Function, error) {
	return nil, nil
}

func (v *controllerMock) getFunctionsByExecutorID(executorID string) ([]*core.Function, error) {
	return nil, nil
}

func (v *controllerMock) getFunctionsByColonyID(colonyID string) ([]*core.Function, error) {
	return nil, nil
}

func (v *controllerMock) getFunctionByID(functionID string) (*core.Function, error) {
	return nil, nil
}

func (v *controllerMock) deleteFunction(functionID string) error {
	return nil
}

func (v *controllerMock) addGenerator(generator *core.Generator) (*core.Generator, error) {
	if v.returnError {
		return nil, errors.New("error")
	}

	return nil, nil
}

func (v *controllerMock) getGenerator(generatorID string) (*core.Generator, error) {
	return nil, nil
}

func (v *controllerMock) resolveGenerator(generatorName string) (*core.Generator, error) {
	return nil, nil
}

func (v *controllerMock) getGenerators(colonyID string, count int) ([]*core.Generator, error) {
	return nil, nil
}

func (v *controllerMock) packGenerator(generatorID string, colonyID, arg string) error {
	return nil
}

func (v *controllerMock) generatorTriggerLoop() {
}

func (v *controllerMock) triggerGenerators() {
}

func (v *controllerMock) submitWorkflow(generator *core.Generator, counter int) {
}

func (v *controllerMock) addCron(cron *core.Cron) (*core.Cron, error) {
	return nil, nil
}

func (v *controllerMock) deleteGenerator(generatorID string) error {
	return nil
}

func (v *controllerMock) getCron(cronID string) (*core.Cron, error) {
	return nil, nil
}

func (v *controllerMock) getCrons(colonyID string, count int) ([]*core.Cron, error) {
	return nil, nil
}

func (v *controllerMock) runCron(cronID string) (*core.Cron, error) {
	return nil, nil
}

func (v *controllerMock) deleteCron(cronID string) error {
	return nil
}

func (v *controllerMock) calcNextRun(cron *core.Cron) time.Time {
	return time.Time{}
}

func (v *controllerMock) startCron(cron *core.Cron) {
}

func (v *controllerMock) triggerCrons() {
}

func (v *controllerMock) cronTriggerLoop() {
}

func (v *controllerMock) resetDatabase() error {
	return nil
}

func (v *controllerMock) stop() {
}

func (v *controllerMock) isLeader() bool {
	return false
}

func (v *controllerMock) tryBecomeLeader() bool {
	return false
}

func (v *controllerMock) timeoutLoop() {
}

func (v *controllerMock) blockingCmdQueueWorker() {
}

func (v *controllerMock) retentionWorker() {
}

func (v *controllerMock) cmdQueueWorker() {
}

type validatorMock struct {
}

func (v *validatorMock) RequireServerOwner(recoveredID string, serverID string) error {
	return nil
}

func (v *validatorMock) RequireColonyOwner(recoveredID string, colonyID string) error {
	return nil
}

func (v *validatorMock) RequireExecutorMembership(recoveredID string, colonyID string, approved bool) error {
	return nil
}

func getTestGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = &http.Request{
		Header: make(http.Header),
	}

	return ctx, w
}

func assertRPCError(t *testing.T, body string) {
	rpcReplyMsg, err := rpc.CreateRPCReplyMsgFromJSON(body)
	assert.Nil(t, err)
	assert.True(t, rpcReplyMsg.Error)
}
