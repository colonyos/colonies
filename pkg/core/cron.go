package core

import (
	"encoding/json"
	"time"
)

type Cron struct {
	ID                 string    `json:"cronid"`
	ColonyID           string    `json:"colonyid"`
	Name               string    `json:"name"`
	CronExpression     string    `json:"cronexpression"`
	Intervall          int       `json:"intervall"`
	Random             bool      `json:"random"`
	NextRun            time.Time `json:"nextrun"`
	LastRun            time.Time `json:"lastrun"`
	WorkflowSpec       string    `json:"workflowspec"`
	LastProcessGraphID string    `json:"lastprocessgraphid"`
}

func CreateCron(colonyID string, name string, cronExpression string, intervall int, random bool, workflowSpec string) *Cron {
	return &Cron{ColonyID: colonyID, Name: name, CronExpression: cronExpression, Intervall: intervall, Random: random, NextRun: time.Time{}, LastRun: time.Time{}, WorkflowSpec: workflowSpec}
}

func ConvertJSONToCron(jsonString string) (*Cron, error) {
	var cron *Cron
	err := json.Unmarshal([]byte(jsonString), &cron)
	if err != nil {
		return nil, err
	}

	return cron, nil
}

func ConvertCronArrayToJSON(crons []*Cron) (string, error) {
	jsonBytes, err := json.MarshalIndent(crons, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func ConvertJSONToCronArray(jsonString string) ([]*Cron, error) {
	var crons []*Cron
	err := json.Unmarshal([]byte(jsonString), &crons)
	if err != nil {
		return crons, err
	}

	return crons, nil
}

func IsCronArraysEqual(crons1 []*Cron, crons2 []*Cron) bool {
	if crons1 == nil || crons2 == nil {
		return false
	}

	counter := 0
	for _, cron1 := range crons1 {
		for _, cron2 := range crons2 {
			if cron1.Equals(cron2) {
				counter++
			}
		}
	}

	if counter == len(crons1) && counter == len(crons2) {
		return true
	}

	return false
}

func (cron *Cron) Equals(cron2 *Cron) bool {
	if cron2 == nil {
		return false
	}

	same := true
	if cron.ID != cron2.ID ||
		cron.ColonyID != cron2.ColonyID ||
		cron.Name != cron2.Name ||
		cron.CronExpression != cron2.CronExpression ||
		cron.Intervall != cron2.Intervall ||
		cron.Random != cron2.Random ||
		cron.NextRun.Unix() != cron2.NextRun.Unix() ||
		cron.LastRun.Unix() != cron2.LastRun.Unix() ||
		cron.WorkflowSpec != cron2.WorkflowSpec ||
		cron.LastProcessGraphID != cron2.LastProcessGraphID {
		same = false
	}

	return same
}

func (cron *Cron) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(cron, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (cron *Cron) HasExpired() bool {
	now := time.Now()
	if now.Sub(cron.NextRun) > 0 {
		return true
	}
	return false
}
