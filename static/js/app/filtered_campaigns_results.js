var map = null;
var campaignTable = null;
var detailsTable = null;

// statuses is a helper map to point result statuses to ui classes
var statuses = {
    "Email Sent": {
        slice: "ct-slice-donut-sent",
        legend: "ct-legend-sent",
        label: "label-success",
        icon: "fa-envelope",
        point: "ct-point-sent"
    },
    "Error Sending": {
        slice: "ct-slice-donut-opened",
        legend: "ct-legend-opened",
        label: "label-warning",
        icon: "fa-envelope",
        point: "ct-point-opened"
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

// labels is a map of campaign statuses to
// CSS classes
var labels = {
    "In progress": "label-primary",
    "Queued": "label-info",
    "Completed": "label-success",
    "Emails Sent": "label-success",
    "Error": "label-danger"
}

// detailLabels is a map of email statuses to
// CSS classes
var detailLabels = {
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
    var error_sending = 0 ;
    var sent = 0;
    var opened = 0;
    var clicked = 0;
    var phishSuccess = 0;
    var submittedData = 0;
    var uniqueCredentials = 0;

    $.each(summaryStats, function(i, campaign) {
        sent += campaign.sent;
        error_sending += campaign.error_sending;
        opened += campaign.opened;
        clicked += campaign.clicked;
        submittedData += campaign.credentialsentered;
        uniqueCredentials += campaign.uniquecredentialsentered;
    });
    return {'sent': sent, 'error_sending': error_sending, 'opened': opened, 'clicked': clicked, 'submitted-data': submittedData, 'unique-credentials': uniqueCredentials};
}

function percent(frac) {
    return Math.floor(frac * 400)/4;
}
function updateSummaryStats(summaryStats) {
    var totals = getSummaryTotals(summaryStats);
    var filter = encodeURIComponent(getFilter());
    var ahref = "<a class='btn btn-primary' href='/filterresults?filter=";
    var sentStatuses = "&statuses=Email+Sent&statuses=Email+Opened&statuses=Clicked+Link&statuses=Submitted+Data";
    var openedStatuses = "&statuses=Email+Opened&statuses=Clicked+Link&statuses=Submitted+Data";
    var clickedStatuses = "&statuses=Clicked+Link&statuses=Submitted+Data";
    var submittedDataStatuses = "&statuses=Submitted+Data";
    var linkSent = ahref + filter + sentStatuses + "'>" + totals['sent'] + " (" +percent(parseInt(totals['sent'])/parseInt(totals['sent']))+ "%)"+'</a>';
    var linkOpened = ahref + filter + openedStatuses + "'>" + totals['opened'] + " (" +percent(parseInt(totals['opened'])/parseInt(totals['sent']))+ "%)" + '</a>';
    var linkClicked = ahref + filter + clickedStatuses + "'>" + totals['clicked'] + " (" +percent(parseInt(totals['clicked'])/parseInt(totals['sent']))+ "%)" + '</a>';
    var linkUniqueCredentials = ahref + filter + submittedDataStatuses + "'>" + totals['unique-credentials'] + " (" +percent(parseInt(totals['unique-credentials'])/parseInt(totals['sent']))+ "%)" + '</a>';
    $("#emails-sent").html(linkSent);
    $("#emails-opened").html(linkOpened);
    $("#emails-clicked").html(linkClicked);
    $("#emails-unique").html(linkUniqueCredentials);
    $("#email_sent_value").html(totals['sent']);
    $("#email_sent_bar").css("width",percent(parseInt(totals['sent'])/parseInt(totals['sent']))+ "%");
    $("div.email_sent_per").html(percent(parseInt(totals['sent'])/parseInt(totals['sent']))+ "%");
    $("#email_click_value").html(totals['clicked']);
    $("#email_click_bar").css("width",percent(parseInt(totals['clicked'])/parseInt(totals['sent']))+ "%");
    $("div.email_click_per").html(percent(parseInt(totals['clicked'])/parseInt(totals['sent']))+ "%");
    $("#email_open_value").html(totals['opened']);
    $("#email_open_bar").css("width",percent(parseInt(totals['opened'])/parseInt(totals['sent']))+ "%");
    $("div.email_open_per").html(percent(parseInt(totals['opened'])/parseInt(totals['sent']))+ "%");
    $("#email_credential_value").html(totals['unique-credentials']);
    $("#email_credential_bar").css("width", percent(parseInt(totals['unique-credentials'])/parseInt(totals['sent']))+ "%");
    $("div.email_credential_per").html(percent(parseInt(totals['unique-credentials'])/parseInt(totals['sent']))+ "%");
}

function getFilter() {
    var filter = getUrlParameter('filter');
    if (filter === undefined) {
        filter = $("#filter").val();
    }
    return filter;
}

function getMatchExact() {
    var matchExact = getUrlParameter('exactmatch');
    if (matchExact === undefined) {
        matchExact = 'off';
    }
    return matchExact;
}

function populateSearchForm() {
    var filter = getUrlParameter('filter');
    if (filter === undefined) {
    } else {
        if (filter !== "") {
            $("#filter").val(filter);
        }
    }
    var matchExact = getUrlParameter('exactmatch');
    if (matchExact === undefined) {
        document.getElementById("exactmatch").checked = false;
    } else {
        document.getElementById("exactmatch").checked = true;
    }
}

function load(filter, matchExact) {
    $("#loading").show();
    api.campaignssummarystats.get(filter, matchExact)
        .success(function(summaryStats) {
            // createEmailsSentPieChart(summaryStats);
            // createEmailsOpenedPieChart(summaryStats);
            // createEmailsClickedPieChart(summaryStats);
            // createUniqueCredsPieChart(summaryStats);
            updateSummaryStats(summaryStats);
            barChart(summaryStats );
            pieChart(summaryStats);
            $("#loading").hide()
            $("#campaign-results-body").show();
            populateNamesTable(summaryStats);
        }).error(function() {
            $("#loading").hide()
            errorFlash("Error fetching campaigns")
        });

    var statuses = ['Email Sent', 'Email Opened', 'Clicked Link', 'Submitted Data'];
    if (getUrlParameter('details') === 'true') {
        api.phishingresults.get(filter, matchExact, statuses)
            .success(function(details) {
                $("#details-body").show();
                populateDetailsTable(details);
            }).error(function() {
                errorFlash("Error fetching result details")
            });
    }
}
//custom charts function by star using amchart //
function pieChart(summaryStats){
    var totals = getSummaryTotals(summaryStats);
    console.log(totals['error_sending']);
    var chart = AmCharts.makeChart( "sentpiechart", {
        "type": "pie",
        "theme": "none",
        "titles": [ {
            "text": "Emails Sent",
            "size": 16
        } ],
        "dataProvider": [ {
        "status": "Email Sent",
        "color":"#0D8ECF",
        "value": totals['sent']
        }, 
        {
        "status": "Error sending",
        "color":"#FF9E01",
        "value": totals['error_sending']
        }

      ],
      "valueField": "value",
      "colorField": "color",
      "titleField": "status",
      "labelText": "[[title]]: [[value]]",
      "outlineAlpha": 0.4,
      "innerRadius": "50%",
      "labelRadius": 25,
      "depth3D": 15,
      "balloonText": "[[title]]<br><span style='font-size:14px'><b>[[value]]</b> </span>",
      "angle": 30,
      "export": {
        "enabled": true
      }
    } );
    var chart = AmCharts.makeChart( "openpiechart", {
        "type": "pie",
        "theme": "none",
        "titles": [ {
            "text": "Emails Opened",
            "size": 16
        } ],
        "dataProvider": [ {
        "status": "Opened",
        "color":"#FF9E01",
        "value": totals['opened']
        }, 
        {
        "status": "Unopened",
        "color":"#0D8ECF",
        "value": totals['sent'] - totals['opened']
        }

      ],
      "valueField": "value",
      "titleField": "status",
      "outlineAlpha": 0.4,
      "colorField": "color",
      "labelText": "[[title]]: [[value]]",
      "labelRadius": 25,
      "innerRadius": "50%",
      "depth3D": 15,
      "balloonText": "[[title]]<br><span style='font-size:14px'><b>[[value]]</b> </span>",
      "angle": 30,
      "export": {
        "enabled": true
      }
    } );
    var chart = AmCharts.makeChart( "linkpiechart", {
        "type": "pie",
        "theme": "none",
        "titles": [ {
            "text": "Links Clicked",
            "size": 16
        } ],
        "dataProvider": [ {
        "status": "Clicked",
        "color":"#FF9E01",
        "value": totals['clicked']
        }, 
        {
        "status": "Didn't clicked",
        "color":"#0D8ECF",
        "value": totals['sent'] - totals['clicked']
        }

      ],
      "valueField": "value",
      "titleField": "status",
      "outlineAlpha": 0.4,
      "labelText": "[[title]]: [[value]]",
      "innerRadius": "50%",
      "colorField": "color",
      "labelRadius": 25,
      "depth3D": 15,
      "balloonText": "[[title]]<br><span style='font-size:14px'><b>[[value]]</b> </span>",
      "angle": 30,
      "export": {
        "enabled": true
      }
    } );
    var chart = AmCharts.makeChart( "credentialpiechart", {
        "type": "pie",
        "theme": "none",
        "titles": [ {
            "text": "Credentials Entered",
            "size": 16
        } ],
        "dataProvider": [ {
        "status": "Credentials Entered",
        "color":"#0D8ECF",
        "value": totals['unique-credentials']
        }, 
        {
        "status": "No Credentials",
        "color":"#FF9E01",
        "value": totals['sent'] - totals['unique-credentials']
        }

      ],
      "valueField": "value",
      "titleField": "status",
      "outlineAlpha": 0.4,
      "colorField": "color",
      "labelRadius": 25,
      "labelText": "[[title]]: [[value]]",
      "innerRadius": "50%",
      "depth3D": 15,
      "balloonText": "[[title]]<br><span style='font-size:14px'><b>[[value]]</b> </span>",
      "angle": 30,
      "export": {
        "enabled": true
      }
    } );

}
function barChart(summaryStats){
    var totals = getSummaryTotals(summaryStats);
    
    var chart = AmCharts.makeChart("barchartdiv", {
        "theme": "none",
        "type": "serial",
        
        "dataProvider": [{
            "status": "Emails Sent",
            "done": totals['sent'],
            "notyet": totals['error_sending'],
            "errorballoontitle":"Error Sending: "
        }, {
            "status": "Emails Opened",
            "done": totals['opened'],
            "notyet": totals['sent'] - totals['opened'],
            "errorballoontitle":"Unopened: "
        }, {
            "status": "Links Clicked",
            "done": totals['clicked'],
            "notyet": totals['sent'] - totals['clicked'],
            "errorballoontitle":"Didn't Click: "
        }, {
            "status": "Credentials Entered",
            "done": totals['unique-credentials'],
            "notyet":totals['sent'] -  totals['unique-credentials'],
            "errorballoontitle":"No Credentials: "
        }],
        "startDuration": 1,
        "graphs": [ {
            "balloonText": "[[title]] [[category]]: <b>[[value]]</b>",
            "balloonFunction" : adjustBalloonText,
            "fillAlphas": 0.7,
            "lineAlpha": 0.2,
            "title": "Error",
            "type": "column",
            "valueField": "notyet"
        },{
            "balloonText": "[[category]]: <b>[[value]]</b>",
            "fillAlphas": 0.7,
            "fillColors":"#0D8ECF",
            "lineAlpha": 0.2,
            "title": "Success",
            "type": "column",
            "valueField": "done"
        }],
        "plotAreaFillAlphas": 0.1,
        "depth3D": 20,
        "angle": 30,
        "rotate": true,
        "categoryField": "status",
        "categoryAxis": {
            "gridPosition": "start"
        },
        "export": {
            "enabled": true
         }
    });
    function adjustBalloonText(graphDataItem, chart){
        var value = graphDataItem.values.value;
        var errormsg = graphDataItem.serialDataItem.dataContext.errorballoontitle
        return errormsg + value;
    }
}
function populateDetailsTable(results) {
    detailsTable.clear().draw();
    if (results.length > 2000) {
        results = results.slice(0, 2000);
        $("#details-overflow").html("<b>First 2000 results shown</b>");
    }
    $.each(results, function(i, r) {
        if (r.status === "Submitted Data") {
            r.status = "Submitted Credentials";
        }
        label = detailLabels[r.status] || "label-default";
        campaignsahref = '/campaigns/';
        var status = "<span class=\"label " + label + "\">" + r.status + "</span>";
        var emailahref = '/target?email=' + encodeURI(r.email) + "&firstname=" + encodeURI(r.first_name) + "&lastname=" + encodeURI(r.last_name);
        var emaillink = "<a href='" + emailahref + "'>" + escapeHtml(r.email) + "</a>";
        detailsTable.row.add([
            escapeHtml(r.first_name) + ' ' + escapeHtml(r.last_name),
            emaillink,
            status,
        ]);
        $('[data-toggle="tooltip"]').tooltip();
    });
    detailsTable.draw();
}

function submitSearchForm() {
    document.getElementById('search').submit();
}

function loadAll() {
    var searchForm = document.getElementById('search');
    searchForm.reset();
    searchForm.submit();
}

function populateNamesTable(summaryStats) {
    if (summaryStats.length == 0) {
        $("#campaign-results-body").hide();
        errorFlash(" Campaign not found!");
    } else {
        campaignTable.clear().draw();
        $.each(summaryStats, function(i, campaign) {
            label = labels[campaign.status] || "label-default";
            var campaignahref = '/campaigns/' + campaign.id;
            var filterhref ="/filteredcampaigns?filter=" + encodeURIComponent(campaign.name) +"&exactmatch=on&details=true";
            var campaignlink = "<a data-toggle='tooltip' data-placement='right' title='View Results' href='" + filterhref + "'>" + escapeHtml(campaign.name) + "</a>";
            var created = moment(campaign.created_date).format('MMMM Do YYYY, h:mm:ss a');
            var results = "<div class='text-center'><a class='btn btn-outline blue btn-sm ' data-toggle='tooltip' data-placement='left' title='View Campaign' href='" + campaignahref + "'>" + "<i class='fa fa-bar-chart'></i></a></div>";
            var sent = "<span class='pull-right'>" + campaign.sent + "</span>";
            campaignTable.row.add([
                campaignlink,
                created,
                sent,
                results,
            ]);
        });
        $('[data-toggle="tooltip"]').tooltip();
        campaignTable.draw();
    }
}

function createEmailsOpenedPieChart(summaryStats) {
    var email_opts = {
        donut: true,
        donutWidth: 50,
        chartPadding: 0,
        showLabel: true
    };
    var totals = getSummaryTotals(summaryStats);
    var email_data = {
        series: [{meta: 'Opened', slice: 'ct-slice-donut-opened', legend: 'ct-legend-opened', value:totals['opened']}, {meta: 'Unopened', slice: 'ct-slice-donut-sent', legend: "ct-legend-sent", value:totals['sent'] - totals['opened']}]
    };
    $("#email_chart_legend").html("")
    $.each(email_data['series'], function(idx) {
        var legend = email_data['series'][idx]['legend'];
        var meta = email_data['series'][idx]['meta'];
        $("#email_chart_legend").append('<li><span class="' + legend + '"></span><br/>' + meta + '</li>')
    })
    var email_chart = new Chartist.Pie("#email_chart", email_data, email_opts)
    email_chart.on('draw', function(data) {
        if (data.meta !== undefined) {
            data.element.addClass(statuses[data.meta].slice);
        }
    })
    // Setup the average chart listeners
    $piechart = $("#email_chart")
    var $pietoolTip = $piechart
        .append('<div class="chartist-tooltip"></div>')
        .find('.chartist-tooltip')
        .hide();

    $piechart.on('mouseenter', '.ct-slice-donut', function() {
        var $point = $(this)
        value = $point.attr('ct:value')
        label = $point.attr('ct:meta')
        $pietoolTip.html(label + ': ' + value.toString()).show();
    });

    $piechart.on('mouseleave', '.ct-slice-donut', function() {
        $pietoolTip.hide();
    });
    $piechart.on('mousemove', function(event) {
        $pietoolTip.css({
            left: (event.offsetX || event.originalEvent.layerX) - $pietoolTip.width() / 2 - 10,
            top: (event.offsetY + 40 || event.originalEvent.layerY) - $pietoolTip.height() - 80
        });
    });
}

function createEmailsClickedPieChart(summaryStats) {
    var opts = {
        donut: true,
        donutWidth: 50,
        chartPadding: 0,
        showLabel: true
    };
    var totals = getSummaryTotals(summaryStats);
    var data = {
        series: [{meta: 'Clicked', slice: 'ct-slice-donut-clicked', legend: 'ct-legend-clicked', value:totals['clicked']}, {meta: "Didn't Click", slice: 'ct-slice-donut-sent', legend: "ct-legend-sent", value:totals['sent'] - totals['clicked']}]
    };
    $("#email_clicked_chart_legend").html("")
    $.each(data['series'], function(idx) {
        var legend = data['series'][idx]['legend'];
        var meta = data['series'][idx]['meta'];
        $("#email_clicked_chart_legend").append('<li><span class="' + legend + '"></span><br/>' + meta + '</li>')
    })
    var email_clicked_chart = new Chartist.Pie("#email_clicked_chart", data, opts)
    email_clicked_chart.on('draw', function(data) {
        if (data.meta !== undefined) {
            data.element.addClass(statuses[data.meta].slice);
        }
    })
    // Setup the average chart listeners
    $piechart = $("#email_clicked_chart")
    var $pietoolTip = $piechart
        .append('<div class="chartist-tooltip"></div>')
        .find('.chartist-tooltip')
        .hide();

    $piechart.on('mouseenter', '.ct-slice-donut', function() {
        var $point = $(this)
        value = $point.attr('ct:value')
        label = $point.attr('ct:meta')
        $pietoolTip.html(label + ': ' + value.toString()).show();
    });

    $piechart.on('mouseleave', '.ct-slice-donut', function() {
        $pietoolTip.hide();
    });
    $piechart.on('mousemove', function(event) {
        $pietoolTip.css({
            left: (event.offsetX || event.originalEvent.layerX) - $pietoolTip.width() / 2 - 10,
            top: (event.offsetY + 40 || event.originalEvent.layerY) - $pietoolTip.height() - 80
        });
    });
}

function createEmailsSentPieChart(summaryStats) {
    var opts = {
        donut: true,
        donutWidth: 50,
        chartPadding: 0,
        showLabel: true
    };
    var totals = getSummaryTotals(summaryStats);
    var data = {
        series: [{meta: 'Email Sent', slice: 'ct-slice-donut-sent', legend: 'ct-legend-sent', value:totals['sent']}, {meta: "Error Sending", slice: 'ct-slice-donut-opened', legend: "ct-legend-opened", value:totals['error_sending']}]
    };
    $("#email_sent_chart_legend").html("")
    $.each(data['series'], function(idx) {
        var legend = data['series'][idx]['legend'];
        var meta = data['series'][idx]['meta'];
        $("#email_sent_chart_legend").append('<li><span class="' + legend + '"></span><br/>' + meta + '</li>')
    })
    var email_sent_chart = new Chartist.Pie("#email_sent_chart", data, opts)
    email_sent_chart.on('draw', function(data) {
        if (data.meta !== undefined) {
            data.element.addClass(statuses[data.meta].slice);
        }
    })
    // Setup the average chart listeners
    $piechart = $("#email_sent_chart")
    var $pietoolTip = $piechart
        .append('<div class="chartist-tooltip"></div>')
        .find('.chartist-tooltip')
        .hide();

    $piechart.on('mouseenter', '.ct-slice-donut', function() {
        var $point = $(this)
        value = $point.attr('ct:value')
        label = $point.attr('ct:meta')
        $pietoolTip.html(label + ': ' + value.toString()).show();
    });

    $piechart.on('mouseleave', '.ct-slice-donut', function() {
        $pietoolTip.hide();
    });
    $piechart.on('mousemove', function(event) {
        $pietoolTip.css({
            left: (event.offsetX || event.originalEvent.layerX) - $pietoolTip.width() / 2 - 10,
            top: (event.offsetY + 40 || event.originalEvent.layerY) - $pietoolTip.height() - 80
        });
    });
}

function createUniqueCredsPieChart(summaryStats) {
    var opts = {
        donut: true,
        donutWidth: 50,
        chartPadding: 0,
        showLabel: true
    };
    var totals = getSummaryTotals(summaryStats);
    var data = {
        series: [{meta: 'Credentials Entered', slice: 'ct-slice-donut-clicked', legend: 'ct-legend-clicked', value:totals['unique-credentials']}, {meta: "No Credentials", slice: 'ct-slice-donut-sent', legend: "ct-legend-sent", value:totals['sent'] - totals['unique-credentials']}]
    };
    $("#email_creds_chart_legend").html("")
    $.each(data['series'], function(idx) {
        var legend = data['series'][idx]['legend'];
        var meta = data['series'][idx]['meta'];
        $("#email_creds_chart_legend").append('<li><span class="' + legend + '"></span><br/>' + meta + '</li>')
    })
    var email_creds_chart = new Chartist.Pie("#email_creds_chart", data, opts)
    email_creds_chart.on('draw', function(data) {
        if (data.meta !== undefined) {
            data.element.addClass(statuses[data.meta].slice);
        }
    })
    // Setup the average chart listeners
    $piechart = $("#email_creds_chart")
    var $pietoolTip = $piechart
        .append('<div class="chartist-tooltip"></div>')
        .find('.chartist-tooltip')
        .hide();

    $piechart.on('mouseenter', '.ct-slice-donut', function() {
        var $point = $(this)
        value = $point.attr('ct:value')
        label = $point.attr('ct:meta')
        $pietoolTip.html(label + ': ' + value.toString()).show();
    });

    $piechart.on('mouseleave', '.ct-slice-donut', function() {
        $pietoolTip.hide();
    });
    $piechart.on('mousemove', function(event) {
        $pietoolTip.css({
            left: (event.offsetX || event.originalEvent.layerX) - $pietoolTip.width() / 2 - 10,
            top: (event.offsetY + 40 || event.originalEvent.layerY) - $pietoolTip.height() - 80
        });
    });
}

$(document).ready(function() {
    $("#loading").hide()
    populateSearchForm();
    campaignTable = $("#campaignTable").DataTable({
        columnDefs: [{
            orderable: false,
            targets: "no-sort"
        }]
    });
    detailsTable = $("#detailsTable").DataTable({
        columnDefs: [{
            orderable: false,
            targets: "no-sort"
        }]
    });
    var filter = getUrlParameter('filter');
    load(getFilter(), getMatchExact());
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
};