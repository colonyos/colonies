package kvstore

import (
	"testing"
	"time"
)

// Test struct for KVStore functionality
type User struct {
	ColonyName string `json:"colonyname"`
	ID         string `json:"userid"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Active     bool   `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
}

func TestKVStore(t *testing.T) {
	// Create a KVStore for User objects
	kv := NewKVStore[*User]()

	user1 := &User{
		ColonyName: "colony1",
		ID:         "user123",
		Name:       "Alice Johnson",
		Email:      "alice@example.com",
		Active:     true,
		CreatedAt:  time.Now(),
	}

	// Test Put and Get with type safety
	err := kv.Put("users/alice", user1)
	if err != nil {
		t.Errorf("Failed to put user: %v", err)
	}

	retrievedUser, err := kv.Get("users/alice")
	if err != nil {
		t.Errorf("Failed to get user: %v", err)
	}

	// No type assertion needed! retrievedUser is already *User
	if retrievedUser.Name != "Alice Johnson" {
		t.Errorf("Expected 'Alice Johnson', got '%s'", retrievedUser.Name)
	}

	if retrievedUser.ColonyName != "colony1" {
		t.Errorf("Expected 'colony1', got '%s'", retrievedUser.ColonyName)
	}
}

func TestArrays(t *testing.T) {
	kv := NewKVStore[*User]()

	// Create array
	err := kv.CreateArray("users")
	if err != nil {
		t.Errorf("Failed to create array: %v", err)
	}

	user1 := &User{
		ColonyName: "colony1",
		ID:         "user123",
		Name:       "Alice Johnson",
		Email:      "alice@example.com",
		Active:     true,
	}

	user2 := &User{
		ColonyName: "colony2",
		ID:         "user456",
		Name:       "Bob Smith",
		Email:      "bob@example.com",
		Active:     false,
	}

	// Append to array with type safety
	err = kv.AppendToArray("users", user1)
	if err != nil {
		t.Errorf("Failed to append user1: %v", err)
	}

	err = kv.AppendToArray("users", user2)
	if err != nil {
		t.Errorf("Failed to append user2: %v", err)
	}

	// Access array elements with type safety
	firstUser, err := kv.Get("users/0/value")
	if err != nil {
		t.Errorf("Failed to get first user: %v", err)
	}

	// No type assertion needed!
	if firstUser.Name != "Alice Johnson" {
		t.Errorf("Expected 'Alice Johnson', got '%s'", firstUser.Name)
	}

	secondUser, err := kv.Get("users/1/value")
	if err != nil {
		t.Errorf("Failed to get second user: %v", err)
	}

	if secondUser.Name != "Bob Smith" {
		t.Errorf("Expected 'Bob Smith', got '%s'", secondUser.Name)
	}
}

func TestSearch(t *testing.T) {
	kv := NewKVStore[*User]()

	err := kv.CreateArray("users")
	if err != nil {
		t.Errorf("Failed to create array: %v", err)
	}

	users := []*User{
		{ColonyName: "colony1", ID: "user1", Name: "Alice", Email: "alice@example.com", Active: true},
		{ColonyName: "colony2", ID: "user2", Name: "Bob", Email: "bob@example.com", Active: false},
		{ColonyName: "colony1", ID: "user3", Name: "Charlie", Email: "charlie@example.com", Active: true},
	}

	for _, user := range users {
		err = kv.AppendToArray("users", user)
		if err != nil {
			t.Errorf("Failed to append user: %v", err)
		}
	}

	// Search by JSON field name with type safety
	results, err := kv.FindInArray("users", "colonyname", "colony1")
	if err != nil {
		t.Errorf("Failed to search by colonyname: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for colony1, got %d", len(results))
	}

	// No type assertion needed for results!
	for _, result := range results {
		if result.Value.ColonyName != "colony1" {
			t.Errorf("Expected colonyname 'colony1', got '%s'", result.Value.ColonyName)
		}
	}

	// Search by active status
	activeResults, err := kv.FindInArray("users", "active", true)
	if err != nil {
		t.Errorf("Failed to search by active: %v", err)
	}

	if len(activeResults) != 2 {
		t.Errorf("Expected 2 active users, got %d", len(activeResults))
	}

	// Verify type safety - all results are *User
	for _, result := range activeResults {
		if !result.Value.Active {
			t.Error("Found inactive user in active search results")
		}
	}
}

