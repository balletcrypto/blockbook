package server

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strings"
)

func PrintRequestParams(r *http.Request) {

	requestFullPath := r.URL.Path
	requestMethod := r.Method

	if requestFullPath == "/" && requestMethod == "HEAD" {
		// avoid print redundant log
		return
	}

	if queryValues := r.URL.Query(); len(queryValues) > 0 {
		// GET
		paramsArr := make([]string, 0, len(queryValues))
		for k, v := range queryValues {
			paramsArr = append(paramsArr, fmt.Sprintf("%s=%s", k, v[0]))
		}
		requestFullPath += "?" + strings.Join(paramsArr, "&")
	}

	logMessage := fmt.Sprintf("%s %s ",
		requestMethod,
		requestFullPath,
	)

	if body, _ := ioutil.ReadAll(r.Body); len(body) > 0 {
		logMessage += "\nRequest Params: "
		logMessage += string(body)
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}

	glog.Info(logMessage)

}
