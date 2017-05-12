'use strict';

var eventsTable = null;
// labels is a map of email statuses to
// CSS classes
var labels = {
    "Email Sent": "label-success",
    "Email Opened": "label-warning",
    "Clicked Link": "label-warning",
    "Submitted Credentials": "label-danger",
    "Error": "label-danger"
}

$(document).ready(function() {
    $("#loading").hide()
    eventsTable = $("#eventsTable").DataTable({
        columnDefs: [{
            orderable: false,
            targets: "no-sort"
        }]
    });

    var email = getUrlParameter('email');
    if (email === undefined) {
    } else {
        loadEvents(email);
    }
});

function loadEvents(email) {
    $("#loading").show()
    api.targetevents.get(email)
        .success(function(events) {
            api.campaignnames.get(0)
                .success(function(campaigns) {
                    $("#loading").hide()
                    populateEventsTable(campaigns, events);
                }).error(function() {
                    errorFlash("Error fetching events")
                });

        }).error(function() {
            errorFlash("Error fetching events")
        });
}

function populateEventsTable(campaigns, events) {
    eventsTable.clear().draw();
    if (events.length > 2000) {
        events = events.slice(0, 2000);
        $("#events-overflow").html("<b>First 2000 events shown</b>");
    }
    $.each(events, function(i, e) {
        var time = moment(e.time).format('MMMM Do YYYY, h:mm:ss a');
        var campaignname = "";
        var message = e.message;
        if (message === "Submitted Data") {
            message = "Submitted Credentials";
        }
        var label = labels[message] || "label-default";
        var status = "<span class=\"label " + label + "\">" + message + "</span>";
        for (var i = 0; i < campaigns.length; i++) {
            if (campaigns[i].id === e.campaignid) {
                campaignname = campaigns[i].name;
                break;
            }
        }
        var filterhref ="/filteredcampaigns?filter=" + encodeURIComponent(campaignname) +"&exactmatch=on&details=true";
        var campaignlink = "<a data-toggle='tooltip' data-placement='right' title='View Results' href='" + filterhref + "'>" + escapeHtml(campaignname) + "</a>";
        eventsTable.row.add([
            campaignlink,
            time,
            status,
            ]);
        $('[data-toggle="tooltip"]').tooltip();
    });
    eventsTable.draw();
    $('#events-body').show();
}

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
