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
	db "cronnest/database"

)

var SshCltConfig ssh.ClientConfig
var GetHostCrontabScriptBase64Content string
var CreateHostCronJobJobScriptBase64Content string
var UpdateHostCronJobJobScriptBase64Content string
var DeleteHostCronJobJobScriptBase64Content string


func MakeHostKv(host models.Host) map[string]interface{} {
	row := make(map[string]interface{})
	row["id"] = host.Id
	row["address"] = host.Address
	row["status"] = host.Status
	row["created_at"] = host.CreatedAt.Format("2006-01-02 15:04:05")
	row["updated_at"] = host.UpdatedAt.Format("2006-01-02 15:04:05")
	return row
}

func MakeHostgroupKv(hg models.Hostgroup) map[string]interface{} {
	row := make(map[string]interface{})
	row["id"] = hg.Id
	row["name"] = hg.Name
	row["created_at"] = hg.CreatedAt.Format("2006-01-02 15:04:05")
	row["updated_at"] = hg.UpdatedAt.Format("2006-01-02 15:04:05")
	return row
}

func MakeHostCrontabKv(hc models.HostCrontab) map[string]interface{} {
	row := make(map[string]interface{})
	row["id"] = hc.Id
	row["host"] = hc.HostId
	row["status"] = hc.Status
	row["msg"] = hc.Msg
	row["tab"] = hc.Tab
	row["created_at"] = hc.CreatedAt.Format("2006-01-02 15:04:05")
	row["updated_at"] = hc.UpdatedAt.Format("2006-01-02 15:04:05")
	if hc.LastSucceed != nil {
		row["last_succeed"] = (*hc.LastSucceed).Format("2006-01-02 15:04:05")
	} else {
		row["last_succeed"] = nil
	}

	return row
}

func UpdateHostCrontabRecord(host models.Host) map[string]interface{} {
	crontabMdl := models.HostCrontab{}
	db.DB.Table("host_crontab").Where("host_id=?", host.Id).Find(&crontabMdl)

	cmd := fmt.Sprintf("echo %s | base64 -d > /tmp/get_host_crontab && chmod +x /tmp/get_host_crontab && /tmp/get_host_crontab %s %s",
		GetHostCrontabScriptBase64Content, configure.SystemCrontabFile, configure.UserCrontabDir)
	succ, output := Command(host.Address, configure.SSH["port"], cmd)
	currTime := time.Now()
	if succ == true {
		tab := map[string]interface{}{}
		json.Unmarshal([]byte(output), &tab)
		jsonData, _ := json.Marshal(tab)

		crontabMdl.Status = "successful"
		crontabMdl.Tab = jsonData
		crontabMdl.Msg = "done"
		crontabMdl.LastSucceed = &currTime
	} else {
		crontabMdl.Status = "failed"
		crontabMdl.Msg = output
		if crontabMdl.Id == 0 {
			tab := map[string]interface{}{}
			jsonData, _ := json.Marshal(tab)
			crontabMdl.Tab = jsonData
		}
	}
	crontabMdl.UpdatedAt = currTime

	if crontabMdl.Id > 0{
		db.DB.Table("host_crontab").Save(&crontabMdl)
	} else {
		crontabMdl.HostId = host.Id
		crontabMdl.CreatedAt = currTime
		db.DB.Table("host_crontab").Create(&crontabMdl)
	}
	crontabData := MakeHostCrontabKv(crontabMdl)
	return crontabData
}

func MakeOperationRecordKv(record models.OperationRecord) map[string]interface{} {
	row := make(map[string]interface{})
	row["id"] = record.Id
	row["resource_type"] = record.SourceType
	row["resource_id"] = record.SourceId
	row["resource_label"] = record.SourceLabel
	row["operation_type"] = record.OperationType
	row["data"] = record.Data
	row["user"] = record.User
	row["created_at"] = record.CreatedAt.Format("2006-01-02 15:04:05")
	return row
}


func Command(host string, port string, cmd string) (bool, string) {
	addr := fmt.Sprintf("%s:%s", host, port)

	sshClient, err := ssh.Dial("tcp", addr, &SshCltConfig)
	if err != nil {
		errMsg := fmt.Sprintf("SSH连接%s Client失败, %v", host, err)
		return false, errMsg
	}
	defer sshClient.Close()

	sshSession, err := sshClient.NewSession()
	if err != nil {
		errMsg := fmt.Sprintf("SSH连接创建Session失败, %v", err)
		return false, errMsg
	}
	defer sshSession.Close()

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
	if sshConf["privateKeyPath"] != "" {
		SshCltConfig.Auth = []ssh.AuthMethod{makeSshAuthMethod(sshConf["privateKeyPath"])}
	} else {
		SshCltConfig.Auth = []ssh.AuthMethod{ssh.Password(sshConf["password"])}
	}

	SshCltConfig.Timeout = time.Duration(5) * time.Second

	GetHostCrontabScriptBase64Content = makeScriptBase64Content("./scripts/get_host_crontab.py")
	CreateHostCronJobJobScriptBase64Content = makeScriptBase64Content("./scripts/create_host_crontab_job.py")
	UpdateHostCronJobJobScriptBase64Content = makeScriptBase64Content("./scripts/update_host_crontab_job.py")
	DeleteHostCronJobJobScriptBase64Content = makeScriptBase64Content("./scripts/delete_host_crontab_job.py")
}