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

jobs = json.loads(base64.b64decode(sys.argv[1]))


cron_dir = '/var/spool/cron/'
job_script_dir = '/opt/cronnest/'
job_output_dir = '/var/log/cronnest/'
valids = []
removeds = []
faileds = []

created, updated, skipped, cleaned, failed = [], [], [], [], []

def update_jobs(jobs):
    print('# 更新任务\n')
    global valids, faileds
    for j in jobs:
        job_name = j['Name']
        job_desc = j['Description']
        job_user = j['Sysuser']
        job_slices = j['Spec']
        job_comment = '%s: %s' %  (job_name, job_desc)
        job_script = os.path.join(job_script_dir, 'cronnest_job_%d.script' % j['Id'])
        job_command = '%s' % job_script
        job_content = j['Content']
        job_enabled = j['Status'] == 'enabled'
        job_row = '%s %s # %s' % (job_slices, job_command, job_comment)
        job_label = '[%s][%s] @ %s' % (job_name, job_desc, job_user)
        print(' 更新 %s' % job_label)
        try:
            update_job_script(job_script, job_content)

            cron = CronTab(user=job_user)

            print(' --> 任务条目 [%s%s %s # %s]' % ('' if job_enabled else '# ',
                                                        job_slices, job_command, job_comment))
            job_iter = cron.find_command(re.compile(r'cronnest_job_%d.script' % j['Id']))
            for job in job_iter:
                print(u' --> 原任务条目 [%s%s %s # %s]' % ('' if job.enabled else '# ',
                                                            job.slices, job.command, job.comment))
                if job.comment != job_comment or job.command != job_command or \
                    str(job.slices) != job_slices or job.enabled != job_enabled:
                    print(' --> 此任务条目发生了变化，准备变更任务条目')
                    job.set_comment(job_comment)
                    job.set_command(job_command)
                    job.setall(job_slices)
                    job.enable(job_enabled)
                    cron.write()
                    valids.append(job_label)
                    print(' --> 已变更此任务条目')
                else:
                    valids.append(job_label)
                    print(' --> 此任务条目没有发生变化，略过')
                break
            else:
                print(' --> 原任务条目 [无]')
                print(' --> 此任务条目之前不存在，准备创建新的任务条目')
                job = cron.new(command=job_command, comment=job_comment)
                job.setall(job_slices)
                job.enable(job_enabled)
                cron.write()
                valids.append(job_label)
                print(' --> 已创建新的任务条目')
        except Exception as e:
            errmsg = '未知错误, %s' % str(e)
            faileds.append('%s 更新失败：%s' % (job_label, errmsg))
            print(' --> %s' % errmsg)
        else:
            print(' --> 成功\n')

    if not jobs:
            print(' 没有需要更新的任务\n')

    print('# 更新任务结束\n\n')

def update_job_script(job_script, job_content):
    print(' --> 任务脚本 [%s]' % job_script)
    if not os.path.isfile(job_script):
        print(' --> 任务脚本之前不存在，准备创建新脚本')
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

def clear_jobs(jobs):
    print('# 清理任务\n')
    global removeds, faileds

    user_job_mp = {}
    for job in jobs:
        job_id = job['Id']
        job_user = job['Sysuser']
        if job_user not in user_job_mp:
             user_job_mp[job_user] = {}
        user_job_mp[job_user][job_id] = job

    tabnames = os.listdir(cron_dir)
    remove_jobs = []
    for tabname in tabnames:
        tabfile = os.path.join(cron_dir, tabname)
        cron = CronTab(tabfile=tabfile)
        job_iters = cron.find_command(re.compile(r'cronnest_job_.*.script'))
        for job in job_iters:
            job_id = int(job.command.split('.')[0].split('_')[-1])
            if not user_job_mp.get(tabname, {}).get(job_id):
                job_enabled = job.enabled
                job_slices = str(job.slices)
                job_command = job.command
                job_comment = job.comment
                remove_jobs.append((job_id, job_enabled, job_slices, job_command, job_comment, tabname, tabfile))

    for job_id, job_enabled, job_slices, job_command, job_comment, job_user, tabfile in remove_jobs:
        job_name = job_comment.split(':')[0]
        job_desc = ':'.join(job_comment.split(':')[1:])
        job_script = os.path.join(job_script_dir, 'cronnest_job_%d.script' % job_id)
        job_label = '[%s][%s] @ %s' % (job_name, job_desc, job_user)
        print(' 删除 %s' % job_label)
        try:
            cron = CronTab(tabfile=tabfile)
            job_iters = cron.find_command(job_command)
            for job in job_iters:
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
            faileds.append('%s 删除失败：%s' % (job_label, errmsg))
            print(' --> %s' % errmsg)
        else:
            removeds.append(job_label)
            print(' --> 成功\n')

    if not remove_jobs:
        print(' 没有需要删除的任务\n')

    print('# 清理任务结束\n\n')

def main():
    global valids, faileds
    update_jobs(jobs)
    clear_jobs(jobs)


    print('# 列出有效任务（%s）\n' % len(valids))
    if valids:
        for job in valids:
            print(' %s' % job)
        else:
            print('')
    else:
        print(' 没有有效的任务\n')
    print('# 列出有效任务结束\n\n')

    if faileds:
        print('# 本次操作异常（%s）\n' % len(faileds))
        for msg in faileds:
            print(' %s' % msg)
        exit(1)
    print("# 本次操作成功完成")
    exit(0)

main()