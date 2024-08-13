package dns

type IDns interface {
	ListRecords() (map[string]string, error)
	AppendRecord(name string, content string) error
	RemoveRecord(name string, content string) error
}
