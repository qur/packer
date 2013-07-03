package kvm

import (
	"net"
	"path/filepath"
	"encoding/json"
	"log"
	"bytes"
	"fmt"
)

type kvmControl struct {
	path     string
	c        net.Conn
	info     *qmpInfo
	qmp      chan *qmpInfo
	response chan map[string]interface{}
	active   bool
}

type notRunning struct {
	k *kvmControl
}

func (n *notRunning) Error() string {
	return fmt.Sprintf("'%s' does appear to be connected to a running KVM instance", n.k.path)
}

type versionNum struct {
	Major, Minor, Micro int
}

type qmpVersion struct {
	Qemu    versionNum
	Package string
}

type qmpInfo struct {
	Version      qmpVersion
	Capabilities []string
}

type execute struct {
	Execute   string        `json:"execute"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

type timestamp struct {
	Seconds, Microseconds int
}

type response struct {
	QMP       *qmpInfo               `json:"QMP,omitempty"`
	Return    map[string]interface{} `json:"return,omitempty"`
	Event     string                 `json:"event,omitempty"`
	Timestamp *timestamp             `json:"timestamp,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

func NewKvmControl(kvmPath string) (*kvmControl, error) {
	path := filepath.Join(filepath.Dir(kvmPath), "control")
	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}
	k := &kvmControl{
		path:     path,
		c:        conn,
		qmp:      make(chan *qmpInfo),
		response: make(chan map[string]interface{}),
		active:   true,
	}
	go k.receiver()
	k.info = <-k.qmp
	_, err = k.Execute("qmp_capabilities")
	if err != nil {
		return nil, err
	}
	return k, nil
}

func (k *kvmControl) Close() error {
	k.active = false
	return k.c.Close()
}

func (k *kvmControl) send(val interface{}) error {
	encoded, err := json.Marshal(val)
	if err != nil {
		return err
	}
	_, err = k.c.Write(encoded)
	if err != nil {
		return err
	}
	return nil
}

func (k *kvmControl) receiver() {
	buf := make([]byte, 8192)
	s := 0
	for k.active {
		n, err := k.c.Read(buf[s:])
		if !k.active {
			log.Printf("receiver: !active\n")
			break
		}
		if err != nil {
			log.Printf("receiver: %s\n", err)
			break
		}
		parts := bytes.Split(buf[:s+n], []byte("\n"))
		for _, part := range parts[:len(parts)-1] {
			k.handleMsg(part)
		}
		finalPart := parts[len(parts)-1]
		s = len(finalPart)
		if s > 0 {
			copy(buf, finalPart)
		}
	}
	// Someone may be waiting for a response ...
	select {
	case k.response <- nil:
	default:
	}
}

func (k *kvmControl) handleMsg(msg []byte) {
	resp := &response{}
	err := json.Unmarshal(msg, resp)
	if err != nil {
		log.Printf("handleMsg: %s\n", err)
		return
	}
	if resp.QMP != nil {
		log.Printf("QMP: %+v\n", *resp.QMP)
		k.qmp <- resp.QMP
		return
	}
	if len(resp.Event) > 0 {
		log.Printf("event: %+v\n", resp)
		return
	}
	k.response <- resp.Return
}

func (k *kvmControl) Execute(command string, args ...interface{}) (map[string]interface{}, error) {
	if !k.active {
		return nil, &notRunning{k}
	}

	msg := &execute{
		Execute:   command,
		Arguments: args,
	}
	err := k.send(msg)
	if err != nil {
		return nil, err
	}

	resp := <-k.response
	if resp == nil {
		return nil, &notRunning{k}
	}

	return resp, nil
}
