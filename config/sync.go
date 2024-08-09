package config

type SyncEntity struct {
	Resource     string
	Table        string
	FilterColumn string `mapstructure:"filter_column"`
	IdColumn     string `mapstructure:"id_column"`
	Store        bool
	Locale       bool
	QueueGroup   string `mapstructure:"queue_group"`
	Mappings     []SyncEntityMapping
}

type SyncEntityMapping struct {
	Source      string
	Destination string
}

func (m SyncEntityMapping) GetSource() string {
	return m.Source
}

func (m SyncEntityMapping) GetDestination() string {
	return m.Destination
}
