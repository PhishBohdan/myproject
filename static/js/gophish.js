function errorFlash(message) {
    $("#flashes").empty()
    $("#flashes").append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
        <i class=\"fa fa-exclamation-circle\"></i> " + message + "</div>")
}

function successFlash(message) {
    $("#flashes").empty()
    $("#flashes").append("<div style=\"text-align:center\" class=\"alert alert-success\">\
        <i class=\"fa fa-check-circle\"></i> " + message + "</div>")
}

function modalError(message) {
    $("#modal\\.flashes").empty().append("<div style=\"text-align:center\" class=\"alert alert-danger\">\
        <i class=\"fa fa-exclamation-circle\"></i> " + message + "</div>")
}

function query(endpoint, method, data, async) {
    return $.ajax({
        url: "/api" + endpoint + "?api_key=" + user.api_key,
        async: true,
        method: method,
        data: JSON.stringify(data),
        dataType: "json",
        contentType: "application/json"
    })
}

function queryJSON(endpoint, method, data, async) {
    return $.ajax({
        url: "/api" + endpoint + "?api_key=" + user.api_key,
        async: true,
        method: method,
        data: data,
        dataType: "json",
        contentType: "application/json"
    })
}

function escapeHtml(text) {
    return $("<div/>").text(text).html()
}

function unescapeHtml(html) {
    return $("<div/>").html(html).text()
}

