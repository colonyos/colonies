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

// Channel Rate Limiting - Token bucket configuration for channel message rate limiting
const CHANNEL_RATE_LIMIT_MESSAGES_PER_SECOND = 100.0 // Maximum sustained message rate per process
const CHANNEL_RATE_LIMIT_BURST_SIZE = 500            // Maximum burst size (token bucket capacity)

// Channel Message Size - Maximum payload size for channel messages
const CHANNEL_MAX_MESSAGE_SIZE = 10 * 1024 * 1024 // 10 MB - allows large payloads for database results, file transfers, etc.

// Channel Subscriber Buffer - Buffer size for push notification channels
const CHANNEL_SUBSCRIBER_BUFFER_SIZE = 10000 // Number of messages buffered per subscriber before disconnection

// Channel Log Size - Maximum number of entries per channel log
const CHANNEL_MAX_LOG_ENTRIES = 10000 // Maximum entries per channel, oldest removed when exceeded

// Channel Limit - Maximum number of channels per process
const CHANNEL_MAX_CHANNELS_PER_PROCESS = 100 // Maximum channels a single process can have