func TestFieldValue(t *testing.T) {
	kv := NewKVStore[*User]()

	user := &User{
		ColonyName: "test-colony",
		ID:         "test123",
		Name:       "Test User",
		Email:      "test@example.com",
		Active:     true,
	}

	err := kv.Put("current_user", user)
	if err != nil {
		t.Errorf("Failed to store user: %v", err)
	}

	// Get field value by JSON tag name
	name, err := kv.GetFieldValue("current_user", "name")
	if err != nil {
		t.Errorf("Failed to get name field: %v", err)
	}

	if name != "Test User" {
		t.Errorf("Expected 'Test User', got '%v'", name)
	}

	// Get field by JSON tag "colonyname"
	colonyName, err := kv.GetFieldValue("current_user", "colonyname")
	if err != nil {
		t.Errorf("Failed to get colonyname field: %v", err)
	}

	if colonyName != "test-colony" {
		t.Errorf("Expected 'test-colony', got '%v'", colonyName)
	}
}

// Test different types with generics
func TestWithDifferentTypes(t *testing.T) {
	// String store
	stringStore := NewKVStore[string]()
	
	err := stringStore.Put("config/host", "localhost")
	if err != nil {
		t.Errorf("Failed to put string: %v", err)
	}
	
	host, err := stringStore.Get("config/host")
	if err != nil {
		t.Errorf("Failed to get string: %v", err)
	}
	
	// host is automatically string type, no casting needed
	if host != "localhost" {
		t.Errorf("Expected 'localhost', got '%s'", host)
	}

	// Integer store  
	intStore := NewKVStore[int]()
	
	err = intStore.Put("config/port", 8080)
	if err != nil {
		t.Errorf("Failed to put int: %v", err)
	}
	
	port, err := intStore.Get("config/port")
	if err != nil {
		t.Errorf("Failed to get int: %v", err)
	}
	
	// port is automatically int type
	if port != 8080 {
		t.Errorf("Expected 8080, got %d", port)
	}

	// Struct store (non-pointer)
	type Config struct {
		Host string
		Port int
	}
	
	configStore := NewKVStore[Config]()
	
	config := Config{Host: "localhost", Port: 8080}
	err = configStore.Put("app_config", config)
	if err != nil {
		t.Errorf("Failed to put config: %v", err)
	}
	
	retrievedConfig, err := configStore.Get("app_config")
	if err != nil {
		t.Errorf("Failed to get config: %v", err)
	}
	
	// retrievedConfig is automatically Config type
	if retrievedConfig.Host != "localhost" {
		t.Errorf("Expected 'localhost', got '%s'", retrievedConfig.Host)
	}
}

func TestArraysWithDifferentTypes(t *testing.T) {
	// Array of strings
	stringArrayStore := NewKVStore[string]()
	
	err := stringArrayStore.CreateArray("languages")
	if err != nil {
		t.Errorf("Failed to create string array: %v", err)
	}
	
	languages := []string{"Go", "Rust", "Python"}
	for _, lang := range languages {
		err = stringArrayStore.AppendToArray("languages", lang)
		if err != nil {
			t.Errorf("Failed to append language: %v", err)
		}
	}
	
	// Access with type safety
	firstLang, err := stringArrayStore.Get("languages/0/value")
	if err != nil {
		t.Errorf("Failed to get first language: %v", err)
	}
	
	// firstLang is automatically string type
	if firstLang != "Go" {
		t.Errorf("Expected 'Go', got '%s'", firstLang)
	}

	// Array of integers
	intArrayStore := NewKVStore[int]()
	
	err = intArrayStore.CreateArray("numbers")
	if err != nil {
		t.Errorf("Failed to create int array: %v", err)
	}
	
	numbers := []int{1, 2, 3, 4, 5}
	for _, num := range numbers {
		err = intArrayStore.AppendToArray("numbers", num)
		if err != nil {
			t.Errorf("Failed to append number: %v", err)
		}
	}
	
	// Access with type safety
	secondNum, err := intArrayStore.Get("numbers/1/value")
	if err != nil {
		t.Errorf("Failed to get second number: %v", err)
	}
	
	// secondNum is automatically int type
	if secondNum != 2 {
		t.Errorf("Expected 2, got %d", secondNum)
	}
}

