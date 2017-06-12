var map = null;
var campaignTable = null;
var detailsTable = null;
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
var Campaign_array = [];
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
function init_campaigns(summaryStats) {
    var error_sending = 0 ;
    var sent = 0;
    var opened = 0;
    var clicked = 0;
    var phishSuccess = 0;
    var submittedData = 0;
    var uniqueCredentials = 0;
    var i = 0;
    $.each(summaryStats, function(i, campaign) {
        sent += campaign.sent;
        error_sending += campaign.error_sending;
        opened += campaign.opened;
        clicked += campaign.clicked;
        submittedData += campaign.credentialsentered;
        uniqueCredentials += campaign.uniquecredentialsentered;
        var new_item = $('#total_detail').clone();
        var chart_id = "filtered_barchart"+i;
        var campaign_status = "In progress";
        if(campaign.status == "Completed"){
            campaign_status ="Completed";
        }
        new_item.attr('id',"filtered"+i);
        new_item.find("#dashboard_amchart_1").attr("id",chart_id);
        new_item.find("#portlet_tab_1").attr("id","portlet_tab_1_"+i);
        new_item.find("a[href='#portlet_tab_1']").attr("href","#portlet_tab_1_"+i);
        new_item.find("#portlet_tab_2").attr("id","portlet_tab_2_"+i);
        new_item.find("a[href='#portlet_tab_2']").attr("href","#portlet_tab_2_"+i);
        new_item.find("div#sentpiechart").attr("id","sentpiechart_flitered_"+i);
        new_item.find("div#openpiechart").attr("id","openpiechart_flitered_"+i);
        new_item.find("div#clickpiechart").attr("id","clickpiechart_flitered_"+i);
        new_item.find("div#credentialpiechart").attr("id","credentialpiechart_flitered_"+i);
        new_item.find("a.accordion-toggle").attr("data-id",i);
        $('#tab_1').append(new_item);
        init_campaigns_percent("filtered"+i,campaign.name, campaign.sent, campaign.error_sending, campaign.opened, campaign.clicked, campaign.uniquecredentialsentered,campaign_status);
        i++;
    });
    init_campaigns_percent("total_detail","Totals", sent, error_sending, opened, clicked, uniqueCredentials,"Totals");

}
function make_Item_charts(id) {
    var chart_id = "filtered_barchart"+id;"filtered_barchart"+id;
    var campaign = Campaign_array[parseInt(id)];
    console.log(campaign);
    var chartData = [
        chart_id,
        campaign.sent,
        campaign.error_sending,
        campaign.opened,
        campaign.clicked,
        campaign.uniquecredentialsentered
    ];
    var sentData = [
      "sent",
      "sentpiechart_flitered_"+id,
      campaign.sent,
      campaign.error_sending
    ];
    var openedData = [
      "opened",
      "openpiechart_flitered_"+id,
      campaign.sent,
      campaign.opened
    ];
    var clickedData = [
      "clicked",
      "clickpiechart_flitered_"+id,
      campaign.sent,
      campaign.clicked
    ];
    var credetialData = [
      "unique-credentials",
      "credentialpiechart_flitered_"+id,
      campaign.sent,
      campaign.uniquecredentialsentered
    ]
    Chart.initItemBarchart(chartData);
    Chart.initItemPiechart(sentData);
    Chart.initItemPiechart(openedData);
    Chart.initItemPiechart(clickedData);
    Chart.initItemPiechart(credetialData);
}


