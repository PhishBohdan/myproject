    // labels is a map of campaign statuses to
// CSS classes
var labels = {
    "In progress": "label-primary",
    "Queued": "label-info",
    "Completed": "label-success",
    "Emails Sent": "label-success",
    "Error": "label-danger"
}

var campaigns = []
var campaign = {}

// Launch attempts to POST to /campaigns/
function launch() {
    swal({
        title: "Are you sure?",
        text: "This will schedule the campaign to be launched.",
        type: "question",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Launch",
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        showLoaderOnConfirm: true,
        preConfirm: function() {
            return new Promise(function(resolve, reject) {
                groups = []
                $.each($("#groupTable").DataTable().rows().data(), function(i, group) {
                    groups.push({
                        name: unescapeHtml(group[0])
                    })
                })
                campaign = {
                        name: $("#name").val(),
                        template: {
                            name: $("#template").val()
                        },
                        url: $("#url").val(),
                        page: {
                            name: $("#page").val()
                        },
                        smtp: {
                            name: $("#profile").val()
                        },
                        launch_date: moment($("#launch_date").val(), "MM/DD/YYYY hh:mm a").format(),
                        groups: groups
                    }
                    // Submit the campaign
                api.campaigns.post(campaign)
                    .success(function(data) {
                        resolve()
                        campaign = data
                    })
                    .error(function(data) {
                        $("#modal\\.flashes").empty().append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
            <i class=\"fa fa-exclamation-circle\"></i> " + data.responseJSON.message + "</div>")
                        swal.close()
                    })
            })
        }
    }).then(function() {
        swal(
            'Campaign Scheduled!',
            'This campaign has been scheduled for launch!',
            'success'
        );
        $('button:contains("OK")').on('click', function() {
            window.location = "/campaigns/" + campaign.id.toString()
        })
    })
}

// Attempts to send a test email by POSTing to /campaigns/
function sendTestEmail() {
    var test_email_request = {
        template: {
            name: $("#template").val()
        },
        first_name: $("input[name=to_first_name]").val(),
        last_name: $("input[name=to_last_name]").val(),
        email: $("input[name=to_email]").val(),
        position: $("input[name=to_position]").val(),
        department: $("input[name=to_department]").val(),
        url: $("#url").val(),
        page: {
            name: $("#page").val()
        },
        smtp: {
            name: $("#profile").val()
        }
    }
    btnHtml = $("#sendTestModalSubmit").html()
    $("#sendTestModalSubmit").html('<i class="fa fa-spinner fa-spin"></i> Sending')
        // Send the test email
    api.send_test_email(test_email_request)
        .success(function(data) {
            $("#sendTestEmailModal\\.flashes").empty().append("<div style=\"text-align:center\" class=\"alert alert-success\">\
            <i class=\"fa fa-check-circle\"></i> Email Sent!</div>")
            $("#sendTestModalSubmit").html(btnHtml)
        })
        .error(function(data) {
            $("#sendTestEmailModal\\.flashes").empty().append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
            <i class=\"fa fa-exclamation-circle\"></i> " + data.responseJSON.message + "</div>")
            $("#sendTestModalSubmit").html(btnHtml)
        })
}

function dismiss() {
    $("#modal\\.flashes").empty()
    $("#name").val("")
    $("#template").val("")
    $("#page").val("")
    $("#url").val("")
    $("#profile").val("")
    $("#groupSelect").val("")
    $("#modal").modal('hide')
    $("#groupTable").dataTable().DataTable().clear().draw()
}

function deleteCampaign(idx) {
    swal({
        title: "Are you sure?",
        text: "This will delete the campaign. This can't be undone!",
        type: "warning",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Delete " + campaigns[idx].name,
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        preConfirm: function() {
            return new Promise(function(resolve, reject) {
                api.campaignId.delete(campaigns[idx].id)
                    .success(function(msg) {
                        resolve()
                    })
                    .error(function(data) {
                        reject(data.responseJSON.message)
                    })
            })
        }
    }).then(function() {
        swal(
            'Campaign Deleted!',
            'This campaign has been deleted!',
            'success'
        );
        $('button:contains("OK")').on('click', function() {
            location.reload()
        })
    })
}

function edit(campaign) {
    // Clear the bloodhound instance
    group_bh.clear();
    template_bh.clear();
    page_bh.clear();
    profile_bh.clear();
    if (campaign == "new") {
        api.groups.get()
            .success(function(groups) {
                if (groups.length == 0) {
                    modalError("No groups found!")
                    return false;
                } else {
                    group_bh.add(groups)
                }
            })
        api.templates.get()
            .success(function(templates) {
                if (templates.length == 0) {
                    modalError("No templates found!")
                    return false
                } else {
                    template_bh.add(templates)
                }
            })
        api.pages.get()
            .success(function(pages) {
                if (pages.length == 0) {
                    modalError("No pages found!")
                    return false
                } else {
                    page_bh.add(pages)
                }
            })
        api.SMTP.get()
            .success(function(profiles) {
                if (profiles.length == 0) {
                    modalError("No profiles found!")
                    return false
                } else {
                    profile_bh.add(profiles)
                }
            })
    }
}

