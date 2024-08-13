package dns

type Record struct {
	Name    string
	Content string
}

func AppendDnsRecord(api IDns, record Record) error {
	return api.AppendRecord(record.Name, record.Content)
}

func RemoveDnsRecord(api IDns, record Record) error {
	return api.RemoveRecord(record.Name, record.Content)
}

func ListDnsRecords(api IDns) (map[string]string, error) {
	return api.ListRecords()
}