/*
Define our API Endpoints
*/
var api = {
    // campaigns contains the endpoints for /campaigns
    campaigns: {
        // get() - Queries the API for GET /campaigns
        get: function() {
            return query("/campaigns/", "GET", {}, false)
        },
        // post() - Posts a campaign to POST /campaigns
        post: function(data) {
            return query("/campaigns/", "POST", data, false)
        }
    },

    // campaignslimit contains the endpoints for /campaignslimit
    campaignslimit: {
        // get() - Queries the API for GET /campaignslimit
        get: function(max) {
            return query("/campaignslimit/" + max, "GET", {}, false)
        },
    },
    // campaignsrange contains the endpoints for /campaignsrange
    campaignsrange: {
        // get() - Queries the API for GET /campaigns
        get: function(min, max, filter, matchExact) {
            return queryJSON("/campaignsrange/" + min + '-' + max, "GET", {'filter' : filter, 'matchExact' : matchExact}, false)
        },
    },
    // campaignssummarystats contains the endpoints for /campaignssummarystats
    campaignssummarystats: {
        // get() - Queries the API for GET /campaigns
        get: function(filter, matchExact) {
            return queryJSON("/campaignssummarystats/", "GET", {'filter' : filter, 'matchExact' : matchExact}, false)
        },
    },
    // campaignnames contains the endpoints for /campaignnames
    campaignnames: {
        // get() - Queries the API for GET /campaigns
        get: function(max) {
            return query("/campaignnames/" + max, "GET", {}, false)
        },
    },
    // campaignId contains the endpoints for /campaigns/:id
    campaignId: {
        // get() - Queries the API for GET /campaigns/:id
        get: function(id) {
            return query("/campaigns/" + id, "GET", {}, true)
        },
        // delete() - Deletes a campaign at DELETE /campaigns/:id
        delete: function(id) {
            return query("/campaigns/" + id, "DELETE", {}, false)
        },
        // results() - Queries the API for GET /campaigns/:id/results
        results: function(id) {
            return query("/campaigns/" + id + "/results", "GET", {}, true)
        },
        // complete() - Completes a campaign at POST /campaigns/:id/complete
        complete: function(id) {
            return query("/campaigns/" + id + "/complete", "GET", {}, true)
        }
    },
    // phishingresults contains the endpoints for /phishingresults
    phishingresults: {
        // get() - Queries the API for GET /targetresults
        get: function(filter, matchExact, status) {
            return queryJSON("/phishingresults/", "GET", {'filter' : filter, 'matchExact' : matchExact, 'status': status}, false)
        },
    },

    targetevents: {
        // get() - Queries the API for GET /targetevents
        get: function(email) {
            return queryJSON("/targetevents/", "GET", {'email' : email}, false)
        }
    },
    // groups contains the endpoints for /groups
    groups: {
        // get() - Queries the API for GET /groups
        get: function() {
            return query("/groups/", "GET", {}, false)
        },
        // post() - Posts a group to POST /groups
        post: function(group) {
            return query("/groups/", "POST", group, false)
        }
    },
    // groupId contains the endpoints for /groups/:id
    groupId: {
        // get() - Queries the API for GET /groups/:id
        get: function(id) {
            return query("/groups/" + id, "GET", {}, false)
        },
        // put() - Puts a group to PUT /groups/:id
        put: function(group) {
            return query("/groups/" + group.id, "PUT", group, false)
        },
        // delete() - Deletes a group at DELETE /groups/:id
        delete: function(id) {
            return query("/groups/" + id, "DELETE", {}, false)
        }
    },
    // templates contains the endpoints for /templates
    templates: {
        // get() - Queries the API for GET /templates
        get: function() {
            return query("/templates/", "GET", {}, false)
        },
        // post() - Posts a template to POST /templates
        post: function(template) {
            return query("/templates/", "POST", template, false)
        }
    },
    // templateId contains the endpoints for /templates/:id
    templateId: {
        // get() - Queries the API for GET /templates/:id
        get: function(id) {
            return query("/templates/" + id, "GET", {}, false)
        },
        // put() - Puts a template to PUT /templates/:id
        put: function(template) {
            return query("/templates/" + template.id, "PUT", template, false)
        },
        // delete() - Deletes a template at DELETE /templates/:id
        delete: function(id) {
            return query("/templates/" + id, "DELETE", {}, false)
        }
    },
    // pages contains the endpoints for /pages
    pages: {
        // get() - Queries the API for GET /pages
        get: function() {
            return query("/pages/", "GET", {}, false)
        },
        // post() - Posts a page to POST /pages
        post: function(page) {
            return query("/pages/", "POST", page, false)
        }
    },
    // pageId contains the endpoints for /pages/:id
    pageId: {
        // get() - Queries the API for GET /pages/:id
        get: function(id) {
            return query("/pages/" + id, "GET", {}, false)
        },
        // put() - Puts a page to PUT /pages/:id
        put: function(page) {
            return query("/pages/" + page.id, "PUT", page, false)
        },
        // delete() - Deletes a page at DELETE /pages/:id
        delete: function(id) {
            return query("/pages/" + id, "DELETE", {}, false)
        }
    },
    // SMTP contains the endpoints for /smtp
    SMTP: {
        // get() - Queries the API for GET /smtp
        get: function() {
            return query("/smtp/", "GET", {}, false)
        },
        // post() - Posts a SMTP to POST /smtp
        post: function(smtp) {
            return query("/smtp/", "POST", smtp, false)
        }
    },
    // SMTPId contains the endpoints for /smtp/:id
    SMTPId: {
        // get() - Queries the API for GET /smtp/:id
        get: function(id) {
            return query("/smtp/" + id, "GET", {}, false)
        },
        // put() - Puts a SMTP to PUT /smtp/:id
        put: function(smtp) {
            return query("/smtp/" + smtp.id, "PUT", smtp, false)
        },
        // delete() - Deletes a SMTP at DELETE /smtp/:id
        delete: function(id) {
            return query("/smtp/" + id, "DELETE", {}, false)
        }
    },
    // import handles all of the "import" functions in the api
    import_email: function(raw) {
        return query("/import/email", "POST", {}, false)
    },
    // clone_site handles importing a site by url
    clone_site: function(req) {
        return query("/import/site", "POST", req, false)
    },
    // send_test_email sends an email to the specified email address
    send_test_email: function(req) {
        return query("/util/send_test_email", "POST", req, true)
    }
}

// Register our moment.js datatables listeners
$(document).ready(function() {
    $.fn.dataTable.moment('MMMM Do YYYY, h:mm:ss a');
    // Setup tooltips
    $('[data-toggle="tooltip"]').tooltip()
});
