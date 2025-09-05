// Package constants defines system-wide constants used throughout the ColonyOS server
package constants

// API Limits - Maximum values for API requests to prevent abuse
const MAX_COUNT = 100         // Maximum number of items that can be requested in list operations
const MAX_DAYS = 30           // Maximum number of days for log search operations
const MAX_LOG_COUNT = 500     // Maximum number of log entries that can be requested at once

// Test Configuration - Default values used in test environments
const TESTHOST = "localhost" // Default hostname for test servers
const TESTPORT = 28088       // Default port for test servers

// Background Processing Periods - How frequently various system tasks run
const RELEASE_PERIOD = 1              // Period in seconds when processes are checked for max exec time or max wait time
const GENERATOR_TRIGGER_PERIOD = 1000 // Period in milliseconds when generators are evaluated and triggered
const CRON_TRIGGER_PERIOD = 1000      // Period in milliseconds when cron jobs are evaluated and triggered

// Process Priority Limits - Valid range for process priority values
const MIN_PRIORITY = -50000 // Minimum allowed priority for processes (lower numbers = lower priority)
const MAX_PRIORITY = 50000  // Maximum allowed priority for processes (higher numbers = higher priority)