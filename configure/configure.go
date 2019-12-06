package configure

var PgSQL string
var Log string

func init() {
	PgSQL = "host=localhost user=zhouwang dbname=cronnest sslmode=disable password=123456"
	Log = "/var/log/cronnest/cronnest.log"
}
