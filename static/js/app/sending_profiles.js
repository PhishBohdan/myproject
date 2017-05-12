var profiles = []

// Attempts to send a test email by POSTing to /campaigns/
function sendTestEmail() {
    var test_email_request = {
        template: {},
        first_name: $("input[name=to_first_name]").val(),
        last_name: $("input[name=to_last_name]").val(),
        email: $("input[name=to_email]").val(),
        department: $("input[name=to_department]").val(),
        url: '',
        smtp: {
            from_address: $("#from").val(),
            host: $("#host").val(),
            username: $("#username").val(),
            password: $("#password").val(),
            ignore_cert_errors: $("#ignore_cert_errors").prop("checked")
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

// Save attempts to POST to /smtp/
function save(idx) {
    var profile = {}
    profile.name = $("#name").val()
    profile.interface_type = $("#interface_type").val()
    profile.from_address = $("#from").val()
    profile.host = $("#host").val()
    profile.username = $("#username").val()
    profile.password = $("#password").val()
    profile.ignore_cert_errors = $("#ignore_cert_errors").prop("checked")
    if (idx != -1) {
        profile.id = profiles[idx].id
        api.SMTPId.put(profile)
            .success(function(data) {
                successFlash("Profile edited successfully!")
                load()
                dismiss()
            })
            .error(function(data) {
                modalError(data.responseJSON.message)
            })
    } else {
        // Submit the profile
        api.SMTP.post(profile)
            .success(function(data) {
                successFlash("Profile added successfully!")
                load()
                dismiss()
            })
            .error(function(data) {
                modalError(data.responseJSON.message)
            })
    }
}

function dismiss() {
    $("#modal\\.flashes").empty()
    $("#name").val("")
    $("#interface_type").val("SMTP")
    $("#from").val("")
    $("#host").val("")
    $("#username").val("")
    $("#password").val("")
    $("#ignore_cert_errors").prop("checked", true)
    $("#modal").modal('hide')
}

function deleteProfile(idx) {
    if (confirm("Delete " + profiles[idx].name + "?")) {
        api.SMTPId.delete(profiles[idx].id)
            .success(function(data) {
                successFlash(data.message)
                load()
            })
    }
}

function edit(idx) {
    $("#modalSubmit").unbind('click').click(function() {
        save(idx)
    })
    var profile = {}
    if (idx != -1) {
        profile = profiles[idx]
        $("#name").val(profile.name)
        $("#interface_type").val(profile.interface_type)
        $("#from").val(profile.from_address)
        $("#host").val(profile.host)
        $("#username").val(profile.username)
        $("#password").val(profile.password)
        $("#ignore_cert_errors").prop("checked", profile.ignore_cert_errors)
    }
}

function copy(idx) {
    $("#modalSubmit").unbind('click').click(function() {
        save(-1)
    })
    var profile = {}
    profile = profiles[idx]
    $("#name").val("Copy of " + profile.name)
    $("#interface_type").val(profile.interface_type)
    $("#from").val(profile.from_address)
    $("#host").val(profile.host)
    $("#username").val(profile.username)
    $("#password").val(profile.password)
    $("#ignore_cert_errors").prop("checked", profile.ignore_cert_errors)
}

function load() {
    $("#profileTable").hide()
    $("#emptyMessage").hide()
    $("#loading").show()
    api.SMTP.get()
        .success(function(ss) {
            profiles = ss
            $("#loading").hide()
            if (profiles.length > 0) {
                $("#profileTable").show()
                profileTable = $("#profileTable").DataTable({
                    destroy: true,
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }]
                });
                profileTable.clear()
                $.each(profiles, function(i, profile) {
                    profileTable.row.add([
                        escapeHtml(profile.name),
                        profile.interface_type,
                        moment(profile.modified_date).format('MMMM Do YYYY, h:mm:ss a'),
                        "<div class='pull-right'><span data-toggle='modal' data-target='#modal'><button class='btn btn-primary' data-toggle='tooltip' data-placement='left' title='Edit Profile' onclick='edit(" + i + ")'>\
                    <i class='fa fa-pencil'></i>\
                    </button></span>\
		    <span data-toggle='modal' data-target='#modal'><button class='btn btn-primary' data-toggle='tooltip' data-placement='left' title='Copy Profile' onclick='copy(" + i + ")'>\
                    <i class='fa fa-copy'></i>\
                    </button></span>\
                    <button class='btn btn-danger' data-toggle='tooltip' data-placement='left' title='Delete Profile' onclick='deleteProfile(" + i + ")'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>"
                    ]).draw()
                })
                $('[data-toggle="tooltip"]').tooltip()
            } else {
                $("#emptyMessage").show()
            }
        })
        .error(function() {
            $("#loading").hide()
            errorFlash("Error fetching profiles")
        })
}

$(document).ready(function() {
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
    load()
})
