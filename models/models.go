package models

import (
	"encoding/json"
	"time"
)

type Job struct {
	Id			int64
	Name		string
	Status		string
	Description	string
	Mailto		string
	Spec		string
	Content		string
	Log			string
	Output		string
	Created		time.Time
	Updated		time.Time
	Hosts		json.RawMessage `sql:"type:jsonb"`
	Sysuser		string
}

type Host struct {
	Id			int64
	Address		string
	Status		string
	Created		time.Time
	Updated		time.Time
}

type OperationRecord struct {
	Id				int64
	ResourceType	string
	ResourceId		int64
	OperationType	string
	Data			json.RawMessage `sql:"type:jsonb"`
	User			string
	Created			time.Time
}