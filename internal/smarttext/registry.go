package smarttext

import "sync"

type Registry struct {
	mu      sync.RWMutex
	entries map[string]string
}

func NewRegistry() *Registry {
	return &Registry{
		entries: defaultEntries(),
	}
}

func NewRegistryWithOverrides(overrides map[string]string) *Registry {
	r := &Registry{
		entries: defaultEntries(),
	}
	for k, v := range overrides {
		r.entries[k] = v
	}
	return r
}

func (r *Registry) Get(name string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.entries[name]
}

func (r *Registry) Set(name, description string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[name] = description
}

func defaultEntries() map[string]string {
	return map[string]string{
		"count":    "Number of items",
		"name":     "The name",
		"age":      "The age",
		"id":       "The identifier",
		"path":     "The file path",
		"key":      "The key",
		"value":    "The value",
		"index":    "The index",
		"size":     "The size",
		"len":      "The length",
		"length":   "The length",
		"config":   "The configuration",
		"message":  "The message",
		"msg":      "The message",
		"err":      "An error",
		"error":    "An error",
		"ctx":      "The context",
		"addr":     "The address",
		"address":  "The address",
		"port":     "The port",
		"host":     "The host",
		"timeout":  "The timeout duration",
		"time":     "The time",
		"date":     "The date",
		"dir":      "The directory path",
		"file":     "The file path",
		"src":      "The source",
		"dest":     "The destination",
		"target":   "The target",
		"source":   "The source",
		"data":     "The data",
		"mode":     "The mode",
		"flag":     "The flag",
		"flags":    "The flags",
		"args":     "The arguments",
		"opts":     "The options",
		"options":  "The options",
		"limit":    "The limit",
		"max":      "The maximum value",
		"min":      "The minimum value",
		"total":    "The total",
		"sum":      "The sum",
	}
}
