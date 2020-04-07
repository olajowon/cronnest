package configure

var PgSQL string
var Log map[string]string
var SystemCrontabFile string
var UserCrontabDir string
var SSH map[string]string

func init() {
	PgSQL = "host=localhost user=root dbname=cronnest sslmode=disable password=123456"	// pgsql
	Log = map[string]string {
		"request": "/var/log/cronnest/request.log",
		"cronnest": "/var/log/cronnest/cronnest.log",
	}
	SSH = map[string]string {
		"user": "zhouwang",
		"password": "",
		"privateKeyPath": "/Users/zhouwang/.ssh/id_rsa",	// ras 私钥绝对路径 （优先）
		"port": "22",										// 端口，注意是字符串
	}
	SystemCrontabFile = "/etc/crontab"	// 系统crontab文件
	UserCrontabDir = "/var/spool/cron"	// 用户crontab目录
}
