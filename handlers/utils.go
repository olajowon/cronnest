package handlers

import (
	"cronnest/models"
	"cronnest/configure"
	"golang.org/x/crypto/ssh"
	"time"
	"io/ioutil"
	"log"
	"encoding/base64"
	"fmt"
	"encoding/json"
)

var SshCltConfig ssh.ClientConfig
var UpdateCronJobScriptBase64Content string
var RemoveCronJobScriptBase64Content string

func MakeJobKv(job models.Job) map[string]interface{} {
	row := make(map[string]interface{})
	row["id"] = job.Id
	row["status"] = job.Status
	row["comment"] = job.Comment
	row["spec"] = job.Spec
	row["content"] = job.Content
	row["log"] = job.Log
	row["created"] = job.Created.Format("2006-01-02 15:04:05")
	row["updated"] = job.Updated.Format("2006-01-02 15:04:05")
	row["host"] = job.Host
	row["sysuser"] = job.Sysuser
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


func DoHostCronJob(job models.Job, active string) (bool, string) {
	addr := fmt.Sprintf("%s:%s", job.Host, configure.SSH["port"])

	sshClient, err := ssh.Dial("tcp", addr, &SshCltConfig)
	if err != nil {
		errMsg := fmt.Sprintf("创建 SSH Client 失败, %v", err)
		return false, errMsg
	}
	defer sshClient.Close()

	sshSession, err := sshClient.NewSession()
	if err != nil {
		errMsg := fmt.Sprintf("创建 SSH Session 失败, %v", err)
		return false, errMsg
	}
	defer sshSession.Close()

	jsonJob, _ := json.Marshal(job)

	byteJob := []byte(jsonJob)
	base64Job := base64.StdEncoding.EncodeToString(byteJob)

	var scriptBase64Content string
	if active == "update" {
		scriptBase64Content = UpdateCronJobScriptBase64Content
	} else if active == "remove" {
		scriptBase64Content = RemoveCronJobScriptBase64Content
	}

	cmd := fmt.Sprintf("echo %s | base64 -d > /tmp/cronnest_%s_job && chmod +x /tmp/cronnest_%s_job && /tmp/cronnest_%s_job %s",
		scriptBase64Content, active, active, active, base64Job)

	output, err := sshSession.CombinedOutput(cmd)
	if err != nil {
		errMsg := fmt.Sprintf("执行出错, %v, \n%s", err, output)
		return false, errMsg
	}

	return true, string(output)
}


func makeScriptBase64Content(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("makeScriptBase64Content error: read file failed, %v", err)
	}
	content := string(b)
	byteContent := []byte(content)
	return base64.StdEncoding.EncodeToString(byteContent)
}

func makeSshAuthMethod(keyPath string) ssh.AuthMethod {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatalf("makeSshAuthMethod error: read keyfile failed, %v", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("makeSshAuthMethod error: parse private key failed, %v", err)
	}
	return ssh.PublicKeys(signer)
}


func init() {
	sshConf := configure.SSH
	SshCltConfig = ssh.ClientConfig{}
	SshCltConfig.User = sshConf["user"]
	SshCltConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	SshCltConfig.Auth = []ssh.AuthMethod{makeSshAuthMethod(sshConf["privateKeyPath"])}
	SshCltConfig.Timeout = time.Duration(5) * time.Second

	UpdateCronJobScriptBase64Content = makeScriptBase64Content("./scripts/update_cron_job.py")
	RemoveCronJobScriptBase64Content = makeScriptBase64Content("./scripts/remove_cron_job.py")
}