function copy(idx) {
    group_bh.clear();
    template_bh.clear();
    page_bh.clear();
    profile_bh.clear();
    api.groups.get()
        .success(function(groups) {
            if (groups.length == 0) {
                modalError("No groups found!")
                return false;
            } else {
                group_bh.add(groups)
            }
        })
    api.templates.get()
        .success(function(templates) {
            if (templates.length == 0) {
                modalError("No templates found!")
                return false
            } else {
                template_bh.add(templates)
            }
        })
    api.pages.get()
        .success(function(pages) {
            if (pages.length == 0) {
                modalError("No pages found!")
                return false
            } else {
                page_bh.add(pages)
            }
        })
    api.SMTP.get()
        .success(function(profiles) {
            if (profiles.length == 0) {
                modalError("No profiles found!")
                return false
            } else {
                profile_bh.add(profiles)
            }
        })
        // Set our initial values
    var campaign = campaigns[idx]
    $("#name").val("Copy of " + campaign.name)
    $("#template").val(campaign.template.name)
    $("#page").val(campaign.page.name)
    $("#profile").val(campaign.smtp.name)
    $("#url").val(campaign.url)
}

function processCampaignsLimit(campaignTable, campaigns) {
    $.each(campaigns, function(i, campaign) {
        label = labels[campaign.status] || "label-default";
        campaignsahref = '/campaigns/';
        copycampaign = "<span data-toggle='modal' data-target='#modal'><button class='btn btn-outline purple btn-sm ' data-toggle='tooltip' data-placement='left' title='Copy Campaign' onclick='copy(" + i + ")'>\
        <i class='fa fa-copy'></i>\
        </button></span>";
        deletecampaign = "<button class='btn btn-outline red btn-sm ' onclick='deleteCampaign(" + i + ")' data-toggle='tooltip' data-placement='left' title='Delete Campaign'>\
        <i class='fa fa-trash-o'></i>\
        </button></div>";
        if (typeof restricteduser !== 'undefined') {
            campaignsahref = '/campaigns/';
            copycampaign = "";
            deletecampaign = "";
        }
        campaignTable.row.add([
            escapeHtml(campaign.name),
            moment(campaign.created_date).format('MMMM Do YYYY, h:mm:ss a'),
            "<span class=\"label " + label + "\">" + campaign.status + "</span>",
            "<div class='text-center'><a class='btn btn-outline blue btn-sm ' href='" + campaignsahref + campaign.id + "' data-toggle='tooltip' data-placement='left' title='View Results'>\
        <i class='fa fa-bar-chart'></i>\
        </a>" + copycampaign + deletecampaign
        ]).draw()
        $('[data-toggle="tooltip"]').tooltip()
    });
}

