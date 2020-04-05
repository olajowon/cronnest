new Vue({
    el: '#app',
    delimiters: ['${', '}'],
    data: {
        mouseOver: null,
        pageName: "crontab",
        locking: null,
        tree: {
            name: 'Crontab',
            open: true,
            hostgroups: []
        },
        queryParams: {hostgroup: null, host: null, tab: "", row: null},
        currHost: {},
        openUpdateTabJobModal: true,
        updateJobData: {host_id: null, tab: "", old_job: {}, new_job: {}},
        createJobData: {host_id: null, tab: "", job: {}},
        createHostgroupData: {name: ""},
        updateHostgroupData: {hostgroup: null, data: {name: ""}},
        addHostgroupHostsData: {hostgroup: null, data: {hosts: ""}},
        removeHostgroupHostsData: {hostgroup: null, data: {hosts: []}},
        updateHostData: {host: null, data: {address: ""}},
    },
    watch: {
        "currHost.id": function(n, o) {
            if (n > 0) {
                this.getHostCrontab(this.currHost)
            }
        },
        locking: function (n, o) {
            if (n != null && n != "") {
                $("#lockingModal").modal('show')
            } else {
                $("#lockingModal").modal('hide')
            }
        }
    },
    mounted: function () {
        this.initQueryParams();
        this.init();
    },
    methods: {
        init() {
            let _this = this;
            axios.get("/api/hostgroups/")
                .then(function(resp){
                    let hostgroups = resp.data.data;
                    hostgroups.forEach(function (hg, _) {
                        if (hg.id == _this.queryParams.hostgroup) {
                            hg.open = true;
                            hg.hosts = [];
                            axios.get("/api/hostgroups/"+ hg.id +"/hosts/")
                                .then(function(resp){
                                    let hosts = resp.data.data;
                                    hg.hosts = hosts;
                                    hg.hosts.forEach(function (h, _) {
                                        if (h.id == _this.queryParams.host) {
                                            _this.currHost = $.extend({crontab: {}, hostgroup: hg.id}, h);
                                        }
                                    })
                                })
                        } else {
                            hg.open = false;
                            hg.hosts = [];
                        }
                    })
                    _this.tree.hostgroups = hostgroups;
                })
        },

        initQueryParams() {
            for (let k in this.queryParams) {
                let v = this.getQueryVariable(k);
                if (v) {
                    this.queryParams[k] = decodeURI(v);
                } else {
                    this.queryParams[k] = null;
                }
            }
        },

        getHostCrontab(h) {
            let _this = this
            this.locking = "正在获取Crontab..."
            axios.get("/api/hosts/"+ h.id +"/crontab/")
                .then(function(resp){
                    let crontab = resp.data.data;

                    if (crontab.tab) {
                        crontab.tab = _this.makeNewTab(crontab.tab)
                    }
                    h.crontab = crontab;
                    _this.locking = null;
                })
        },

        makeNewTab(tab) {
            let newtab = {};
            if (tab.root) {
                newtab.root = tab.root;
            }
            for (let k in tab) {
                if (k!='system' && k!='root') {
                    newtab[k] = tab[k];
                }
            }
            if (tab.system) {
                newtab.system = tab.system;
            }
            return newtab
        },

        getHostgroupHost: function(hg) {
            let _this = this;
            axios.get("/api/hostgroups/"+ hg.id +"/hosts/")
                .then(function(resp){
                    let hosts = resp.data.data;
                    hg.hosts = hosts;
                })
        },

        handleHostgroupNodeClick(hg) {
            hg.open = !hg.open;
            if (hg.open == true && hg.hosts.length == 0) {
                this.getHostgroupHost(hg)
            }
        },

        handleHostClick(h, hg) {
            if (h.id != this.currHost.id) {
                this.currHost = $.extend({crontab: {}, hostgroup: hg.id}, h);
                let href = window.location.pathname + "?hostgroup=" + hg.id + "&host=" + h.id;
                window.history.pushState({}, null, href);
                this.queryParams = {
                    hostgoups: hg.id,
                    host: h.id,
                    crontab: null,
                    row: null
                }
            }
        },

        getQueryVariable: function (variable) {
            let query = window.location.search.substring(1);
            let vars = query.split("&");
            for (let i=0;i<vars.length;i++) {
                let pair = vars[i].split("=");
                if(pair[0] == variable){return pair[1];}
            }
            return false;
        },

        createJob(hostId, tab) {
            this.createJobData.host_id = hostId,
                this.createJobData.tab = tab,
                this.createJobData.job = {
                    enabled: true,
                    slices: "",
                    command: "",
                    comment: "",
                }
            if (tab == "system") {
                this.createJobData.job.user = "";
            }

            $("#createJobModal").modal("show");
        },

        submitCreateJob(createJobData) {
            let hostId = createJobData.host_id;
            let data = {
                tab: createJobData.tab,
                job: createJobData.job,
            }
            let _this = this;
            _this.locking = "正在创建主机Crontab Job，请稍等...";
            axios.post("/api/hosts/"+ hostId +"/crontab/job/", data)
                .then(function(resp){
                    if (_this.currHost.id = createJobData.host_id) {
                        _this.currHost.crontab.tab[createJobData.tab].push(createJobData.job)
                    }
                    _this.locking = null;
                    toastr.success("创建成功")
                    $("#createJobModal").modal("hide");
                })
                .catch(function(error){
                    _this.locking = null;
                    if (error.response && error.response.data && error.response.data.msg) {
                        toastr.error(error.response.data.msg.replace(/\n/g,"<br>"))
                    } else {
                        toastr.error(error)
                    }
                });
        },

        updateJob(hostId, tab, job) {
            this.updateJobData.host_id = hostId;
            this.updateJobData.tab = tab;
            this.updateJobData.old_job = job;
            this.updateJobData.new_job =  {
                enabled: job.enabled,
                slices: job.slices,
                command: job.command,
                comment: job.comment,
            };
            if (tab == "system") {
                this.updateJobData.new_job.user = job.user;
            }

            $("#updateJobModal").modal("show");
        },

        submitUpdateJob(updateJobData) {
            let hostId = updateJobData.host_id;
            let data = {
                tab: updateJobData.tab,
                old_job: updateJobData.old_job,
                new_job: updateJobData.new_job
            }
            let _this = this;
            _this.locking = "正在修改主机Crontab Job，请稍等...";
            axios.put("/api/hosts/"+ hostId +"/crontab/job/", data)
                .then(function(resp){
                    updateJobData.old_job.enabled = updateJobData.new_job.enabled;
                    updateJobData.old_job.slices = updateJobData.new_job.slices;
                    updateJobData.old_job.command = updateJobData.new_job.command;
                    updateJobData.old_job.comment = updateJobData.new_job.comment;
                    if (updateJobData.tab == "system") {
                        updateJobData.old_job.user = updateJobData.new_job.user;
                    }
                    _this.locking = null;
                    toastr.success("修改成功")
                    $("#updateJobModal").modal("hide");
                })
                .catch(function(error){
                    _this.locking = null;
                    if (error.response && error.response.data && error.response.data.msg) {
                        toastr.error(error.response.data.msg.replace(/\n/g,"<br>"))
                    } else {
                        toastr.error(error)
                    }
                });
        },

        deleteJob(hostId, tab, job) {
            let r = confirm("删除后不可恢复，确定删除此Job？");
            if(r == true) {
                let data = {
                    tab: tab,
                    job: job,
                }
                let _this = this;
                _this.locking = "正在删除主机Crontab Job，请稍等...";
                axios.delete("/api/hosts/"+ hostId +"/crontab/job/", {data: data})
                    .then(function(resp){
                        if (_this.currHost.id = hostId) {
                            _this.currHost.crontab.tab = _this.makeNewTab(resp.data.data.tab)
                        }
                        _this.locking = null;
                        toastr.success("删除成功")
                    })
                    .catch(function(error){
                        _this.locking = null;
                        if (error.response && error.response.data && error.response.data.msg) {
                            toastr.error(error.response.data.msg.replace(/\n/g,"<br>"))
                        } else {
                            toastr.error(error)
                        }
                    });
            }
        },

        createHostgroup() {
            this.createHostgroupData.name = "";
            $("#createHostgroupModal").modal("show");
        },

        submitCreateHostgroup(createHgData) {
            let _this = this;
            _this.locking = "正在创建主机组，请稍等...";
            axios.post("/api/hostgroups/", createHgData)
                .then(function(resp){
                    let hg = resp.data.data;
                    hg.open = false;
                    hg.hosts = [];
                    _this.tree.hostgroups.push(hg)
                    _this.locking = null;
                    toastr.success("创建成功")
                    $("#createHostgroupModal").modal("hide");
                })
                .catch(function(error){
                    _this.locking = null;
                    if (error.response && error.response.data && error.response.data.msg) {
                        toastr.error(error.response.data.msg.replace(/\n/g,"<br>"))
                    } else {
                        toastr.error(error)
                    }
                });
        },

        updateHostgroup(hg) {
            this.updateHostgroupData.hostgroup = hg;
            this.updateHostgroupData.data.name = hg.name;
            $("#updateHostgroupModal").modal("show");
        },

        submitUpdateHostgroup(updateHgData) {
            let _this = this;
            _this.locking = "正在修改主机组，请稍等...";
            axios.put("/api/hostgroups/" + updateHgData.hostgroup.id + "/", updateHgData.data)
                .then(function(resp){
                    updateHgData.hostgroup.name = updateHgData.data.name;

                    _this.locking = null;
                    toastr.success("修改成功")
                    $("#updateHostgroupModal").modal("hide");
                })
                .catch(function(error){
                    _this.locking = null;
                    if (error.response && error.response.data && error.response.data.msg) {
                        toastr.error(error.response.data.msg.replace(/\n/g,"<br>"))
                    } else {
                        toastr.error(error)
                    }
                });
        },

        deleteHostgroup(hg, idx) {
            let r = confirm("删除后不可恢复，确定删除此主机组？\n注：仅属于该主机组的主机将会被一并删除！");
            if(r == true) {
                let _this = this;
                _this.locking = "正在删除主机组，请稍等...";
                axios.delete("/api/hostgroups/"+ hg.id +"/")
                    .then(function(resp){
                        _this.tree.hostgroups.splice(idx, 1);
                        _this.locking = null;
                        toastr.success("删除成功")
                    })
                    .catch(function(error){
                        _this.locking = null;
                        if (error.response && error.response.data && error.response.data.msg) {
                            toastr.error(error.response.data.msg.replace(/\n/g,"<br>"))
                        } else {
                            toastr.error(error)
                        }
                    });
            }
        },

        addHostgroupHosts(hg) {
            this.addHostgroupHostsData.hostgroup = hg;
            this.addHostgroupHostsData.data.hosts = "";
            $("#addHostgroupHostsModal").modal("show")
        },

        submitAddHostgroupHosts(addHgHostData) {
            let _this = this;
            _this.locking = "正在为主机组添加主机，请稍等...";
            console.log(addHgHostData.data.hosts.split('\n'))
            axios.post("/api/hostgroups/"+ addHgHostData.hostgroup.id +"/hosts/",
                {hosts: addHgHostData.data.hosts.split('\n')})
                .then(function(resp){
                    _this.locking = null;
                    toastr.success("添加成功")
                    $("#addHostgroupHostsModal").modal("hide");
                    addHgHostData.hostgroup.open = true;
                    _this.getHostgroupHost(addHgHostData.hostgroup);
                })
                .catch(function(error){
                    _this.locking = null;
                    if (error.response && error.response.data && error.response.data.msg) {
                        toastr.error(error.response.data.msg.replace(/\n/g,"<br>"))
                    } else {
                        toastr.error(error)
                    }
                });
        },

        deleteHost(h) {
            let r = confirm("删除后不可恢复，确定删除此主机？\n注：其他组内的该主机也会被一并删除！");
            if(r == true) {
                let _this = this;
                _this.locking = "正在删除主机，请稍等...";
                axios.delete("/api/hosts/"+ h.id +"/")
                    .then(function(resp){
                        for (let i=0; i<_this.tree.hostgroups.length; i++) {
                            for (let j=_this.tree.hostgroups[i].hosts.length-1; j>=0; j--) {
                                if (_this.tree.hostgroups[i].hosts[j].id == h.id) {
                                    _this.tree.hostgroups[i].hosts.splice(j, 1);
                                }
                            }
                        }

                        _this.locking = null;
                        toastr.success("删除成功")
                    })
                    .catch(function(error){
                        _this.locking = null;
                        if (error.response && error.response.data && error.response.data.msg) {
                            toastr.error(error.response.data.msg.replace(/\n/g,"<br>"))
                        } else {
                            toastr.error(error)
                        }
                    });
            }
        },

        updateHost(h) {
            this.updateHostData.host = h;
            this.updateHostData.data.address = h.address;
            $("#updateHostModal").modal("show");
        },

        submitUpdateHost(updateData) {
            let _this = this;
            _this.locking = "正在修改主机，请稍等...";
            axios.put("/api/hosts/" + updateData.host.id + "/", updateData.data)
                .then(function(resp){
                    for (let i=0; i<_this.tree.hostgroups.length; i++) {
                        for (let j=0; j<_this.tree.hostgroups[i].hosts.length; j++) {
                            if (_this.tree.hostgroups[i].hosts[j].id == updateData.host.id) {
                                _this.tree.hostgroups[i].hosts[j].address = updateData.data.address;
                            }
                        }
                    }
                    _this.locking = null;
                    toastr.success("修改成功")
                    $("#updateHostModal").modal("hide");
                })
                .catch(function(error){
                    _this.locking = null;
                    if (error.response && error.response.data && error.response.data.msg) {
                        toastr.error(error.response.data.msg.replace(/\n/g,"<br>"))
                    } else {
                        toastr.error(error)
                    }
                });
        },

        removeHostgroupHosts(hg) {
            this.removeHostgroupHostsData.hostgroup = hg;
            this.removeHostgroupHostsData.data.hosts = [];
            let _this = this;
            axios.get("/api/hostgroups/"+ hg.id +"/hosts/")
                .then(function(resp){
                    let hosts = resp.data.data;
                    for (let i=0; i<hosts.length; i++) {
                        _this.removeHostgroupHostsData.data.hosts.push([hosts[i].address, false])
                    }
                    $("#removeHostgroupHostsModal").modal("show")
                })

        },
        submitRemoveHostgroupHosts(removeHgHostData) {
            console.log(removeHgHostData)
            let removeHosts = [];
            for (let i=0; i<removeHgHostData.data.hosts.length; i++) {
                if (removeHgHostData.data.hosts[i][1] == true) {
                    removeHosts.push(removeHgHostData.data.hosts[i][0])
                }
            }

            let _this = this;
            _this.locking = "正在从主机组移除主机，请稍等...";
            axios.delete("/api/hostgroups/" + removeHgHostData.hostgroup.id + "/hosts/", {data: {hosts: removeHosts}})
                .then(function(resp){
                    for (let i=removeHgHostData.hostgroup.hosts.length-1; i>=0; i--) {
                        if (removeHosts.indexOf(removeHgHostData.hostgroup.hosts[i].address) >= 0) {
                            removeHgHostData.hostgroup.hosts.splice(i, 1);
                        }
                    }

                    _this.locking = null;
                    toastr.success("移除成功")
                    $("#removeHostgroupHostsModal").modal("hide");
                })
                .catch(function(error){
                    _this.locking = null;
                    if (error.response && error.response.data && error.response.data.msg) {
                        toastr.error(error.response.data.msg.replace(/\n/g,"<br>"))
                    } else {
                        toastr.error(error)
                    }
                });
        }
    }
})