func TestNestedStructures(t *testing.T) {
	kv := NewKVStore[*User]()

	// Create nested structure
	err := kv.CreateArray("colonies")
	if err != nil {
		t.Errorf("Failed to create colonies array: %v", err)
	}

	// This would require mixed types, so for now we'll use a simpler approach
	// In practice, you might have different stores for different types
	
	userStore := NewKVStore[*User]()
	
	err = userStore.CreateArray("colony1/users")
	if err != nil {
		t.Errorf("Failed to create nested user array: %v", err)
	}

	user := &User{
		ColonyName: "colony1",
		ID:         "admin001",
		Name:       "Admin User",
		Email:      "admin@colony1.com",
		Active:     true,
	}

	err = userStore.AppendToArray("colony1/users", user)
	if err != nil {
		t.Errorf("Failed to add user to nested array: %v", err)
	}

	// Access nested user with full type safety
	retrievedUser, err := userStore.Get("colony1/users/0/value")
	if err != nil {
		t.Errorf("Failed to get nested user: %v", err)
	}

	// No type assertion needed!
	if retrievedUser.Name != "Admin User" {
		t.Errorf("Expected 'Admin User', got '%s'", retrievedUser.Name)
	}
}

// Test the restored original interface{} functionality
func TestMixedTypeCompatibility(t *testing.T) {
	// This restores the exact original functionality
	kv := NewMixedKVStore()

	// Test mixed-type storage like the original
	err := kv.Put("config/database/host", "localhost")
	if err != nil {
		t.Errorf("Failed to put string: %v", err)
	}

	err = kv.Put("config/database/port", 5432)
	if err != nil {
		t.Errorf("Failed to put int: %v", err)
	}

	// Retrieve with interface{} (like original)
	host, err := kv.Get("config/database/host")
	if err != nil {
		t.Errorf("Failed to get host: %v", err)
	}

	// Type assertion needed (like original)
	hostStr, ok := host.(string)
	if !ok {
		t.Error("Host is not a string")
	}
	if hostStr != "localhost" {
		t.Errorf("Expected 'localhost', got '%s'", hostStr)
	}

	port, err := kv.Get("config/database/port")
	if err != nil {
		t.Errorf("Failed to get port: %v", err)
	}

	portInt, ok := port.(int)
	if !ok {
		t.Error("Port is not an int")
	}
	if portInt != 5432 {
		t.Errorf("Expected 5432, got %d", portInt)
	}
}

func TestMixedArrayCompatibility(t *testing.T) {
	kv := NewMixedKVStore()

	// Create array with mixed-type functionality
	err := kv.CreateArray("users")
	if err != nil {
		t.Errorf("Failed to create array: %v", err)
	}

	// Store mixed-type data like the original (this was the key functionality!)
	err = kv.AppendToArrayMixed("users", map[string]interface{}{
		"name": "Alice Johnson",
		"age":  30,
		"city": "New York",
		"active": true,
	})
	if err != nil {
		t.Errorf("Failed to append mixed map: %v", err)
	}

	err = kv.AppendToArrayMixed("users", map[string]interface{}{
		"name": "Bob Smith", 
		"age":  25,
		"city": "London",
		"active": false,
	})
	if err != nil {
		t.Errorf("Failed to append second mixed map: %v", err)
	}

	// Test path-based access like the original
	name, err := kv.Get("users/0/name")
	if err != nil {
		t.Errorf("Failed to get users/0/name: %v", err)
	}

	nameStr, ok := name.(string)
	if !ok {
		t.Error("Name is not a string")
	}
	if nameStr != "Alice Johnson" {
		t.Errorf("Expected 'Alice Johnson', got '%s'", nameStr)
	}

	age, err := kv.Get("users/1/age")
	if err != nil {
		t.Errorf("Failed to get users/1/age: %v", err)
	}

	ageInt, ok := age.(int)
	if !ok {
		t.Error("Age is not an int")
	}
	if ageInt != 25 {
		t.Errorf("Expected 25, got %d", ageInt)
	}

	active, err := kv.Get("users/1/active")
	if err != nil {
		t.Errorf("Failed to get users/1/active: %v", err)
	}

	activeBool, ok := active.(bool)
	if !ok {
		t.Error("Active is not a bool")
	}
	if activeBool != false {
		t.Errorf("Expected false, got %v", activeBool)
	}
}

