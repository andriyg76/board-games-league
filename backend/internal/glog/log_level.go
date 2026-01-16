package glog

type LogLevel struct {
	prefix string
	weight int
}

var TRACE = LogLevel{
	prefix: "TRACE",
	weight: -2,
}

var DEBUG = LogLevel{
	prefix: "DEBUG",
	weight: -1,
}

var INFO = LogLevel{
	prefix: " INFO",
	weight: 0,
}

var WARN = LogLevel{
	prefix: " WARN",
	weight: 1,
}

var ERROR = LogLevel{
	prefix: "ERROR",
	weight: 2,
}

var PANIC = LogLevel{
	prefix: "PANIC",
	weight: 2,
}

var FATAL = LogLevel{
	prefix: "FATAL",
	weight: 2,
}

func (l LogLevel) String() string {
	return l.prefix
}
