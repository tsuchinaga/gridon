package gridon

type testLogger struct {
	ILogger
	NoticeCount     int
	NoticeHistory   []interface{}
	WarningCount    int
	WarningHistory  []interface{}
	CashFlowCount   int
	CashFlowHistory []interface{}
}

func (t *testLogger) Notice(v ...interface{}) {
	t.NoticeHistory = append(t.NoticeHistory, v...)
	t.NoticeCount++
}
func (t *testLogger) Warning(v ...interface{}) {
	t.WarningHistory = append(t.WarningHistory, v...)
	t.WarningCount++
}
func (t *testLogger) CashFlow(v ...interface{}) {
	t.CashFlowHistory = append(t.CashFlowHistory, v...)
	t.CashFlowCount++
}
