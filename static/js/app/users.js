var groups = []

// Save attempts to POST or PUT to /groups/
function save(idx) {
    var targets = []
    $.each($("#targetsTable").DataTable().rows().data(), function(i, target) {
        targets.push({
            first_name: unescapeHtml(target[0]),
            last_name: unescapeHtml(target[1]),
            email: unescapeHtml(target[2]),
            position: unescapeHtml(target[3]),
            department: unescapeHtml(target[4])
        })
    })
    var group = {
            name: $("#name").val(),
            targets: targets
        }
        // Submit the group
    if (idx != -1) {
        // If we're just editing an existing group,
        // we need to PUT /groups/:id
        group.id = groups[idx].id
        api.groupId.put(group)
            .success(function(data) {
                successFlash("Group updated successfully!")
                load()
                dismiss()
                $("#modal").modal('hide')
            })
            .error(function(data) {
                modalError(data.responseJSON.message)
            })
    } else {
        // Else, if this is a new group, POST it
        // to /groups
        api.groups.post(group)
            .success(function(data) {
                successFlash("Group added successfully!")
                load()
                dismiss()
                $("#modal").modal('hide')
            })
            .error(function(data) {
                modalError(data.responseJSON.message)
            })
    }
}

function dismiss() {
    $("#targetsTable").dataTable().DataTable().clear().draw()
    $("#name").val("")
    $("#modal\\.flashes").empty()
}

function edit(idx) {
    targets = $("#targetsTable").dataTable({
        destroy: true, // Destroy any other instantiated table - http://datatables.net/manual/tech-notes/3#destroy
        columnDefs: [{
            orderable: false,
            targets: "no-sort"
        }]
    })
    $("#modalSubmit").unbind('click').click(function() {
        save(idx)
    })
    if (idx == -1) {
        group = {}
    } else {
        $('ul.nav-tabs li:first-child').removeClass("active");
        $('ul.nav-tabs li:nth-child(2)').addClass("active");
        $('div#tab_0').removeClass("active");
        $('div#tab_1').addClass("active");
        group = groups[idx]
        $("#name").val(group.name)
        $.each(group.targets, function(i, record) {
            targets.DataTable()
                .row.add([
                    escapeHtml(record.first_name),
                    escapeHtml(record.last_name),
                    escapeHtml(record.email),
                    escapeHtml(record.position),
                    escapeHtml(record.department),
                    '<span style="cursor:pointer;"><i class="fa fa-trash-o"></i></span>'
                ]).draw()
        });
    }
    // Handle file uploads
    $("#csvupload").fileupload({
        dataType: "json",
        add: function(e, data) {
            $("#modal\\.flashes").empty()
            var acceptFileTypes = /(csv|txt)$/i;
            var filename = data.originalFiles[0]['name']
            if (filename && !acceptFileTypes.test(filename.split(".").pop())) {
                modalError("Unsupported file extension (use .csv or .txt)")
                return false;
            }
            data.submit();
        },
        done: function(e, data) {
            $.each(data.result, function(i, record) {
                addTarget(
                    record.first_name,
                    record.last_name,
                    record.email,
                    record.position,
                    record.department);
            });
            targets.DataTable().draw();
        }
    })
}

function deleteGroup(idx) {
    if (confirm("Delete " + groups[idx].name + "?")) {
        api.groupId.delete(groups[idx].id)
            .success(function(data) {
                successFlash(data.message)
                load()
            })
    }
}

function addTarget(firstNameInput, lastNameInput, emailInput, positionInput, departmentInput) {
    // Create new data row.
    var email = escapeHtml(emailInput).toLowerCase();
    var newRow = [
        escapeHtml(firstNameInput),
        escapeHtml(lastNameInput),
        email,
        escapeHtml(positionInput),
        escapeHtml(departmentInput),
        '<span style="cursor:pointer;"><i class="fa fa-trash-o"></i></span>'
    ];

    // Check table to see if email already exists.
    var targetsTable = targets.DataTable();
    var existingRowIndex = targetsTable
        .column(2, {
            order: "index"
        }) // Email column has index of 2
        .data()
        .indexOf(email);
    // Update or add new row as necessary.
    if (existingRowIndex >= 0) {
        targetsTable
            .row(existingRowIndex, {
                order: "index"
            })
            .data(newRow);
    } else {
        targetsTable.row.add(newRow);
    }
}

function load() {
    $("#groupTable").hide()
    $("#emptyMessage").hide()
    $("#loading").show()
    api.groups.get()
        .success(function(gs) {
            $("#loading").hide()
            if (gs.length > 0) {
                groups = gs
                $("#emptyMessage").hide()
                $("#groupTable").show()
                groupTable = $("#groupTable").DataTable({
                    destroy: true,
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }]
                });
                groupTable.clear();
                $.each(groups, function(i, group) {
                    var targets = ""
                    $.each(group.targets, function(i, target) {
                        targets += target.email + ", "
                        if (targets.length > 50) {
                            targets = targets.slice(0, -3) + "..."
                            return false;
                        }
                    })
                    groupTable.row.add([
                        escapeHtml(group.name),
                        escapeHtml(targets),
                        moment(group.modified_date).format('MMMM Do YYYY, h:mm:ss a'),
                        "<div class='text-center'><button class='btn btn-outline blue btn-sm ' onclick='edit(" + i + ")'>\
                    <i class='fa fa-pencil'></i>\
                    </button>\
                    <button class='btn btn-outline red btn-sm ' onclick='deleteGroup(" + i + ")'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                    ]).draw()
                })
            } else {
                $("#emptyMessage").show()
            }
        })
        .error(function() {
            errorFlash("Error fetching groups")
        })
}

$(document).ready(function() {
    load()
        // Setup the event listeners
        // Handle manual additions
    $("#targetForm").submit(function() {
        addTarget(
            $("#firstName").val(),
            $("#lastName").val(),
            $("#email").val(),
            $("#position").val(),
            $("#department").val());
        targets.DataTable().draw();

        // Reset user input.
        $("#targetForm>div>input").val('');
        $("#firstName").focus();
        return false;
    });
    // Handle Deletion
    $("#targetsTable").on("click", "span>i.fa-trash-o", function() {
        targets.DataTable()
            .row($(this).parents('tr'))
            .remove()
            .draw();
    });
    $("#modal").on("hide.bs.modal", function() {
        dismiss();
    });
});
