package gridon

type testLogger struct {
	ILogger
	NoticeCount    int
	NoticeHistory  []interface{}
	WarningCount   int
	WarningHistory []interface{}
}

func (t *testLogger) Notice(v ...interface{}) {
	t.NoticeHistory = append(t.NoticeHistory, v...)
	t.NoticeCount++
}
func (t *testLogger) Warning(v ...interface{}) {
	t.WarningHistory = append(t.WarningHistory, v...)
	t.WarningCount++
}
