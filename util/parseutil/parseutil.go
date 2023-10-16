package parseutil

import (
	"github.com/tidwall/gjson"

	"egreg10us/faultylauncher/util/downloadutil"
)

func ParseJSON(json string,remote bool) (gjson.Result,error) {
	if remote {
		jsonBytes,err := downloadutil.GetJSON(json)
		return gjson.Parse(string(jsonBytes)),err
	} else {
		return gjson.Parse(json),nil
	}
}
