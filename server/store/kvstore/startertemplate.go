package kvstore

import (
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/pkg/errors"
)

// Client provides an implementation of the KVStore interface using the pluginapi.
type Client struct {
	client *pluginapi.Client
}

// NewKVStore creates a new KVStore backed by the given pluginapi.Client.
func NewKVStore(client *pluginapi.Client) KVStore {
	return Client{
		client: client,
	}
}

// GetTemplateData is a sample method to get a key-value pair from the KV store.
func (kv Client) GetTemplateData(userID string) (string, error) {
	var templateData string
	err := kv.client.KV.Get("template_key-"+userID, &templateData)
	if err != nil {
		return "", errors.Wrap(err, "failed to get template data")
	}
	return templateData, nil
}

// Get retrieves the value associated with the given key.
// It assumes that values are stored as strings.
func (c Client) Get(key string) (interface{}, error) {
	var value string
	err := c.client.KV.Get(key, &value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get value")
	}
	return value, nil
}

// Set stores the provided string value under the given key.
func (c Client) Set(key string, value string) error {
	if _, err := c.client.KV.Set(key, value); err != nil {
		return errors.Wrap(err, "failed to set value")
	}
	return nil
}
