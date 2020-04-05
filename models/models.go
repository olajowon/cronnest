package models

import (
	"encoding/json"
	"time"
)

type Hostgroup struct {
	Id			int64
	Name		string
	CreatedAt		time.Time
	UpdatedAt		time.Time
}

type HostCrontab struct {
	Id			int64
	HostId		int64
	Status		string
	Msg		    string
	Tab 		json.RawMessage `sql:"type:jsonb"`
	CreatedAt	time.Time
	UpdatedAt	time.Time
	LastSucceed *time.Time `gorm:"default:NULL"`
}

type Host struct {
	Id			int64
	Address		string
	Status		string
	CreatedAt	time.Time
	UpdatedAt	time.Time
}

type HostgroupHost struct {
	HostgroupId	int64
	HostId		int64
}

type OperationRecord struct {
	Id				int64
	SourceType	string
	SourceId		int64
	SourceLabel	string
	OperationType	string
	Data			json.RawMessage `sql:"type:jsonb"`
	User			string
	CreatedAt		time.Time
}