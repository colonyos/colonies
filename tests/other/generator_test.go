
func TestAddGeneratorCounter(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	colonyID := env.colonyID

	generator := utils.FakeGeneratorSingleProcess(t, colonyID)
	generator.Trigger = 10
	addedGenerator, err := client.AddGenerator(generator, env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedGenerator)

	go func() {
		for i := 0; i < 1000; i++ {
			err = client.PackGenerator(addedGenerator.ID, "arg"+strconv.Itoa(i), env.executorPrvKey)
			assert.Nil(t, err)
			time.Sleep(50 * time.Millisecond)
		}
	}()

	for {
		process, err := client.Assign(colonyID, 3, env.executorPrvKey)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("Args", process.ProcessSpec.Args)
		fmt.Println("AssignedExecutorID", process.AssignedExecutorID)
		fmt.Println("Closing ProcessID", process.ID)
		fmt.Println("My ExecutorID", env.executorID)
		err = client.Close(process.ID, env.executorPrvKey)
		assert.Nil(t, err)
		fmt.Println()
	}

	server.Shutdown()
	<-done
}