function percent(frac) {
    return Math.floor(frac * 400)/4;
}
function init_campaigns_percent(id,name, sent, error_sending, opened, clicked, uniquecredentialsentered, status){
    var labelClass = "success";
    if (status == "Completed"){
        labelClass ="danger"
    }
    var errorsend_percent= percent(parseInt(error_sending)/parseInt(sent))+"%";
    var opened_percent= percent(parseInt(opened)/parseInt(sent))+"%";
    var clicked_percent= percent(parseInt(clicked)/parseInt(sent))+"%";
    var credential_percent= percent(parseInt(uniquecredentialsentered)/parseInt(sent))+"%";
    $("#"+id+" span.campaign_name").html(name);
    $("#"+id+" .ribbon-shadow").text(status);
    $("#"+id+" #emails-opened").html(opened);
    $("#"+id+" #emails-clicked").html(clicked);
    $("#"+id+" #emails-unique").html(uniquecredentialsentered);
    $("#"+id+" #email_sent_value").html(sent);
    $("#"+id+" #email_sent_bar").css("width",percent((parseInt(sent)-parseInt(error_sending))/parseInt(sent))+ "%");
    $("#"+id+" div.email_sent_per").html(percent((parseInt(sent)-parseInt(error_sending))/parseInt(sent))+ "%");
    $("#"+id+" #email_click_value").html(clicked);
    $("#"+id+" #email_click_bar").css("width",clicked_percent);
    $("#"+id+" div.email_click_per").html(clicked_percent);
    $("#"+id+" #email_open_value").html(opened);
    $("#"+id+" #email_open_bar").css("width",opened_percent);
    $("#"+id+" div.email_open_per").html(opened_percent);
    $("#"+id+" #email_credential_value").html(uniquecredentialsentered);
    $("#"+id+" #email_credential_bar").css("width", credential_percent);
    $("#"+id+" div.email_credential_per").html(credential_percent);
}
function updateSummaryStats(summaryStats) {
    var totals = getSummaryTotals(summaryStats);
    init_campaigns(summaryStats);
    var filter = encodeURIComponent(getFilter());
    var ahref = "<a class='btn btn-primary' href='/filterresults?filter=";
    var sentStatuses = "&statuses=Email+Sent&statuses=Email+Opened&statuses=Clicked+Link&statuses=Submitted+Data";
    var openedStatuses = "&statuses=Email+Opened&statuses=Clicked+Link&statuses=Submitted+Data";
    var clickedStatuses = "&statuses=Clicked+Link&statuses=Submitted+Data";
    var submittedDataStatuses = "&statuses=Submitted+Data";
    var linkSent = ahref + filter + sentStatuses + "'>" + totals['sent'] + " (" +percent(totals['sent']/parseInt(totals['sent']))+ "%)"+'</a>';
    var linkOpened = ahref + filter + openedStatuses + "'>" + totals['opened'] + " (" +percent(parseInt(totals['opened'])/parseInt(totals['sent']))+ "%)" + '</a>';
    var linkClicked = ahref + filter + clickedStatuses + "'>" + totals['clicked'] + " (" +percent(parseInt(totals['clicked'])/parseInt(totals['sent']))+ "%)" + '</a>';
    var linkUniqueCredentials = ahref + filter + submittedDataStatuses + "'>" + totals['unique-credentials'] + " (" +percent(parseInt(totals['unique-credentials'])/parseInt(totals['sent']))+ "%)" + '</a>';
    var chartData = [
        "dashboard_amchart_1",
        totals['sent'],
        totals['error_sending'],
        totals['opened'],
        totals['clicked'],
        totals['unique-credentials'] 
    ];
    var sentData = [
      "sent",
      "sentpiechart",
      totals['sent'],
      totals['error_sending']
    ];
    var openedData = [
      "opened",
      "openpiechart",
      totals['sent'],
      totals['opened']
    ];
    var clickedData = [
      "clicked",
      "clickpiechart",
      totals['sent'],
      totals["clicked"]
    ];
    var credetialData = [
      "unique-credentials",
      "credentialpiechart",
      totals['sent'],
      totals['unique-credentials']
    ]
    Chart.initItemBarchart(chartData);
    Chart.initItemPiechart(sentData);
    Chart.initItemPiechart(openedData);
    Chart.initItemPiechart(clickedData);
    Chart.initItemPiechart(credetialData);
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
function appendDetailData(summaryStats) {
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
function load(filter, matchExact) {
    $("#loading").show();
    api.campaignssummarystats.get(filter, matchExact)
        .success(function(summaryStats) {
            Campaign_array = summaryStats;
            updateSummaryStats(summaryStats);
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
    $(document).on('click', ".accordion-toggle", function () {
        $(this).parents('.panel-heading').siblings('.chartpanel').toggleClass('hidden');
        var id = $(this).attr("data-id");
        if(id != undefined){
            make_Item_charts(id);
        }
       
    });
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

//////// Bar and Pie chart Functions is defined//////
var Chart = {
    initItemBarchart:function(params){
        var chartData = [{
            "status": "Email<br>Sent",
            "numbers": params[1],
            "townSize": 7,
            "activityPercent": 100,
        }, {
            "status": "Error<br>Sending",
            "numbers": params[2],
            "NotActivityPercent": Number((params[2]/params[1])*100).toFixed(2)
        }, {
            "status": "Emails<br>Opened",
            "numbers": params[3],
            "townSize": 16,
            "activityPercent": Number((params[3]/params[1])*100).toFixed(2),
        }, {
            "status": "Unopened",
            "numbers": params[1] - params[3],
            "NotActivityPercent": Number(((params[1] - params[3])/params[1])*100).toFixed(2)
        }, {
            "status": "Links<br>Clicked",
            "numbers": params[4],
            "townSize": 11,
            "activityPercent": Number((params[4]/params[1])*100).toFixed(2),
        }, {
            "status": "Didn't<br>Click",
            "numbers":  params[1] - params[4],
            "NotActivityPercent": Number(((params[1] - params[4])/params[1])*100).toFixed(2)
        }, {
            "status": "Credentials<br>Entered",
            "numbers": params[5],
            "townSize": 18,
            "activityPercent": Number((params[5]/params[1])*100).toFixed(2),
            "townName": "Credentials Entered",
            "townName2": "Credentials Entered",
            "bulletClass": "lastBullet"
        }, {
            "status": "No<br>Credentials",
            "numbers": params[1] - params[5],
            "NotActivityPercent": Number(((params[1] - params[5])/params[1])*100).toFixed(2),
            "alpha": 0.4,
            
        }, {
            "status": " "
        }];
        var chart = AmCharts.makeChart(params[0], {
            type: "serial",
            fontSize: 12,
            fontFamily: "Open Sans",
            dataDateFormat: "YYYY-MM-DD",
            dataProvider: chartData,

            addClassNames: true,
            startDuration: 1,
            color: "#6c7b88",
            marginLeft: 0,

            categoryField: "status",
            categoryAxis: {
                minPeriod: "DD",
                autoGridCount: false,
                gridCount: 50,
                gridAlpha: 0.1,
                gridColor: "#FFFFFF",
                axisColor: "#555555",
            },
            valueAxes: [{
                id: "a1",
                title: "numbers",
                gridAlpha: 0,
                axisAlpha: 0
            }, {
                id: "a2",
                position: "right",
                gridAlpha: 0,
                axisAlpha: 0,
                labelsEnabled: false
            }, {
                id: "a3",
                title: "NotActivityPercent",
                position: "right",
                gridAlpha: 0,
                axisAlpha: 0,
                inside: true,
                NotActivityPercent: "%",
            }],
            graphs: [{
                id: "g1",
                valueField: "numbers",
                title: "number",
                type: "column",
                fillAlphas: 0.7,
                valueAxis: "a1",
                balloonText: "[[value]]",
                legendValueText: "[[value]]",
                legendPeriodValueText: "total: [[value.sum]]",
                lineColor: "#08a3cc",
                alphaField: "alpha",
            }, {
                id: "g2",
                valueField: "activityPercent",
                classNameField: "bulletClass",
                title: "activityPercent",
                type: "line",
                valueAxis: "a2",
                lineColor: "#786c56",
                lineThickness: 1,
                legendValueText: "[[description]]/[[value]]",
                bullet: "round",
                bulletSizeField: "townSize",
                bulletBorderColor: "#02617a",
                bulletBorderAlpha: 1,
                bulletBorderThickness: 2,
                bulletColor: "#89c4f4",
                labelText: "[[townName2]]",
                labelPosition: "right",
                balloonText: "activityPercent:[[value]]%",
                showBalloon: true,
                animationPlayed: true,
            }, {
                id: "g3",
                title: "NotActivityPercent",
                valueField: "NotActivityPercent",
                type: "line",
                valueAxis: "a3",
                lineAlpha: 0.8,
                lineColor: "#e26a6a",
                balloonText: "[[value]]%",
                lineThickness: 1,
                legendValueText: "[[value]]",
                bullet: "square",
                bulletBorderColor: "#e26a6a",
                bulletBorderThickness: 1,
                bulletBorderAlpha: 0.8,
                dashLengthField: "dashLength",
                animationPlayed: true
            }],

            chartCursor: {
                zoomable: false,
                cursorAlpha: 0,
                categoryBalloonColor: "#e26a6a",
                categoryBalloonAlpha: 0.8,
                valueBalloonsEnabled: false
            },
            legend: {
                bulletType: "round",
                equalWidths: false,
                valueWidth: 120,
                useGraphSettings: true,
                color: "#6c7b88"
            }
        });
        chart.addListener("rendered", addListeners);

        function addListeners() {
          var categoryAxis = chart.categoryAxis;
          categoryAxis.addListener("clickItem", handleClick);
          categoryAxis.addListener("rollOverItem", handleOver);
          categoryAxis.addListener("rollOutItem", handleOut);
        }

        function handleClick(event) {
          alert("click");
          console.log(event);
        }

        function handleOut(event) {
          event.target.setAttr("cursor", "default");
          event.target.setAttr("fill", "#000000");
          console.log("out");
          console.log(event);
        }

        function handleOver(event) {
          event.target.setAttr("cursor", "pointer");
          event.target.setAttr("fill", "#CC0000");
          console.log("over");
          console.log(event);
        }
    },
    initItemPiechart:function(params){ //status, id, total, success-params,
      switch(params[0]) {
        case "opened":
          var success_status = " Email Opened";
          var other_status ="Unopened";
          var unsuccess_count = params[2]-params[3];;
          var success_count = params[3];
          break;
        case "clicked":
          success_status = "Links Clicked";
          other_status = "Didn't Clicked";
          var unsuccess_count = params[2]-params[3];
          success_count = params[3];
          break;
        case "unique-credentials":
          success_status = "Credentials Entered";
          other_status ="NO credentials";
          var unsuccess_count = params[2]-params[3];
          success_count = params[3];
          break;
        default:
          success_status = "Email Sent";
          other_status ="Error Sending";
          unsuccess_count = params[3];
          success_count = params[2];
      }
      
      var chart = AmCharts.makeChart(params[1], {
        "type": "pie",
        "startDuration": 0,
        "theme": "light",
        "addClassNames": true,
        "legend":{
          "position":"right",
          "align":"right",
          "autoMargins":false
        },
        "innerRadius": "50%",
        "defs": {
          "filter": [{
            "id": "shadow",
            "width": "200%",
            "height": "200%",
            "feOffset": {
              "result": "offOut",
              "in": "SourceAlpha",
              "dx": 0,
              "dy": 0
            },
            "feGaussianBlur": {
              "result": "blurOut",
              "in": "offOut",
              "stdDeviation": 5
            },
            "feBlend": {
              "in": "SourceGraphic",
              "in2": "blurOut",
              "mode": "normal"
            }
          }]
        },
        "dataProvider": [{
          "status": success_status,
          "litres": success_count,
          "color": "#fdd400"
        }, {
          "status": other_status,
          "litres": unsuccess_count,//total - sucessstatus
          "color": "#67b7dc"
        }],
        "valueField": "litres",
        "titleField": "status",
        "colorField": "color",
        "labelsEnabled": false,
        "export": {
          "enabled": true
      }
      });  
    }
}
