package main

import (
	"os/exec"
	"strconv"
	"syscall"
)

type ChromeInstance struct {
	uuid string
	port int
	cmd  *exec.Cmd
}

//Start chrome met een UUID en return instance

func NewChromeInstance(port int, uuid string) *ChromeInstance {
	cmd := exec.Command("C:/Program Files (x86)/Google/Chrome/Application/chrome.exe")
	//cmd := exec.Command("google-chrome")

	cmd.Args = append(cmd.Args, "--no-first-run")
	cmd.Args = append(cmd.Args, "--no-default-browser-check")
	cmd.Args = append(cmd.Args, "--incognito")
	cmd.Args = append(cmd.Args, "--user-data-dir=C:/Users/daniel/AppData/Local/Temp/"+uuid)
	//cmd.Args = append(cmd.Args, "--user-data-dir=/tmp/"+uuid)
	cmd.Args = append(cmd.Args, "--remote-debugging-port="+strconv.Itoa(port))
	cmd.Args = append(cmd.Args, "about:blank")

	//cmd.Dir = "/usr/bin/"

	return &ChromeInstance{
		uuid: uuid,
		port: port,
		cmd:  cmd,
	}
}

func (c *ChromeInstance) Start() error {
	return c.cmd.Start()
}

func (c *ChromeInstance) Kill() {
	c.cmd.Process.Signal(syscall.SIGKILL)
}

func (c *ChromeInstance) Terminate() {
	c.cmd.Process.Signal(syscall.SIGTERM)
}

func (c *ChromeInstance) GetUUID() string {
	return c.uuid
}

func (c *ChromeInstance) GetPort() int {
	return c.port
}
