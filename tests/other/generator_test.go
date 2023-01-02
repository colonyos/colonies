
func TestAddGeneratorCounter(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator := utils.FakeGeneratorSingleProcess(t, colonyID)
	generator.Trigger = 10
	addedGenerator, err := client.AddGenerator(generator, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	go func() {
		for i := 0; i < 1000; i++ {
			err = client.PackGenerator(addedGenerator.ID, "arg"+strconv.Itoa(i), env.runtimePrvKey)
			assert.Nil(t, err)
			time.Sleep(50 * time.Millisecond)
		}
	}()

	for {
		process, err := client.AssignProcess(colonyID, 3, env.runtimePrvKey)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("Args", process.ProcessSpec.Args)
		fmt.Println("AssignedRuntimeID", process.AssignedRuntimeID)
		fmt.Println("Closing ProcessID", process.ID)
		fmt.Println("My RuntimeID", env.runtimeID)
		err = client.Close(process.ID, env.runtimePrvKey)
		assert.Nil(t, err)
		fmt.Println()
	}

	server.Shutdown()
	<-done
}
