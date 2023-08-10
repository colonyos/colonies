package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLabelEquals(t *testing.T) {
	label1 := &Label{Name: "test_labelname", Files: 10}
	label2 := &Label{Name: "test_labelname", Files: 10}

	assert.True(t, label1.Equals(label2))
	label1.Name = "changed_labelname"
	assert.False(t, label1.Equals(label2))
}

func TestLabelToJSON(t *testing.T) {
	label1 := &Label{Name: "test_labelname", Files: 10}

	jsonStr, err := label1.ToJSON()
	assert.Nil(t, err)

	label2, err := ConvertJSONToLabel(jsonStr)
	assert.Nil(t, err)
	assert.True(t, label1.Equals(label2))
}

func TestIsLabelsArraysEquals(t *testing.T) {
	label1 := &Label{Name: "test_labelname1", Files: 1}
	label2 := &Label{Name: "test_labelname2", Files: 10}
	label3 := &Label{Name: "test_labelname3", Files: 100}
	label4 := &Label{Name: "test_labelname4", Files: 1000}

	labels1 := []*Label{label1, label2}
	labels2 := []*Label{label3, label4}
	assert.True(t, IsLabelArraysEqual(labels1, labels1))
	assert.False(t, IsLabelArraysEqual(labels1, labels2))
}

func TestLabelArrayToJSON(t *testing.T) {
	label1 := &Label{Name: "test_labelname1", Files: 1}
	label2 := &Label{Name: "test_labelname2", Files: 10}
	labels1 := []*Label{label1, label2}

	jsonStr, err := ConvertLabelArrayToJSON(labels1)
	assert.Nil(t, err)

	labels2, err := ConvertJSONToLabelArray(jsonStr)
	assert.Nil(t, err)
	assert.True(t, IsLabelArraysEqual(labels1, labels2))
}
