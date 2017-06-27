var map = null;
var resultsTable = null;

// statuses is a helper map to point result statuses to ui classes
var statuses = {
    "Email Sent": {
        slice: "ct-slice-donut-sent",
        legend: "ct-legend-sent",
        label: "label-success",
        icon: "fa-envelope",
        point: "ct-point-sent"
    },
    "Unopened": {
        slice: "ct-slice-donut-sent",
        legend: "ct-legend-sent",
        label: "label-success",
        icon: "fa-envelope",
        point: "ct-point-sent"
    },
    "Email Opened": {
        slice: "ct-slice-donut-opened",
        legend: "ct-legend-opened",
        label: "label-warning",
        icon: "fa-envelope",
        point: "ct-point-opened"
    },
    "Opened": {
        slice: "ct-slice-donut-opened",
        legend: "ct-legend-opened",
        label: "label-warning",
        icon: "fa-envelope",
        point: "ct-point-opened"
    },
    "Didn't Submit": {
        slice: "ct-slice-donut-sent",
        legend: "ct-legend-sent",
        label: "label-success",
        icon: "fa-envelope",
        point: "ct-point-sent"
    },

    "Didn't Click": {
        slice: "ct-slice-donut-sent",
        legend: "ct-legend-sent",
        label: "label-success",
        icon: "fa-envelope",
        point: "ct-point-sent"
    },
    "Clicked": {
        slice: "ct-slice-donut-clicked",
        legend: "ct-legend-clicked",
        label: "label-danger",
        icon: "fa-mouse-pointer",
        point: "ct-point-clicked"
    },
    "Credentials Entered": {
        slice: "ct-slice-donut-clicked",
        legend: "ct-legend-clicked",
        label: "label-danger",
        icon: "fa-exclamation",
        point: "ct-point-clicked"
    },
    "No Credentials" : {
        slice: "ct-slice-donut-sent",
        legend: "ct-legend-sent",
        label: "label-success",
        icon: "fa-envelope",
        point: "ct-point-sent",
    },
    "Success": {
        slice: "ct-slice-donut-clicked",
        legend: "ct-legend-clicked",
        label: "label-danger",
        icon: "fa-exclamation",
        point: "ct-point-clicked"
    },
    "Error": {
        slice: "ct-slice-donut-error",
        legend: "ct-legend-error",
        label: "label-default",
        icon: "fa-times",
        point: "ct-point-error"
    },
    "Error Sending Email": {
        slice: "ct-slice-donut-error",
        legend: "ct-legend-error",
        label: "label-default",
        icon: "fa-times",
        point: "ct-point-error"
    },
    "Submitted Data": {
        slice: "ct-slice-donut-clicked",
        legend: "ct-legend-clicked",
        label: "label-danger",
        icon: "fa-exclamation",
        point: "ct-point-clicked"
    },
    "Unknown": {
        slice: "ct-slice-donut-error",
        legend: "ct-legend-error",
        label: "label-default",
        icon: "fa-question",
        point: "ct-point-error"
    },
    "Sending": {
        slice: "ct-slice-donut-sending",
        legend: "ct-legend-sending",
        label: "label-primary",
        icon: "fa-spinner",
        point: "ct-point-sending"
    },
    "Campaign Created": {
        label: "label-success",
        icon: "fa-rocket"
    }
}

// labels is a map of email statuses to
// CSS classes
var labels = {
    "Email Sent": "label-success",
    "Email Opened": "label-warning",
    "Clicked Link": "label-warning",
    "Submitted Credentials": "label-danger",
    "Error": "label-danger"
}

var campaign = {}
var bubbles = []

function dismiss() {
    $("#modal\\.flashes").empty()
    $("#modal").modal('hide')
}

function matchFilter(campaignName, filterTerms) {
    for (let filter of filterTerms) {
        if (!campaignName.includes(filter)) {
            return false;
        }
    }
    return true;
}

function exportCampaignsAsCSV() {
    var filter = getFilter();
    var matchExact = getMatchExact();
    window.location.href = '/exportcsv?' + 'filter=' + filter +'&matchexact=' + matchExact;
}

function getSummaryTotals(summaryStats) {
    var sent = 0;
    var opened = 0;
    var clicked = 0;
    var phishSuccess = 0;
    var submittedData = 0;
    var uniqueCredentials = 0;

    $.each(summaryStats, function(i, campaign) {
        sent += campaign.sent;
        opened += campaign.opened;
        clicked += campaign.clicked;
        submittedData += campaign.credentialsentered;
        uniqueCredentials += campaign.uniquecredentialsentered;
    });
    return {'sent': sent, 'opened': opened, 'clicked': clicked, 'submitted-data': submittedData, 'unique-credentials': uniqueCredentials};
}

