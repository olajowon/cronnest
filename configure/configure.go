package configure

var PgSQL string
var Log string
var SSH map[string]string

func init() {
	PgSQL = "host=localhost user=zhouwang dbname=cronnest sslmode=disable password=123456"
	Log = "/var/log/cronnest/cronnest.log"
	SSH = map[string]string {
		"user": "root",
		"password": "",
		"privateKeyPath": "/Users/zhouwang/.ssh/id_rsa",
		"port": "22",
	}
}
