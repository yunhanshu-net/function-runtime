// Package jsonx ...
package jsonx

import (
	"encoding/json"
	"fmt"
	"github.com/yunhanshu-net/runcher/pkg/logger"
)

// MustJSON ..
func MustJSON(el interface{}) string {
	marshal, err := json.Marshal(el)
	if err != nil {
		panic(err)
	}
	return string(marshal)
}

// MustPrintJSON ...
func MustPrintJSON(el interface{}) {
	marshal, err := json.Marshal(el)
	if err != nil {
		fmt.Println(fmt.Sprintf("[jsonx] err:%s el:%+v", err.Error(), el))
		return
	}
	fmt.Println(string(marshal))
}

// JSONString ...
func JSONString(el interface{}) string {
	marshal, err := json.Marshal(el)
	if err != nil {
		logger.Infof("[JSONString] err:%+v err:%s", el, err.Error())
		return ""
	}
	return string(marshal)
}

// String ...
func String(el interface{}) string {
	marshal, err := json.Marshal(el)
	if err != nil {
		return ""
	}
	return string(marshal)
}
