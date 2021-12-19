package core

import (
	"colonies/pkg/crypto"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	WAITING int = 0
	RUNNING     = 1
	SUCCESS     = 2
	FAILED      = 3
)

type ProcessJSON struct {
	ID                 string           `json:"processid"`
	TargetColonyID     string           `json:"targetcolonyid"`
	TargetComputerIDs  []string         `json:"targetcomputerids"`
	AssignedComputerID string           `json:"assignedcomputerid"`
	Status             int              `json:"status"`
	IsAssigned         bool             `json:"isassigned"`
	ComputerType       string           `json:"computertype"`
	SubmissionTime     time.Time        `json:"submissiontime"`
	StartTime          time.Time        `json:"starttime"`
	EndTime            time.Time        `json:"endtime"`
	Deadline           time.Time        `json:"deadline"`
	Timeout            int              `json:"timeout"`
	Retries            int              `json:"retries"`
	MaxRetries         int              `json:"maxretries"`
	Log                string           `json:"log"`
	Mem                int              `json:"mem"`
	Cores              int              `json:"cores"`
	GPUs               int              `json:"gpus"`
	InAttributes       []*AttributeJSON `json:"in"`
	OutAttributes      []*AttributeJSON `json:"out"`
	ErrAttributes      []*AttributeJSON `json:"err"`
}

type Process struct {
	id                 string
	targetColonyID     string
	targetComputerIDs  []string
	assignedComputerID string
	status             int
	isAssigned         bool
	computerType       string
	submissionTime     time.Time
	startTime          time.Time
	endTime            time.Time
	deadline           time.Time
	timeout            int
	retries            int
	maxRetries         int
	log                string
	mem                int
	cores              int
	gpus               int
}

func CreateProcess(targetColonyID string, targetComputerIDs []string, computerType string, timeout int, maxRetries int, mem int, cores int, gpus int) *Process {
	uuid := uuid.New()
	id := crypto.GenerateHashFromString(uuid.String()).String()

	process := &Process{id: id,
		targetColonyID:    targetColonyID,
		targetComputerIDs: targetComputerIDs,
		status:            WAITING,
		isAssigned:        false,
		computerType:      computerType,
		timeout:           timeout,
		maxRetries:        maxRetries,
		mem:               mem,
		cores:             cores,
		gpus:              gpus}

	return process
}

func CreateProcessFromDB(id string, targetColonyID string, targetComputerIDs []string, assignedComputerID string, status int, isAssigned bool, computerType string, submissionTime time.Time, startTime time.Time, endTime time.Time, deadline time.Time, timeout int, retries int, maxRetries int, log string, mem int, cores int, gpus int) *Process {
	return &Process{id: id,
		targetColonyID:     targetColonyID,
		targetComputerIDs:  targetComputerIDs,
		assignedComputerID: assignedComputerID,
		status:             status,
		isAssigned:         isAssigned,
		computerType:       computerType,
		submissionTime:     submissionTime,
		startTime:          startTime,
		endTime:            endTime,
		deadline:           deadline,
		timeout:            timeout,
		retries:            retries,
		maxRetries:         maxRetries,
		log:                log,
		mem:                mem,
		cores:              cores,
		gpus:               gpus}
}

func CreateFromJSON(jsonString string) (*Process, []*Attribute, []*Attribute, []*Attribute, error) {
	var processJSON ProcessJSON
	var inAttributes []*Attribute
	var errAttributes []*Attribute
	var outAttributes []*Attribute

	err := json.Unmarshal([]byte(jsonString), &processJSON)
	if err != nil {
		return nil, inAttributes, errAttributes, outAttributes, err
	}

	inAttributes = convertFromAttributeJSON(processJSON.InAttributes)
	errAttributes = convertFromAttributeJSON(processJSON.ErrAttributes)
	outAttributes = convertFromAttributeJSON(processJSON.OutAttributes)

	process := &Process{id: processJSON.ID,
		targetColonyID:     processJSON.TargetColonyID,
		targetComputerIDs:  processJSON.TargetComputerIDs,
		assignedComputerID: processJSON.AssignedComputerID,
		status:             processJSON.Status,
		isAssigned:         processJSON.IsAssigned,
		computerType:       processJSON.ComputerType,
		submissionTime:     processJSON.SubmissionTime,
		startTime:          processJSON.StartTime,
		endTime:            processJSON.EndTime,
		deadline:           processJSON.Deadline,
		timeout:            processJSON.Timeout,
		retries:            processJSON.Retries,
		maxRetries:         processJSON.MaxRetries,
		log:                processJSON.Log,
		mem:                processJSON.Mem,
		cores:              processJSON.Cores,
		gpus:               processJSON.GPUs}

	return process, inAttributes, errAttributes, outAttributes, nil
}

