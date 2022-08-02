package server

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAddAttributesErr(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	server := &ColoniesServer{}
	key := "test_key"
	value := "test_value"
	attribute := core.CreateAttribute(core.GenerateRandomID(), core.GenerateRandomID(), "", core.OUT, key, value)
	msg := rpc.CreateAddAttributeMsg(attribute)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	server.handleAddAttributeHTTPRequest(c, "", msg.MsgType, jsonString+"error")

	b, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	verifyRPCReplyMsgHasErr(t, b)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	msg2 := rpc.CreateAddAttributeMsg(attribute)
	msg2.MsgType = "invalid_msg_type"
	jsonString, err = msg2.ToJSON()
	server.handleAddAttributeHTTPRequest(c, "", msg.MsgType, jsonString)

	b, err = ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	verifyRPCReplyMsgHasErr(t, b)
}

func TestGetAttributesErr(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	server := &ColoniesServer{}
	msg := rpc.CreateGetAttributeMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	server.handleAddAttributeHTTPRequest(c, "", msg.MsgType, jsonString+"error")

	b, err := ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	verifyRPCReplyMsgHasErr(t, b)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	msg2 := rpc.CreateGetAttributeMsg(core.GenerateRandomID())
	msg2.MsgType = "invalid_msg_type"
	jsonString, err = msg2.ToJSON()
	server.handleAddAttributeHTTPRequest(c, "", msg.MsgType, jsonString)

	b, err = ioutil.ReadAll(w.Body)
	assert.Nil(t, err)
	verifyRPCReplyMsgHasErr(t, b)
}
