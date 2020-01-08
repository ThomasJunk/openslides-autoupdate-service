package redis

import (
	"errors"
	"fmt"

	"github.com/openslides/openslides-autoupdate-service/internal/autoupdate"
)

var errNil = errors.New("nil returned")

// stream parses a redis stream objekt an autoupdate.KeyChanges objekt
func stream(reply interface{}, err error) (string, autoupdate.KeyChanges, error) {
	kc := autoupdate.KeyChanges{}
	if err != nil {
		return "", kc, err
	}
	if reply == nil {
		return "", kc, errNil
	}
	updatedSet := make(map[string]bool)
	streams, ok := reply.([]interface{})
	if !ok {
		return "", kc, fmt.Errorf("Invalid input. Data has to be a list, not %T", reply)
	}
	if len(streams) == 0 {
		return "", kc, fmt.Errorf("Invalid input. No stream in data")
	}
	stream1, ok := streams[0].([]interface{})
	if !ok {
		return "", kc, fmt.Errorf("Invalid input. Stream has to be a two-tuple, not %T", streams[0])
	}
	if len(stream1) != 2 {
		return "", kc, fmt.Errorf("Invalid input. Stream has to be a two-tuple, got %d elements", len(stream1))
	}
	data, ok := stream1[1].([]interface{})
	if !ok {
		return "", kc, fmt.Errorf("Invalid input. Stream data has to be a list, got %T", stream1[1])
	}
	var id string
	for _, v := range data {
		element, ok := v.([]interface{})
		if !ok {
			return "", kc, fmt.Errorf("Invalid input. Stream element has to be a two-tuple, got %T", v)
		}
		if len(element) != 2 {
			return "", kc, fmt.Errorf("Invalid input. Stream element has to be a two-tuple, got %d elements", len(element))
		}
		id, ok = tostr(element[0])
		if !ok {
			return "", kc, fmt.Errorf("Invalid input. Stream ID has to be a string, got %T", element[0])
		}
		kv, ok := element[1].([]interface{})
		if !ok {
			return "", kc, fmt.Errorf("Invalid input. Key values has to be a list of strings, got %T", element[1])
		}
		if len(kv)%2 != 0 {
			return "", kc, fmt.Errorf("Invalid input. Odd number of key value pairs")
		}
		for i := 0; i < len(kv)-1; i += 2 {
			key, ok := tostr(kv[i])
			if !ok {
				return "", kc, fmt.Errorf("Invalid input. Key has to be a string, got %T", kv[i])
			}
			value, ok := tostr(kv[i+1])
			if !ok {
				return "", kc, fmt.Errorf("Invalid input. Values has to be a string, got %T", kv[i+1])
			}
			switch key {
			case "updated":
				if !updatedSet[value] {
					kc.Updated = append(kc.Updated, value)
					updatedSet[value] = true
				}
			default:
				return "", kc, fmt.Errorf("Invalid input. Unknown key \"%s\"", key)
			}
		}
	}
	return id, kc, nil
}

// tostr converts an interface with value string or []byte to string
// this is an helper, because the test-code generates strings but the
// redis code generates []bytes
func tostr(i interface{}) (string, bool) {
	switch rid := i.(type) {
	case string:
		return rid, true
	case []byte:
		return string(rid), true
	default:
		return "", false
	}
}