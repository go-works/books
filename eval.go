package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type Eval struct {
	Files    []File `json:"files"`
	Language string `json:"lang"`
	// TODO: implement me
	Command string `json:"command,omitempty"`
	Stdin   string `json:"stdin,omitempty"`
}

type EvalResponse struct {
	Stdout     string  `json:"stdout"`
	Stderr     string  `json:"stderr"`
	Error      string  `json:"error,omitempty"`
	DurationMS float64 `json:"durationms"`
}

func evalGo(e *Eval) (*EvalResponse, error) {
	d, err := json.Marshal(e)
	must(err)
	uri := "http://localhost:8533/eval"
	body := bytes.NewReader(d)
	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed with '%s'", rsp.Status)
	}
	defer rsp.Body.Close()
	d, err = ioutil.ReadAll(rsp.Body)
	must(err)
	var res EvalResponse
	err = json.Unmarshal(d, &res)
	must(err)
	return &res, nil
}

func testEval(s string) {

}
