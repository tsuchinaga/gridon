package gridon

type testLogger struct {
	ILogger
	noticeHistory  []interface{}
	warningHistory []interface{}
}

func (t *testLogger) Notice(v ...interface{})  { t.noticeHistory = append(t.noticeHistory, v...) }
func (t *testLogger) Warning(v ...interface{}) { t.warningHistory = append(t.warningHistory, v...) }
