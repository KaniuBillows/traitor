initTable()


function initTable() {
    const TimingExecute = 0
    $('#table').bootstrapTable({
        url: '/api/jobList', method: 'get', classes: 'table table-bordered table-hover',  // bootstrap的表格样式
        cache: false, // 是否使用缓存，默认为true，一般来说需要设置一下这个属性
        pagination: true, // 是否显示分页
        sortable: true, // 是否启用排序
        sortOrder: 'asc', // 排序方式
        sidePaginnation: 'client', //分页方式，clien客户端分页，server服务端分页
        uniqueId: 'jobId', // 每一行的唯一标识，通常对应表的主键
        showToggle: false, // 是否显示详细试图和列表视图的切换按钮
        cardView: false, // 是否显示详细试图
        detailView: false, // 是否显示父子表
        columns: [
            {
                filed: 'checked', checkbox: true
            },
            {
                title: 'name', field: 'name', width: 100
            },
            {
                title: "execType", field: "execType", width: 50,
                formatter: (v, r, i) => {
                    return `<div>${v === TimingExecute ? "Timing" : "Delay"}</div>`
                }
            },
            {
                title: "description", field: "description", width: 300
            },
            {
                title: "lastExecTime", width: 50, field: "lastExecTime", formatter: (v, row, i) => {
                    return `<div>${v ?? "never"}</div>`
                }
            },
            {
                title: "nextExecTime", width: 50, formatter: (v, row, i) => {
                    if (row.execType === TimingExecute) {
                        try {
                            let futureMatches = cronMatcher.getFutureMatches(row.cron, {
                                startAt: 0,
                                timezone: 'UTC+' + (0 - new Date().getTimezoneOffset() / 60),
                                matchCount: 5,
                                hasSeconds: true,
                                formatInTimezone: true
                            })
                            if (futureMatches.length === 0) {
                                return '<div>never</div>'
                            }
                            return `<div>${new Date(futureMatches[0]).toLocaleString()}</div>`
                        } catch (err) {
                            return '<div> invalid cron</div>'
                        }
                    } else {
                        return `<div>${row.execAt}</div>`
                    }
                }
            },
            {
                field: '', title: 'operate', width: 200, formatter: (v, row, i) => {
                    let result = "<div style='display: flex;justify-content: space-evenly'>";
                    result += `<div class="form-check form-switch">
                         <input class="form-check-input" type="checkbox" style="height: 2em;    width: 3.5em;" role="switch" id="flexSwitchCheckDefault" ${row.jobState === 0 ? '' : 'checked'}>
                        </div>`
                    result += `<button type="button" class="btn btn-primary" data-bs-toggle="modal" data-bs-target="#editModal" onclick="edit('${i}')">Setting</button>`

                    result += `<button type="button" class="btn btn-primary" onclick="window.open('/edit/${row.jobId}','_blank')" >Edit</button>`

                    result += `<button type="button" class="btn btn-danger" onclick="remove('${row.jobId}')">Remove</button>`

                    result += "</div>"
                    return result
                }
            }]
    })
}

function remove(id) {
    $.ajax({
        type: "DELETE", url: `/api/job?id=${id}`, success: (res) => {
            navClick('job_list')
        }
    })
}


function edit(index) {
    let rows = $('#table').bootstrapTable('getData');
    let row = rows[index]
    $('#jobIdInput').val(row.jobId)
    $('#jobNameInput').val(row.name)
    $('#descriptionInput').val(row.description)
    $('#cronInput').val(row.cron)
}

function save() {
    let id = $('#jobIdInput').val()
    let createFlag = id === null || id === undefined || id === ""
    let job = {
        name: $('#jobNameInput').val(), description: $('#descriptionInput').val(), cron: $('#cronInput').val()
    }
    let url = createFlag ? '/api/job' : `/api/job?id=${id}`
    let method = createFlag ? 'POST' : 'PUT'

    $.ajax({
        url: url,
        type: method,
        data: JSON.stringify(job),
        contentType: 'application/json',
        dataType: 'json',
        success: () => {
            $('#editModal').modal('hide');
            $('#cronInput').attr("class", "form-control")
        },
    })
}

function cronCheck() {
    const input = $('#cronInput')
    const validCron = $('#validCron')
    let cron = input.val()
    try {
        cronParser.parse(cron, {hasSeconds: true})
    } catch (err) {
        input.attr("class", "form-control is-invalid")
        validCron.removeAttr("hidden")
        return
    }
    input.attr("class", "form-control is-valid")
    validCron.attr("hidden", "hidden")

}