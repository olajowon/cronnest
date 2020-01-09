package handlers

import "cronnest/models"

func MakeJobKv(job models.Job) map[string]interface{} {
	row := make(map[string]interface{})
	row["id"] = job.Id
	row["name"] = job.Name
	row["status"] = job.Status
	row["description"] = job.Description
	row["mailto"] = job.Mailto
	row["spec"] = job.Spec
	row["content"] = job.Content
	row["log"] = job.Log
	row["output"] = job.Output
	row["created"] = job.Created.Format("2006-01-02 15:04:05")
	row["updated"] = job.Updated.Format("2006-01-02 15:04:05")
	row["hosts"] = job.Hosts
	row["user"] = job.Sysuser
	return row
}

func MakeHostKv(host models.Host) map[string]interface{} {
	row := make(map[string]interface{})
	row["id"] = host.Id
	row["address"] = host.Address
	row["status"] = host.Status
	row["created"] = host.Created.Format("2006-01-02 15:04:05")
	row["updated"] = host.Updated.Format("2006-01-02 15:04:05")
	return row
}

func MakeOperationRecordKv(record models.OperationRecord) map[string]interface{} {
	row := make(map[string]interface{})
	row["id"] = record.Id
	row["resource_type"] = record.ResourceType
	row["resource_id"] = record.ResourceId
	row["resource_label"] = record.ResourceLabel
	row["operation_type"] = record.OperationType
	row["data"] = record.Data
	row["user"] = record.User
	row["created"] = record.Created.Format("2006-01-02 15:04:05")
	return row
}