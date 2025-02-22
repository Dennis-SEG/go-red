package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Config represents the application configuration
type Config struct {
	values map[string]interface{}
	mu     sync.RWMutex
}

// New creates a new Config instance
func New() *Config {
	return &Config{
		values: make(map[string]interface{}),
	}
}

// LoadFromFile loads configuration from a JSON file
func (c *Config) LoadFromFile(filePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var values map[string]interface{}
	if err := json.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Flatten nested config
	c.values = flattenMap(values, "")

	return nil
}

// SaveToFile saves configuration to a JSON file
func (c *Config) SaveToFile(filePath string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Unflatten the config for saving
	nestedValues := unflattenMap(c.values)

	data, err := json.MarshalIndent(nestedValues, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Set sets a configuration value
func (c *Config) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

// Get gets a configuration value
func (c *Config) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.values[key]
	return value, exists
}

// GetString gets a string configuration value
func (c *Config) GetString(key string) string {
	value, exists := c.Get(key)
	if !exists {
		return ""
	}

	strValue, ok := value.(string)
	if !ok {
		// Try to convert to string
		return fmt.Sprintf("%v", value)
	}

	return strValue
}

// GetInt gets an integer configuration value
func (c *Config) GetInt(key string) int {
	value, exists := c.Get(key)
	if !exists {
		return 0
	}

	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		intValue, err := strconv.Atoi(v)
		if err != nil {
			return 0
		}
		return intValue
	default:
		return 0
	}
}

// GetBool gets a boolean configuration value
func (c *Config) GetBool(key string) bool {
	value, exists := c.Get(key)
	if !exists {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		boolValue, err := strconv.ParseBool(v)
		if err != nil {
			return false
		}
		return boolValue
	case int:
		return v != 0
	case float64:
		return v != 0
	default:
		return false
	}
}

// GetFloat gets a float configuration value
func (c *Config) GetFloat(key string) float64 {
	value, exists := c.Get(key)
	if !exists {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		floatValue, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return floatValue
	default:
		return 0
	}
}

// SetDefault sets a default value if the key doesn't exist
func (c *Config) SetDefault(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.values[key]; !exists {
		c.values[key] = value
	}
}

// Delete removes a configuration value
func (c *Config) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.values, key)
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		if !strings.HasPrefix(key, prefix) {
			continue
		}

		// Remove prefix and convert to lowercase
		configKey := strings.ToLower(strings.TrimPrefix(key, prefix))
		// Replace underscores with dots for nested keys
		configKey = strings.ReplaceAll(configKey, "_", ".")

		c.values[configKey] = value
	}
}

// flattenMap converts a nested map to a flat map with dot-separated keys
func flattenMap(nested map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range nested {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		switch child := v.(type) {
		case map[string]interface{}:
			childMap := flattenMap(child, key)
			for ck, cv := range childMap {
				result[ck] = cv
			}
		default:
			result[key] = v
		}
	}

	return result
}

// unflattenMap converts a flat map with dot-separated keys to a nested map
func unflattenMap(flat map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range flat {
		parts := strings.Split(k, ".")
		current := result

		for i, part := range parts {
			if i == len(parts)-1 {
				// Last part, set the value
				current[part] = v
			} else {
				// Create nested map if it doesn't exist
				if _, exists := current[part]; !exists {
					current[part] = make(map[string]interface{})
				}

				// Move to the next level
				current = current[part].(map[string]interface{})
			}
		}
	}

	return result
}