func TestOriginalGoObjectInMixedStore(t *testing.T) {
	kv := NewMixedKVStore()

	err := kv.CreateArray("system/users")
	if err != nil {
		t.Errorf("Failed to create array: %v", err)
	}

	// Store Go objects in mixed-type maps (original functionality)
	user := &User{
		ColonyName: "prod-colony",
		ID:         "admin001",
		Name:       "Admin User",
		Email:      "admin@prod.com",
		Active:     true,
	}

	err = kv.AppendToArrayMixed("system/users", map[string]interface{}{
		"user_object": user,
		"metadata": map[string]interface{}{
			"created_by": "system",
			"version":    1,
			"timestamp":  "2025-01-01T00:00:00Z",
		},
	})
	if err != nil {
		t.Errorf("Failed to append user object: %v", err)
	}

	// Access the Go object
	userObj, err := kv.Get("system/users/0/user_object")
	if err != nil {
		t.Errorf("Failed to get user object: %v", err)
	}

	retrievedUser, ok := userObj.(*User)
	if !ok {
		t.Error("User object is not *User")
	}

	if retrievedUser.Name != "Admin User" {
		t.Errorf("Expected 'Admin User', got '%s'", retrievedUser.Name)
	}

	// Access metadata
	version, err := kv.Get("system/users/0/metadata/version")
	if err != nil {
		t.Errorf("Failed to get metadata version: %v", err)
	}

	versionInt, ok := version.(int)
	if !ok {
		t.Error("Version is not an int")
	}
	if versionInt != 1 {
		t.Errorf("Expected version 1, got %d", versionInt)
	}
}

