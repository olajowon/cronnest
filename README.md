# Cronnest
一个真正的Crontab管理平台。
***
## Introduction
Cronnest 是一个对Linux Crontab任务进行统一管理的平台工具，通过 SSH实现对远程设备Crontab的全面管理。

**特点：** 没有多余概念，也没有特有名词，更没有学习成本，致力于提供清晰、简单的使用功能。

### DEMO
地址：<http://39.105.81.124:4220/html/crontab/>

用户：guest 

密码：guest123

### Web UI
![image](https://github.com/olajowon/exhibitions/blob/master/cronnest/crontab.png?raw=true)

### 组件
GO	1.3	

Gin 1.4

PgSQL 11 （没错不是MySQL、也不是NoSQL）

VUE	 2.2 （但这不是一个前后端分离项目）

## Installation & Configuration
### 下载
	git clone http://git@github.com:olajowon/cronnest.git

### 修改配置 configure/configure.go

	// PgSQL 连接信息
	PgSQL = "host=localhost user=zhouwang dbname=cronnest sslmode=disable password=123456"	
	// SSH 配置
	SSH = map[string]string {
		"user": "root",
		"password": "",
		"privateKeyPath": "/Users/zhouwang/.ssh/id_rsa",	// ras 私钥绝对路径 （优先）
		"port": "22",										// 端口，注意是字符串
	}
	
### 用户配置 configure/account.go
	Accounts = map[string]string {
		"zhangsan": "123456",	// 用户名: 密码,
	}
	
### 建表Sql
	table.sql // 烦请手动建表

## Start Up

### 编译 
	go build -o cronnest .
	
### 启动 
	./cronnest

### 访问
	http://localhost:9900/html/crontab/	

