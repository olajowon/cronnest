package configure

var Accounts map[string]string

func init() {
	Accounts = map[string]string {
		"admin": "admin",	// 用户名: 密码
		"guest": "guest123",
	}
}