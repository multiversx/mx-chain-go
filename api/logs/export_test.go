package logs

const MsgQueueSize = msgQueueSize

type LogSender struct {
	*logSender
}

func (ls *LogSender) Set(l *logSender) {
	ls.logSender = l
}

func (ls *logSender) Writer() *logWriter {
	return ls.writer
}

func (ls *logSender) SetWriter(lw *logWriter) {
	ls.writer = lw
}
