package availability

type AvailabilityStorageEntity struct {
	Key   string
	Data  string
	Store string
}

func (p *AvailabilityStorageEntity) GetKey() string {
	return p.Key
}

func (p *AvailabilityStorageEntity) GetData() string {
	return p.Data
}

func (p *AvailabilityStorageEntity) GetStore() string {
	return p.Store
}

func (p *AvailabilityStorageEntity) GenerateMappingKey(source, sourceId string) string {
	return ""
}

type Mapping struct {
	Source      string
	Destination string
}

func (m Mapping) GetSource() string {
	return m.Source
}

func (m Mapping) GetDestination() string {
	return m.Destination
}
