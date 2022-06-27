package server

import "testing"

func TestAddGenerator(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)
	server.Shutdown()
	<-done
}
