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
	row["created"] = job.Created
	row["updated"] = job.Updated
	row["hosts"] = job.Hosts
	row["sysuser"] = job.Sysuser
	return row
}