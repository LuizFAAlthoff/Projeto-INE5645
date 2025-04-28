package model

import "net"

type ParsedCommand struct {
	Conn   net.Conn
	Raw    string
	Action string
	Key    string
	Value  string
	Result string
}