$(document).ready(function() {
    $("#launch_date").datetimepicker({
            "widgetPositioning": {
                "vertical": "bottom"
            },
            "showTodayButton": true,
            "defaultDate": moment()
    })
        // Setup multiple modals
        // Code based on http://miles-by-motorcycle.com/static/bootstrap-modal/index.html
    $('.modal').on('hidden.bs.modal', function(event) {
        $(this).removeClass('fv-modal-stack');
        $('body').data('fv_open_modals', $('body').data('fv_open_modals') - 1);
    });
    $('.modal').on('shown.bs.modal', function(event) {
        // Keep track of the number of open modals
        if (typeof($('body').data('fv_open_modals')) == 'undefined') {
            $('body').data('fv_open_modals', 0);
        }
        // if the z-index of this modal has been set, ignore.
        if ($(this).hasClass('fv-modal-stack')) {
            return;
        }
        $(this).addClass('fv-modal-stack');
        // Increment the number of open modals
        $('body').data('fv_open_modals', $('body').data('fv_open_modals') + 1);
        // Setup the appropriate z-index
        $(this).css('z-index', 1040 + (10 * $('body').data('fv_open_modals')));
        $('.modal-backdrop').not('.fv-modal-stack').css('z-index', 1039 + (10 * $('body').data('fv_open_modals')));
        $('.modal-backdrop').not('fv-modal-stack').addClass('fv-modal-stack');
    });
    $.fn.modal.Constructor.prototype.enforceFocus = function() {
        $(document)
            .off('focusin.bs.modal') // guard against infinite focus loop
            .on('focusin.bs.modal', $.proxy(function(e) {
                if (
                    this.$element[0] !== e.target && !this.$element.has(e.target).length
                    // CKEditor compatibility fix start.
                    && !$(e.target).closest('.cke_dialog, .cke').length
                    // CKEditor compatibility fix end.
                ) {
                    this.$element.trigger('focus');
                }
            }, this));
    };
    $('#modal').on('hidden.bs.modal', function(event) {
        dismiss()
    });
    var initialLimit = 20;
    api.campaignslimit.get(initialLimit)
        .success(function(cs) {
            console.log(cs);
            campaigns = cs
            $("#loading").hide()
            if (campaigns.length > 0) {
                $("#campaignTable").show()
                campaignTable = $("#campaignTable").DataTable({
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }]
                });
                processCampaignsLimit(campaignTable, campaigns);
                api.campaignslimit.get(0)
                    .success(function(allCS) {
                        processCampaignsLimit(campaignTable, allCS.slice(initialLimit, allCS.length));
                    })
                    .error(function() {
                        $("#loading").hide()
                        errorFlash("Error fetching campaigns")
                    });

            } else {
                $("#emptyMessage").show()
            }
        })
        .error(function() {
            $("#loading").hide()
            errorFlash("Error fetching campaigns")
        })
    $("#groupForm").submit(function() {
        // Add row to group table.
        var newRow = groupTable.row.add([
            escapeHtml($("#groupSelect").val()),
            '<span style="cursor:pointer;"><i class="fa fa-trash-o"></i></span>'
        ]).draw().node();

        // Set event handler for removing row from group table.
        $(newRow).on("click", "span>i.fa-trash-o", function() {
            groupTable.row($(this).parents('tr'))
                .remove()
                .draw();
        });

        // Clear user input.
        $("#groupSelect").typeahead('val', "");
        return false;
    });
    // Create the group typeahead objects
    groupTable = $("#groupTable").DataTable({
        columnDefs: [{
            orderable: false,
            targets: "no-sort"
        }]
    })
    group_bh = new Bloodhound({
        datumTokenizer: function(g) {
            return Bloodhound.tokenizers.whitespace(g.name)
        },
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        local: []
    })
    group_bh.initialize()
    $("#groupSelect.typeahead.form-control").typeahead({
            hint: true,
            highlight: true,
            minLength: 1
        }, {
            name: "groups",
            source: group_bh,
            templates: {
                empty: function(data) {
                    return '<div class="tt-suggestion">No groups matched that query</div>'
                },
                suggestion: function(data) {
                    return '<div>' + escapeHtml(data.name) + '</div>'
                }
            }
        })
        .bind('typeahead:select', function(ev, group) {
            // Add selected group.
            $("#groupSelect").typeahead('val', group.name);
            $("#groupForm").submit();
        })
        .bind('typeahead:autocomplete', function(ev, group) {
            $("#groupSelect").typeahead('val', group.name)
        });
    // Create the template typeahead objects
    template_bh = new Bloodhound({
        datumTokenizer: function(t) {
            return Bloodhound.tokenizers.whitespace(t.name)
        },
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        local: []
    })
    template_bh.initialize()
    $("#template.typeahead.form-control").typeahead({
            hint: true,
            highlight: true,
            minLength: 1
        }, {
            name: "templates",
            source: template_bh,
            templates: {
                empty: function(data) {
                    return '<div class="tt-suggestion">No templates matched that query</div>'
                },
                suggestion: function(data) {
                    return '<div>' + escapeHtml(data.name) + '</div>'
                }
            }
        })
        .bind('typeahead:select', function(ev, template) {
            $("#template").typeahead('val', template.name)
        })
        .bind('typeahead:autocomplete', function(ev, template) {
            $("#template").typeahead('val', template.name)
        });
    // Create the landing page typeahead objects
    page_bh = new Bloodhound({
        datumTokenizer: function(p) {
            return Bloodhound.tokenizers.whitespace(p.name)
        },
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        local: []
    })
    page_bh.initialize()
    $("#page.typeahead.form-control").typeahead({
            hint: true,
            highlight: true,
            minLength: 1
        }, {
            name: "pages",
            source: page_bh,
            templates: {
                empty: function(data) {
                    return '<div class="tt-suggestion">No pages matched that query</div>'
                },
                suggestion: function(data) {
                    return '<div>' + escapeHtml(data.name) + '</div>'
                }
            }
        })
        .bind('typeahead:select', function(ev, page) {
            $("#page").typeahead('val', page.name)
        })
        .bind('typeahead:autocomplete', function(ev, page) {
            $("#page").typeahead('val', page.name)
        });
    // Create the sending profile typeahead objects
    profile_bh = new Bloodhound({
        datumTokenizer: function(s) {
            return Bloodhound.tokenizers.whitespace(s.name)
        },
        queryTokenizer: Bloodhound.tokenizers.whitespace,
        local: []
    })
    profile_bh.initialize()
    $("#profile.typeahead.form-control").typeahead({
            hint: true,
            highlight: true,
            minLength: 1
        }, {
            name: "profiles",
            source: profile_bh,
            templates: {
                empty: function(data) {
                    return '<div class="tt-suggestion">No profiles matched that query</div>'
                },
                suggestion: function(data) {
                    return '<div>' + escapeHtml(data.name) + '</div>'
                }
            }
        })
        .bind('typeahead:select', function(ev, profile) {
            $("#profile").typeahead('val', profile.name)
        })
        .bind('typeahead:autocomplete', function(ev, profile) {
            $("#profile").typeahead('val', profile.name)
        });
})
