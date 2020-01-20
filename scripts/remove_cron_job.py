#!/usr/bin/python
#-*-coding:utf-8-*-
import sys
import commands
import json
import base64
import os
import re
reload(sys)
sys.setdefaultencoding('utf-8')

try:
    from crontab import CronTab
except:
    os.system('yum install python-pip')
    os.system('pip install python-crontab')
    from crontab import CronTab

job = json.loads(base64.b64decode(sys.argv[1]))

cron_dir = '/var/spool/cron/'
job_script_dir = '/opt/cronnest/'

def remove_job(j):
    tabnames = os.listdir(cron_dir)
    job_id = j['Id']
    for tabname in tabnames:
        job_user = tabname
        tabfile = os.path.join(cron_dir, tabname)
        cron = CronTab(tabfile=tabfile)
        job_iters = cron.find_command('cronnest_job_%d.script' % job_id)
        for job in job_iters:
            job_enabled = job.enabled
            job_slices = str(job.slices)
            job_command = job.command
            job_comment = job.comment
            job_script = os.path.join(job_script_dir, 'cronnest_job_%d.script' % job_id)
            job_label = '[%s] @ %s' % (job_comment, job_user)

            print(' 删除 %s' % job_label)
            try:
                print(' --> 任务脚本 [%s]' % job_script)
                print(' --> 准备删除任务脚本')
                os.system('rm -rf %s' % job_script)
                print(' --> 已删除任务脚本')

                print(' --> 任务条目 [%s%s %s # %s]' % ('' if job_enabled else '# ',
                                                                        job_slices, job_command, job_comment))
                print(' --> 准备删除任务条目[%s]' % job)
                cron.remove(job)
                print(' --> 已删除任务条目')
                cron.write()
            except Exception as e:
                errmsg = '未知错误：%s' % str(e)
                print(' --> %s' % errmsg)
                exit(1)
            else:
                print(' --> 成功\n')

remove_job(job)