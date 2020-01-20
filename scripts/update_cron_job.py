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


def update_job(j):
    job_id = j['Id']
    job_comment = j['Comment']
    job_user = j['Sysuser']
    job_slices = j['Spec']
    job_script = os.path.join(job_script_dir, 'cronnest_job_%d.script' % job_id)
    job_command = '%s' % job_script
    job_content = j['Content']
    job_enabled = j['Status'] == 'enabled'
    job_row = '%s %s # %s' % (job_slices, job_command, job_comment)

    tabnames = os.listdir(cron_dir)
    for tabname in tabnames:
        if job_user != tabname:
            tabfile = os.path.join(cron_dir, tabname)
            job_label = '[%s] @ %s' % (job_comment, tabname)
            cron = CronTab(tabfile=tabfile)
            try:
                job_iter = cron.find_command('cronnest_job_%d.script' % job_id)

                for job in job_iter:
                    print(' 删除 %s' % job_label)
                    print(u' --> 任务条目 [%s%s %s # %s]' % ('' if job.enabled else '# ',
                                                                   job.slices, job.command, job.comment))
                    print(' --> 准备删除任务条目[%s]' % job)
                    cron.remove(job)
                    cron.write()
                    print(' --> 已删除任务条目')
                    print(' --> 成功\n')
                    break
            except Exception as e:
                errmsg = '未知错误, %s' % str(e)
                print(' --> %s' % errmsg)
                exit(1)

    try:
        cron = CronTab(user=job_user)

        job_label = '[%s] @ %s' % (job_comment, job_user)
        print(' 更新 %s' % job_label)

        update_job_script(job_script, job_content)

        print(' --> 任务条目 [%s%s %s # %s]' % ('' if job_enabled else '# ',
                                                    job_slices, job_command, job_comment))
        job_iter = cron.find_command('cronnest_job_%d.script' % job_id)
        for job in job_iter:
            print(u' --> 原任务条目 [%s]' % job)
            if job.comment != job_comment or job.command != job_command or \
                str(job.slices) != job_slices or job.enabled != job_enabled:
                print(' --> 此任务条目发生了变化，准备变更任务条目')
                job.set_comment(job_comment)
                job.set_command(job_command)
                job.setall(job_slices)
                job.enable(job_enabled)
                cron.write()
                print(' --> 已变更此任务条目')
            else:
                print(' --> 此任务条目没有发生变化，略过')
            break
        else:
            print(' --> 原任务条目 [无]')
            print(' --> 此任务条目之前不存在，准备创建新的任务条目')
            job = cron.new(command=job_command, comment=job_comment)
            job.setall(job_slices)
            job.enable(job_enabled)
            cron.write()
            print(' --> 已创建新的任务条目')
    except Exception as e:
        errmsg = '未知错误, %s' % str(e)
        print(' --> %s' % errmsg)
        exit(1)
    else:
        print(' --> 成功\n')


def update_job_script(job_script, job_content):
    print(' --> 任务脚本 [%s]' % job_script)
    if not os.path.isfile(job_script):
        print(' --> 任务脚本之前不存在，准备创建新脚本')
        if not os.path.isdir(job_script_dir):
            os.makedirs(job_script_dir)
        open(job_script, 'w').write(job_content)
        print(' --> 已创建新脚本')
    else:
        content = open(job_script, 'r').read()
        if content != job_content:
            print(' --> 任务脚本内容发生了变化，准备更新脚本内容')
            open(job_script, 'w').write(job_content)
            print(' --> 已更新脚本内容')
        else:
            print(' --> 任务脚本没有发生变化，略过')

update_job(job)