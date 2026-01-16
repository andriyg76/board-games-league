package glog

type compositeOuts struct {
	chain []Output
}

func compositeOutput(out ...Output) Output {
	return compositeOuts{chain: out}
}

func (c compositeOuts) Printf(format string, a ...interface{}) {
	for _, l := range c.chain {
		l.Printf(format, a...)
	}
}

type composite struct {
	chain []Logger
}

func (c composite) Debug(format string, a ...interface{}) {
	for _, l := range c.chain {
		l.Debug(format, a...)
	}
}

func (c composite) IsDebug() bool {
	for _, l := range c.chain {
		if l.IsDebug() {
			return true
		}
	}
	return false
}

func (c composite) DebugLogger() Output {
	var out []Output
	for _, l := range c.chain {
		out = append(out, l.DebugLogger())
	}
	return compositeOutput(out...)
}

func (c composite) Trace(format string, a ...interface{}) {
	for _, l := range c.chain {
		l.Trace(format, a...)
	}
}

func (c composite) TraceLogger() Output {
	var out []Output
	for _, l := range c.chain {
		out = append(out, l.TraceLogger())
	}
	return compositeOutput(out...)
}

func (c composite) IsTrace() bool {
	for _, l := range c.chain {
		if l.IsTrace() {
			return true
		}
	}
	return false
}

func (c composite) Warn(format string, a ...interface{}) {
	for _, l := range c.chain {
		l.Warn(format, a...)
	}
}

func (c composite) IsWarn() bool {
	for _, l := range c.chain {
		if l.IsWarn() {
			return true
		}
	}
	return false
}

func (c composite) Info(format string, a ...interface{}) {
	for _, l := range c.chain {
		l.Info(format, a...)
	}
}

func (c composite) IsInfo() bool {
	for _, l := range c.chain {
		if l.IsInfo() {
			return true
		}
	}
	return false
}

func (c composite) Error(format string, a ...interface{}) error {
	var err error
	for _, l := range c.chain {
		err = l.Error(format, a...)
	}
	return err
}

func (c composite) IsError() bool {
	for _, l := range c.chain {
		if l.IsError() {
			return true
		}
	}
	return false
}

func (c composite) Log(LogLevel LogLevel, format string, a ...interface{}) {
	for _, l := range c.chain {
		l.Log(LogLevel, format, a...)
	}
}

func (c composite) IsEnabled(logLevel LogLevel) bool {
	for _, l := range c.chain {
		if l.IsEnabled(logLevel) {
			return true
		}
	}
	return false
}

func (c composite) GetOutput(LogLevel LogLevel) Output {
	var out []Output
	for _, l := range c.chain {
		out = append(out, l.GetOutput(LogLevel))
	}
	return compositeOutput(out...)
}

func (c composite) Panic(format string, a ...interface{}) {
	for _, l := range c.chain {
		l.Panic(format, a...)
	}
}

func (c composite) Fatal(format string, a ...interface{}) {
	for _, l := range c.chain {
		l.Fatal(format, a...)
	}
}

func DefaultComposite(main Logger, loggers ...Logger) {
	_default = Composite(main, loggers...)
}

func Composite(main Logger, loggers ...Logger) Logger {
	return composite{chain: append([]Logger{main}, loggers...)}
}
