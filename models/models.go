package models

import (
	"encoding/json"
	"time"
)

type Job struct {
	Id			int64
	Comment		string
	Status		string
	Spec		string
	Content		string
	Log			string
	Created		time.Time
	Updated		time.Time
	Host		string
	Sysuser		string
}

type JobInfo struct {
	Id			int64
	Name		string
	Status		string
	Description	string
	Mailto		string
	Spec		string
	Content		string
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
	ResourceLabel	string
	OperationType	string
	Data			json.RawMessage `sql:"type:jsonb"`
	User			string
	Created			time.Time
}

type PushRecord struct {
	Id				int64
	Host			string
	Status			string
	Jobs			json.RawMessage `sql:"type:jsonb"`
	Msg				string
	Created			time.Time
}