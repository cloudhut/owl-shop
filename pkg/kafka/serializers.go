package kafka

import (
	"encoding/json"
)

func SerializeJson(data interface{}) ([]byte, error) {
	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return serialized, nil
}
