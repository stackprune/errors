package errors_test

const (
	testNameNilError             = "nil error"
	testNameStackpruneErrorInput = "stackprune error input"
	testNameWithStack            = "with stack"
	testNameWithWrap             = "with wrap"
	testNameWrappedError         = "wrapped error"
	testMessageFoo               = "foo"
	testMessageWrappedError      = "wrapped error"
	testFormatDetailedVerbose    = "%+v"
	testLogKeyMessage            = "message"
	testLogKeyKind               = "kind"
	testLogKeyStack              = "stack"
	testLogKeyChildren           = "children"
	testLogKindError             = "*errors.Error"
	testLogKindJoinError         = "*errors.JoinError"
	testLogKindStandardError     = "*errors.errorString"
	testLogKeyCustomMessage      = "custom_message"
	testLogKeyCustomKind         = "custom_kind"
	testLogKeyCustomStack        = "custom_stack"
	testLogMessageWrappedEOF     = "wrapped error: EOF"
)
