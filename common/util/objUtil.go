package util

import "encoding/json"

type objUtil struct {
}

var ObjUtil = objUtil{}

func (objUtil) Obj2map(obj interface{}) (map[string]interface{}, error) {
	jsonByte, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err = json.Unmarshal(jsonByte, &result); err != nil {
		return nil, err
	}
	return result, nil
}