// Test recursive search from root
func TestRecursiveSearch(t *testing.T) {
	kv := NewKVStore[*User]()

	// Create complex nested structure
	// /departments/engineering/teams/backend/members (array)
	// /departments/engineering/teams/frontend/members (array) 
	// /departments/marketing/members (array)
	// /freelancers (array)

	// Create the structure
	err := kv.CreateArray("departments/engineering/teams/backend/members")
	if err != nil {
		t.Errorf("Failed to create backend members: %v", err)
	}

	err = kv.CreateArray("departments/engineering/teams/frontend/members")
	if err != nil {
		t.Errorf("Failed to create frontend members: %v", err)
	}

	err = kv.CreateArray("departments/marketing/members")
	if err != nil {
		t.Errorf("Failed to create marketing members: %v", err)
	}

	err = kv.CreateArray("freelancers")
	if err != nil {
		t.Errorf("Failed to create freelancers: %v", err)
	}

	// Add users across different locations
	backendDev1 := &User{
		ColonyName: "prod-colony",
		ID:         "backend001",
		Name:       "Alice Backend",
		Email:      "alice@company.com",
		Active:     true,
	}

	backendDev2 := &User{
		ColonyName: "staging-colony", 
		ID:         "backend002",
		Name:       "Bob Backend",
		Email:      "bob@company.com",
		Active:     true,
	}

	frontendDev := &User{
		ColonyName: "prod-colony",
		ID:         "frontend001", 
		Name:       "Charlie Frontend",
		Email:      "charlie@company.com",
		Active:     true,
	}

	marketingPerson := &User{
		ColonyName: "marketing-colony",
		ID:         "marketing001",
		Name:       "Diana Marketing", 
		Email:      "diana@company.com",
		Active:     false,
	}

	freelancer := &User{
		ColonyName: "prod-colony",
		ID:         "freelancer001",
		Name:       "Eve Freelancer",
		Email:      "eve@freelance.com", 
		Active:     true,
	}

	// Add users to different locations
	err = kv.AppendToArray("departments/engineering/teams/backend/members", backendDev1)
	if err != nil {
		t.Errorf("Failed to add backend dev1: %v", err)
	}

	err = kv.AppendToArray("departments/engineering/teams/backend/members", backendDev2)
	if err != nil {
		t.Errorf("Failed to add backend dev2: %v", err)
	}

	err = kv.AppendToArray("departments/engineering/teams/frontend/members", frontendDev)
	if err != nil {
		t.Errorf("Failed to add frontend dev: %v", err)
	}

	err = kv.AppendToArray("departments/marketing/members", marketingPerson)
	if err != nil {
		t.Errorf("Failed to add marketing person: %v", err)
	}

	err = kv.AppendToArray("freelancers", freelancer)
	if err != nil {
		t.Errorf("Failed to add freelancer: %v", err)
	}

	// Test 1: Search from root for all users with colonyname "prod-colony"
	prodResults, err := kv.FindRecursive("", "colonyname", "prod-colony")
	if err != nil {
		t.Errorf("Failed recursive search for prod-colony: %v", err)
	}

	if len(prodResults) != 3 {
		t.Errorf("Expected 3 prod-colony users, got %d", len(prodResults))
	}

	// Verify the paths are correct
	expectedPaths := map[string]bool{
		"departments/engineering/teams/backend/members/0/value": true,
		"departments/engineering/teams/frontend/members/0/value": true, 
		"freelancers/0/value": true,
	}

	for _, result := range prodResults {
		if !expectedPaths[result.Path] {
			t.Errorf("Unexpected path in results: %s", result.Path)
		}
		if result.Value.ColonyName != "prod-colony" {
			t.Errorf("Expected colonyname 'prod-colony', got '%s'", result.Value.ColonyName)
		}
	}

	// Test 2: Search from specific subtree for users with active=true
	activeResults, err := kv.FindRecursive("departments/engineering", "active", true)
	if err != nil {
		t.Errorf("Failed recursive search in engineering: %v", err)
	}

	if len(activeResults) != 3 {
		t.Errorf("Expected 3 active engineering users, got %d", len(activeResults))
	}

	// Test 3: Search for all users with "name" field (should find everyone)
	allWithName, err := kv.FindAllRecursive("", "name")
	if err != nil {
		t.Errorf("Failed to find all with name: %v", err)
	}

	if len(allWithName) != 5 {
		t.Errorf("Expected 5 users with name field, got %d", len(allWithName))
	}

	// Test 4: Search in non-existent subtree
	_, err = kv.FindRecursive("nonexistent/path", "name", "test")
	if err == nil {
		t.Error("Expected error for non-existent path, got nil")
	}
}