function getFilter() {
    var filter = getUrlParameter('filter');
    if (filter === undefined) {
        filter = $("#filter").val();
    }
    return filter;
}

function getMatchExact() {
    var matchExact = 'off';
    return matchExact;
}

function getStatuses() {
    var statuses = getUrlParameterArray('statuses');
    if (statuses === undefined) {
        statuses = ['Email Sent'];
    }
    return statuses;
}

function populateSearchForm() {
    var filter = getUrlParameter('filter');
    if (filter === undefined) {
    } else {
        if (filter !== "") {
            $("#filter").val(filter);
        }
    }
    var status = getUrlParameterArray('statuses');
    for (var i = 0; i < status.length; i++) {
        switch(status[i]) {
        case "Email Sent":
            document.getElementById("email-sent").selected = true;
            break;
        case "Email Opened":
            document.getElementById("email-opened").selected = true;
            break;
        case "Clicked Link":
            document.getElementById("clicked-link").selected = true;
            break;
        case "Submitted Data":
            document.getElementById("submitted-data").selected = true;
            break;
        }
    }
}

function load(filter, matchExact, statuses) {
    $("#loading").show();
    api.phishingresults.get(filter, matchExact, statuses)
        .success(function(results) {
            $("#loading").hide()
            $("#results-body").show();
            populateResultsTable(results);
        }).error(function() {
            $("#loading").hide()
            errorFlash("Error fetching campaigns")
        });
}

function submitSearchForm() {
    document.getElementById('filterform').submit();
}

function populateResultsTable(results) {
    if (results.length == 0) {
        $("#results-body").hide();
        errorFlash(" Results not found!");
    } else {
        resultsTable.clear().draw();
        if (results.length > 2000) {
            results = results.slice(0, 2000);
            $("#results-overflow").html("<b>First 2000 results shown</b>");
        }
        $.each(results, function(i, r) {
            if (r.status === "Submitted Data") {
                r.status = "Submitted Credentials";
            }
            label = labels[r.status] || "label-default";
            campaignsahref = '/campaigns/';
            var link = "<div class=''><a href='" + campaignsahref + r.campaignid + "'>" + escapeHtml(r.campaignname) + "</a>";
            var status = "<span class=\"label " + label + "\">" + r.status + "</span>";
            var emailahref = '/target?email=' + encodeURI(r.email) + "&firstname=" + encodeURI(r.first_name) + "&lastname=" + encodeURI(r.last_name);
            var emaillink = "<a href='" + emailahref + "'>" + escapeHtml(r.email) + "</a>";
            resultsTable.row.add([
                link,
                emaillink,
                status,
            ]);
            $('[data-toggle="tooltip"]').tooltip();
        });
        resultsTable.draw();
    }
}

$(document).ready(function() {
    $("#loading").hide()
    populateSearchForm();
    resultsTable = $("#resultsTable").DataTable({
        columnDefs: [{
            orderable: false,
            targets: "no-sort"
        }]
    });
    var filter = getUrlParameter('filter');
    if (filter === undefined) {
    } else {
        load(getFilter(), getMatchExact(), getStatuses());
    }
});

var getUrlParameter = function getUrlParameter(sParam) {
    var urlQuery = window.location.search.substring(1);
    urlQuery = urlQuery.replace(/\+/g, '%20');
    var sPageURL = decodeURIComponent(urlQuery),
        sURLVariables = sPageURL.split('&'),
        sParameterName,
        i;

    for (i = 0; i < sURLVariables.length; i++) {
        sParameterName = sURLVariables[i].split('=');

        if (sParameterName[0] === sParam) {
            return sParameterName[1] === undefined ? true : sParameterName[1];
        }
    }
}
function getUrlParameterArray(sParam) {
    var urlQuery = window.location.search.substring(1);
    urlQuery = urlQuery.replace(/\+/g, '%20');
    var arr = [];
    var sPageURL = decodeURIComponent(urlQuery),
        sURLVariables = sPageURL.split('&'),
        sParameterName,
        i;

    for (i = 0; i < sURLVariables.length; i++) {
        sParameterName = sURLVariables[i].split('=');

        if (sParameterName[0] === sParam) {
            if (sParameterName[1] === undefined) {
                arr.push(true);
            } else {
                arr.push(sParameterName[1]);
            }
        }
    }
    return arr;
};
