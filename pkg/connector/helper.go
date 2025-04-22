package connector

import "encoding/json"

func unmarshalProperties(properties interface{}) (map[string]interface{}, error) {
	rawBytes, err := json.Marshal(properties)
	if err != nil {
		return nil, err
	}

	var propsMap map[string]interface{}
	err = json.Unmarshal(rawBytes, &propsMap)
	if err != nil {
		return nil, err
	}

	return propsMap, nil
}