func TestRecursiveSearchMixedTypes(t *testing.T) {
	kv := NewMixedKVStore()

	// Create complex structure with mixed types
	err := kv.CreateArray("companies")
	if err != nil {
		t.Errorf("Failed to create companies: %v", err)
	}

	// Add company data with nested user arrays
	err = kv.AppendToArrayMixed("companies", map[string]interface{}{
		"name": "TechCorp",
		"industry": "Technology",
		"active": true,
	})
	if err != nil {
		t.Errorf("Failed to add TechCorp: %v", err)
	}

	err = kv.CreateArray("companies/0/employees")
	if err != nil {
		t.Errorf("Failed to create employees array: %v", err)
	}

	// Add employees with Go objects
	techUser1 := &User{
		ColonyName: "tech-colony",
		ID:         "tech001",
		Name:       "Tech Alice",
		Email:      "alice@techcorp.com",
		Active:     true,
	}

	techUser2 := &User{
		ColonyName: "tech-colony", 
		ID:         "tech002",
		Name:       "Tech Bob",
		Email:      "bob@techcorp.com",
		Active:     false,
	}

	err = kv.AppendToArrayMixed("companies/0/employees", map[string]interface{}{
		"user": techUser1,
		"role": "Senior Developer",
		"salary": 120000,
	})
	if err != nil {
		t.Errorf("Failed to add tech user 1: %v", err)
	}

	err = kv.AppendToArrayMixed("companies/0/employees", map[string]interface{}{
		"user": techUser2, 
		"role": "Junior Developer",
		"salary": 80000,
	})
	if err != nil {
		t.Errorf("Failed to add tech user 2: %v", err)
	}

	// Add another company
	err = kv.AppendToArrayMixed("companies", map[string]interface{}{
		"name": "StartupInc",
		"industry": "Fintech", 
		"active": false,
	})
	if err != nil {
		t.Errorf("Failed to add StartupInc: %v", err)
	}

	err = kv.CreateArray("companies/1/employees")
	if err != nil {
		t.Errorf("Failed to create startup employees: %v", err)
	}

	startupUser := &User{
		ColonyName: "startup-colony",
		ID:         "startup001", 
		Name:       "Startup Charlie",
		Email:      "charlie@startupinc.com",
		Active:     true,
	}

	err = kv.AppendToArrayMixed("companies/1/employees", map[string]interface{}{
		"user": startupUser,
		"role": "CTO",
		"equity": 5.0,
	})
	if err != nil {
		t.Errorf("Failed to add startup user: %v", err)
	}

	// Test recursive search for users by JSON tags
	techColonyResults, err := kv.FindRecursive("", "colonyname", "tech-colony")
	if err != nil {
		t.Errorf("Failed recursive search for tech-colony: %v", err)
	}

	if len(techColonyResults) != 2 {
		t.Errorf("Expected 2 tech-colony users, got %d", len(techColonyResults))
	}

	// Verify we found the right users
	for _, result := range techColonyResults {
		user, ok := result.Value.(*User)
		if !ok {
			t.Error("Result value is not a *User")
		}
		if user.ColonyName != "tech-colony" {
			t.Errorf("Expected colonyname 'tech-colony', got '%s'", user.ColonyName)
		}
	}

	// Test search for active users across the entire tree
	activeUserResults, err := kv.FindRecursive("", "active", true)
	if err != nil {
		t.Errorf("Failed recursive search for active users: %v", err)
	}

	if len(activeUserResults) != 2 {  // techUser1 and startupUser
		t.Errorf("Expected 2 active users, got %d", len(activeUserResults))
	}

	// Test search within specific company subtree
	companyResults, err := kv.FindRecursive("companies/0", "name", "Tech Alice")
	if err != nil {
		t.Errorf("Failed search in company subtree: %v", err)
	}

	if len(companyResults) != 1 {
		t.Errorf("Expected 1 result for Tech Alice, got %d", len(companyResults))
	}

	if companyResults[0].Path != "companies/0/employees/0/user" {
		t.Errorf("Expected path 'companies/0/employees/0/user', got '%s'", companyResults[0].Path)
	}
}

func TestRecursiveSearchEdgeCases(t *testing.T) {
	kv := NewKVStore[*User]()

	// Test search in empty store
	emptyResults, err := kv.FindRecursive("", "name", "test")
	if err != nil {
		t.Errorf("Failed search in empty store: %v", err)
	}

	if len(emptyResults) != 0 {
		t.Errorf("Expected 0 results in empty store, got %d", len(emptyResults))
	}

	// Add some data
	user := &User{Name: "Test User", ColonyName: "test-colony"}
	err = kv.Put("single_user", user)
	if err != nil {
		t.Errorf("Failed to put single user: %v", err)
	}

	// Test search from root
	rootResults, err := kv.FindRecursive("", "name", "Test User")
	if err != nil {
		t.Errorf("Failed search from root: %v", err)
	}

	if len(rootResults) != 1 {
		t.Errorf("Expected 1 result from root, got %d", len(rootResults))
	}

	if rootResults[0].Path != "single_user" {
		t.Errorf("Expected path 'single_user', got '%s'", rootResults[0].Path)
	}

	// Test search with root path variations
	slashResults, err := kv.FindRecursive("/", "name", "Test User")
	if err != nil {
		t.Errorf("Failed search with '/' path: %v", err)
	}

	if len(slashResults) != 1 {
		t.Errorf("Expected 1 result with '/' path, got %d", len(slashResults))
	}
}