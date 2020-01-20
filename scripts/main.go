package main

import (
	_ "github.com/jinzhu/gorm/dialects/postgres"
	db "cronnest/database"
	"cronnest/models"
	"fmt"
	"encoding/json"
	"encoding/base64"
	"cronnest/configure"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"github.com/sirupsen/logrus"
	"os"
	"time"
	"sync"
)

const LogPath = "/tmp/cronnest-worker.log"
const Forks = 1

var Log = logrus.New()
var HostAddrs []string
var AgentBase64Content string
var SshClientConfig ssh.ClientConfig

func pushJobs(hostAddr string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println(hostAddr)

	status, jobs, msg := pushJobsTask(hostAddr)

	if status != "successful" {
		Log.Errorf("pushJobs(%s) error, Task error, status: %s", hostAddr, status)
	} else {
		Log.Infof("pushJobs(%s), Task {status: %s}", hostAddr, status)
	}

	jsonJobs, _ := json.Marshal(jobs)
	transaction := db.DB.Begin()
	pushRecord := models.PushRecord{Host:hostAddr, Status:status, Jobs:jsonJobs, Msg:msg, Created:time.Now()}
	result := transaction.Table("push_record").Create(&pushRecord)
	if result.Error != nil {
		transaction.Rollback()
		Log.Errorf("pushJobs(%s) error: create push record failed, %v", hostAddr, result.Error)
		return
	}

	host := models.Host{}
	transaction.Table("host").Where("address = ?", hostAddr).First(&host)
	host.PushStatus = status
	host.Updated = time.Now()
	result = transaction.Table("host").Save(&host)
	if result.Error != nil {
		transaction.Rollback()
		Log.Errorf("pushJobs(%s) error: update host push status failed, %v", hostAddr, result.Error)
		return
	}
	transaction.Commit()
	Log.Infof("pushJobs(%s) finish, Task {status: %s}", hostAddr, status)
}


func pushJobsTask(hostAddr string) (string, []models.JobInfo, string){
	var jobs []models.JobInfo
	query := fmt.Sprintf("hosts @> '[\"%v\"]'::jsonb", hostAddr)
	result := db.DB.Table("job").Where(query).Find(&jobs)

	if result.Error != nil {
		errMsg := fmt.Sprintf("从数据库获取主机Job失败, %v", result.Error)
		Log.Errorf("pushJobsTask(%s) error: %s", hostAddr, errMsg)
		return "failed", jobs, errMsg
	}

	jsonJobs, _ := json.Marshal(jobs)

	byteJobs := []byte(jsonJobs)
	base64Jobs := base64.StdEncoding.EncodeToString(byteJobs)

	cmd := fmt.Sprintf("echo %s | base64 -d > /tmp/cronnest_agt && chmod +x /tmp/cronnest_agt && /tmp/cronnest_agt %s",
		AgentBase64Content, base64Jobs)

	addr := fmt.Sprintf("%s:%s", hostAddr, configure.SSH["port"])

	sshClient, err := ssh.Dial("tcp", addr, &SshClientConfig)
	if err != nil {
		errMsg := fmt.Sprintf("创建 SSH Client 失败, %v", err)
		Log.Errorf("pushJobsTask(%s) error: %s", hostAddr, errMsg)
		return "failed", jobs, errMsg
	}
	defer sshClient.Close()

	sshSession, err := sshClient.NewSession()
	if err != nil {
		errMsg := fmt.Sprintf("创建 SSH Session 失败, %v", err)
		Log.Errorf("pushJobsTask(%s) error: %s", hostAddr, errMsg)
		return "failed", jobs, errMsg
	}
	defer sshSession.Close()

	output, err := sshSession.CombinedOutput(cmd)
	if err != nil {
		errMsg := fmt.Sprintf("同步Job至主机失败, %v, \n%s", err, output)
		return "failed", jobs, errMsg
	}

	return "successful", jobs, string(output)
}


func makeSshClientConfig() {
	sshConf := configure.SSH
	SshClientConfig = ssh.ClientConfig{}
	SshClientConfig.User = sshConf["user"]
	SshClientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	SshClientConfig.Auth = []ssh.AuthMethod{makeSshAuthMethod(sshConf["privateKeyPath"])}
	SshClientConfig.Timeout = time.Duration(5) * time.Second
}


func makeAgentContent()  {
	b, err := ioutil.ReadFile("agent.py")
	if err != nil {
		log.Fatalf("makeAgentContent error: read file failed, %v", err)
	}
	content := string(b)
	byteContent := []byte(content)
	AgentBase64Content = base64.StdEncoding.EncodeToString(byteContent)
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


func makeHostAddrs() {
	var hosts []models.Host
	result := db.DB.Table("host").Where("status = ?", "enabled").Find(&hosts)
	if result.Error == nil {
		var hostAddrs []string
		for _, host := range hosts {
			hostAddrs = append(hostAddrs, host.Address)
		}
		HostAddrs = hostAddrs
	} else {
		Log.Errorf("makeHostAddrs error: %v", result.Error)
	}
}


func pushEntry() {
	begin := time.Now().Unix()
	for startIdx := 0; startIdx < len(HostAddrs); startIdx += Forks {
		var endIdx int
		if startIdx + Forks > len(HostAddrs) {
			endIdx = len(HostAddrs)
		} else {
			endIdx = startIdx + Forks
		}

		addrs := HostAddrs[startIdx: endIdx]
		fmt.Println(addrs)

		var wg sync.WaitGroup
		for _, hostAddr := range addrs {
			wg.Add(1)
			go pushJobs(hostAddr, &wg)
		}
		wg.Wait()
	}
	finish := time.Now().Unix()
	duration := finish - begin
	if duration >= 60 {
		Log.Warnf("pushEntry warning: 推送任务耗时达到 %d 秒，请添加线程或清理无用主机，保证推送任务 60 秒内能够完成！")
	}
}


func init() {
	Log.Out = os.Stdout
	file, err := os.OpenFile(LogPath, os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		Log.Out = file
	}
}


func main() {
	defer db.DB.Close()

	makeSshClientConfig()
	makeAgentContent()
	makeHostAddrs()
	fmt.Println(HostAddrs)

	pushEntry()
	//c := cron.New()
	//spec := "0 */1 * * * ?"
	//c.AddFunc(spec, pushEntry)
	//c.Start()

	select {}
}