func (process *Process) ID() string {
	return process.id
}

func (process *Process) TargetColonyID() string {
	return process.targetColonyID
}

func (process *Process) TargetComputerIDs() []string {
	return process.targetComputerIDs
}

func (process *Process) AssignedComputerID() string {
	return process.assignedComputerID
}

func (process *Process) SetAssignedComputerID(computerID string) {
	process.assignedComputerID = computerID
}

func (process *Process) Status() int {
	return process.status
}

func (process *Process) SetStatus(status int) {
	process.status = status
}

func (process *Process) ComputerType() string {
	return process.computerType
}

func (process *Process) SubmissionTime() time.Time {
	return process.submissionTime
}
func (process *Process) SetSubmissionTime(submissionTime time.Time) {
	process.submissionTime = submissionTime
}

func (process *Process) StartTime() time.Time {
	return process.startTime
}

func (process *Process) SetStartTime(startTime time.Time) {
	process.startTime = startTime
}

func (process *Process) EndTime() time.Time {
	return process.endTime
}

func (process *Process) SetEndTime(endTime time.Time) {
	process.endTime = endTime
}

func (process *Process) Deadline() time.Time {
	return process.deadline
}

func (process *Process) Timeout() int {
	return process.timeout
}

func (process *Process) Retries() int {
	return process.retries
}

func (process *Process) MaxRetries() int {
	return process.maxRetries
}

func (process *Process) Log() string {
	return process.log
}

func (process *Process) Mem() int {
	return process.mem
}

func (process *Process) Cores() int {
	return process.cores
}

func (process *Process) GPUs() int {
	return process.gpus
}

func (process *Process) Assigned() bool {
	return process.isAssigned
}

func (process *Process) Assign() {
	process.isAssigned = true
}

func (process *Process) Unassign() {
	process.isAssigned = false
}

func (process *Process) WaitingTime() time.Duration {
	return process.StartTime().Sub(process.SubmissionTime())
}

func (process *Process) ProcessingTime() time.Duration {
	return process.EndTime().Sub(process.StartTime())
}

func (process *Process) ToJSON(inAttributes []*Attribute, errAttributes []*Attribute, outAttributes []*Attribute) (string, error) {
	inAttributesJSON := convertToAttributeJSON(inAttributes)
	errAttributesJSON := convertToAttributeJSON(errAttributes)
	outAttributesJSON := convertToAttributeJSON(outAttributes)

	processJSON := &ProcessJSON{ID: process.id,
		TargetColonyID:     process.targetColonyID,
		TargetComputerIDs:  process.targetComputerIDs,
		AssignedComputerID: process.assignedComputerID,
		Status:             process.status,
		IsAssigned:         process.isAssigned,
		ComputerType:       process.computerType,
		SubmissionTime:     process.submissionTime,
		StartTime:          process.startTime,
		EndTime:            process.endTime,
		Deadline:           process.deadline,
		Timeout:            process.timeout,
		Retries:            process.retries,
		MaxRetries:         process.maxRetries,
		Log:                process.log,
		Mem:                process.mem,
		Cores:              process.cores,
		GPUs:               process.gpus,
		InAttributes:       inAttributesJSON,
		ErrAttributes:      errAttributesJSON,
		OutAttributes:      outAttributesJSON}

	jsonString, err := json.Marshal(processJSON)
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
