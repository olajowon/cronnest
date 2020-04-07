package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"cronnest/models"
	db "cronnest/database"
	lg "cronnest/logger"
	"fmt"
	"encoding/base64"
	"cronnest/configure"
	"golang.org/x/crypto/ssh"
	"encoding/json"
	"time"
)

func CreateHostCrontabJob(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()

	hId := c.Param("hId")
	hostMdl := models.Host{}
	db.DB.Table("host").Where("id = ?", hId).First(&hostMdl)
	if hostMdl.Id == 0{
		c.JSON(http.StatusNotFound, gin.H{"msg": "Host not found"})
		return
	}

	rawData, _ := c.GetRawData()
	byteJob := []byte(rawData)
	base64Job := base64.StdEncoding.EncodeToString(byteJob)

	cmd := fmt.Sprintf("echo %s | base64 -d > /tmp/cronnest_create_job && chmod +x /tmp/cronnest_create_job && /tmp/cronnest_create_job %s %s %s",
		CreateHostCronJobJobScriptBase64Content, configure.SystemCrontabFile, configure.UserCrontabDir, base64Job)

	addr := fmt.Sprintf("%s:%s", hostMdl.Address, configure.SSH["port"])

	status, output := cmdStatusOutput(addr, cmd)
	if status == 400 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": output})
	} else if status == 500 {
		msg := fmt.Sprintf("SSH创建主机Crontab job失败, %s", output)
		lg.Logger.Error(msg)
		c.JSON(http.StatusInternalServerError, gin.H{"msg": msg})
	} else {
		created := map[string]interface{}{}
		json.Unmarshal([]byte(output), &created)

		recordData, _ := json.Marshal(map[string]interface{}{"created_job": created})
		operRecord := models.OperationRecord{SourceType:"host_crontab_job", SourceId:hostMdl.Id,
			SourceLabel: hostMdl.Address, OperationType: "create", Data:recordData, User: user, CreatedAt: time.Now()}
		if result := db.DB.Table("operation_record").Create(&operRecord); result.Error != nil {
			msg := fmt.Sprintf("记录操作失败, %v", result.Error)
			lg.Logger.Error(msg)
			c.JSON(http.StatusInternalServerError,
				gin.H{"code": 500, "msg": msg})
			return
		}

		crontab := UpdateHostCrontabRecord(hostMdl)
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"created": created, "tab": crontab["tab"]}})
	}
}

func UpdateHostCrontabJob(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()
	hId := c.Param("hId")
	hostMdl := models.Host{}
	db.DB.Table("host").Where("id = ?", hId).First(&hostMdl)
	if hostMdl.Id == 0{
		c.JSON(http.StatusNotFound, gin.H{"msg": "Host not found"})
		return
	}
	rawData, _ := c.GetRawData()
	byteJob := []byte(rawData)
	base64Job := base64.StdEncoding.EncodeToString(byteJob)

	cmd := fmt.Sprintf("echo %s | base64 -d > /tmp/cronnest_update_job && chmod +x /tmp/cronnest_update_job && /tmp/cronnest_update_job %s %s %s",
		UpdateHostCronJobJobScriptBase64Content, configure.SystemCrontabFile, configure.UserCrontabDir, base64Job)

	addr := fmt.Sprintf("%s:%s", hostMdl.Address, configure.SSH["port"])

	status, output := cmdStatusOutput(addr, cmd)

	if status == 400 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": output})
	} else if status == 500 {
		c.JSON(http.StatusInternalServerError, gin.H{ "msg": output})
	} else {
		updated := map[string]interface{}{}
		json.Unmarshal([]byte(output), &updated)

		recordData, _ := json.Marshal(map[string]interface{}{"updated_job": updated})
		operRecord := models.OperationRecord{SourceType:"host_crontab_job", SourceId:hostMdl.Id,
			SourceLabel: hostMdl.Address, OperationType: "update", Data:recordData, User: user, CreatedAt: time.Now()}
		if result := db.DB.Table("operation_record").Create(&operRecord); result.Error != nil {
			msg := fmt.Sprintf("记录操作失败, %v", result.Error)
			lg.Logger.Error(msg)
			c.JSON(http.StatusInternalServerError,
				gin.H{"code": 500, "msg": msg})
			return
		}

		crontab := UpdateHostCrontabRecord(hostMdl)
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"updated": updated, "tab": crontab["tab"]}})
	}
}


func DeleteHostCrontabJob(c *gin.Context) {
	user, _, _ := c.Request.BasicAuth()
	hId := c.Param("hId")
	hostMdl := models.Host{}
	db.DB.Table("host").Where("id = ?", hId).First(&hostMdl)
	if hostMdl.Id == 0{
		c.JSON(http.StatusNotFound, gin.H{"msg": "Host not found"})
		return
	}
	rawData, _ := c.GetRawData()
	byteJob := []byte(rawData)
	base64Job := base64.StdEncoding.EncodeToString(byteJob)

	cmd := fmt.Sprintf("echo %s | base64 -d > /tmp/cronnest_delete_job && " +
		"chmod +x /tmp/cronnest_delete_job && /tmp/cronnest_delete_job %s %s %s",
		DeleteHostCronJobJobScriptBase64Content, configure.SystemCrontabFile, configure.UserCrontabDir, base64Job)

	addr := fmt.Sprintf("%s:%s", hostMdl.Address, configure.SSH["port"])

	status, output := cmdStatusOutput(addr, cmd)

	if status == 400 {
		c.JSON(http.StatusBadRequest, gin.H{"msg": output})
	} else if status == 500 {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": output})
	} else {
		removed := map[string]interface{}{}
		json.Unmarshal([]byte(output), &removed)
		fmt.Println(output)
		recordData, _ := json.Marshal(map[string]interface{}{"deleted_job": removed})
		operRecord := models.OperationRecord{SourceType:"host_crontab_job", SourceId:hostMdl.Id,
			SourceLabel: hostMdl.Address, OperationType: "delete",
			Data:recordData, User: user, CreatedAt: time.Now()}
		if result := db.DB.Table("operation_record").Create(&operRecord); result.Error != nil {
			msg := fmt.Sprintf("记录操作失败, %v", result.Error)
			lg.Logger.Error(msg)
			c.JSON(http.StatusInternalServerError,
				gin.H{"code": 500, "msg": msg})
			return
		}

		crontab := UpdateHostCrontabRecord(hostMdl)
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"removed": removed, "tab": crontab["tab"]}})
	}
}


func cmdStatusOutput(addr string, cmd string) (int64, string){
	sshClient, err := ssh.Dial("tcp", addr, &SshCltConfig)
	if err != nil {
		errMsg := fmt.Sprintf("SSH连接%v失败, %v", addr, err)
		return 500, errMsg
	}
	defer sshClient.Close()

	sshSession, err := sshClient.NewSession()
	if err != nil {
		errMsg := fmt.Sprintf("SSH连接创建Session失败, %v", err)
		return 500, errMsg
	}
	defer sshSession.Close()

	output, err := sshSession.CombinedOutput(cmd)
	if err != nil {
		if err.Error() == "Process exited with status 40" {
			return 400, string(output)
		} else {
			return 500, string(output)
		}
	}
	return 200, string(output)
}