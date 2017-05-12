var campaigns = []
    // labels is a map of campaign statuses to
    // CSS classes
var labels = {
    "In progress": "label-primary",
    "Queued": "label-info",
    "Completed": "label-success",
    "Emails Sent": "label-success",
    "Error": "label-danger"
}

function deleteCampaign(idx) {
    if (confirm("Delete " + campaigns[idx].name + "?")) {
        api.campaignId.delete(campaigns[idx].id)
            .success(function(data) {
                successFlash(data.message)
                location.reload()
            })
    }
}

function processCampaigns(campaignTable, campaigns) {
           $.each(campaigns, function(i, campaign) {
                var campaign_date = moment(campaign.created_date).format('MMMM Do YYYY, h:mm:ss a')
                var label = labels[campaign.status] || "label-default";
                // Add it to the table
                campaignsahref = '/campaigns/';
                copycampaign = "<span data-toggle='modal' data-target='#modal'><button class='btn btn-primary' data-toggle='tooltip' data-placement='left' title='Copy Campaign' onclick='copy(" + i + ")'>\
                    <i class='fa fa-copy'></i>\
                    </button></span>";
                deletecampaign = "<button class='btn btn-danger' onclick='deleteCampaign(" + i + ")' data-toggle='tooltip' data-placement='right' title='Delete Campaign'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>";
                if (typeof restricteduser !== 'undefined') {
                    campaignsahref = '/campaigns/';
                    deletecampaign = "";
                }
                campaignTable.row.add([
                    escapeHtml(campaign.name),
                    campaign_date,
                    "<span class=\"label " + label + "\">" + campaign.status + "</span>",
                    "<div class='pull-right'><a class='btn btn-primary' href='" + campaignsahref + campaign.id + "' data-toggle='tooltip' data-placement='right' title='View Results'>\
                    <i class='fa fa-bar-chart'></i>\
                    </a>" + deletecampaign
                ]).draw()
            });
}

$(document).ready(function() {
    var initialMax = 20;
    api.campaignnames.get(initialMax)
        .success(function(cs) {
            campaigns = cs;
            $("#loading").hide();
            if (campaigns.length > 0) {
                $("#dashboard").show();
            } else {
                $("#emptyMessage").show();
            }
            campaignTable = $("#campaignTable").DataTable({
                columnDefs: [{
                    orderable: false,
                    targets: "no-sort"
                }]
            });
            processCampaigns(campaignTable, campaigns);
            api.campaignnames.get(0)
                .success(function(allCS) {
                    processCampaigns(campaignTable, allCS.slice(initialMax, allCS.length));
                });
        });

    api.campaigns.get()
        .success(function(cs) {
            $("#loading").hide()
            campaigns = cs
            if (campaigns.length > 0) {
                $("#dashboard").show()
                    // Create the overview chart data
                var overview_data = {
                    labels: [],
                    series: [
                        []
                    ]
                }
                var average_data = {
                    series: []
                }
                var overview_opts = {
                    axisX: {
                        showGrid: false
                    },
                    showArea: true,
                    plugins: []
                }
                var average_opts = {
                    donut: true,
                    donutWidth: 40,
                    chartPadding: 0,
                    showLabel: false
                }
                var average = 0
                // campaignTable = $("#campaignTable").DataTable({
                //     columnDefs: [{
                //         orderable: false,
                //         targets: "no-sort"
                //     }]
                // });
                $.each(campaigns, function(i, campaign) {
                    var campaign_date = moment(campaign.created_date).format('MMMM Do YYYY, h:mm:ss a')
                    var label = labels[campaign.status] || "label-default";
                    // Add it to the table

                    campaignsahref = '/campaigns/';
                    copycampaign = "<span data-toggle='modal' data-target='#modal'><button class='btn btn-primary' data-toggle='tooltip' data-placement='left' title='Copy Campaign' onclick='copy(" + i + ")'>\
                    <i class='fa fa-copy'></i>\
                    </button></span>";
                    deletecampaign = "<button class='btn btn-danger' onclick='deleteCampaign(" + i + ")' data-toggle='tooltip' data-placement='right' title='Delete Campaign'>\
                    <i class='fa fa-trash-o'></i>\
                    </button></div>";
                    if (typeof restricteduser !== 'undefined') {
                        campaignsahref = '/campaigns/';
                        deletecampaign = "";
                        var abortcampaign = true;
                        for (var i = 0; i < restricteduser.campaignids.length; i++) {
                            if (restricteduser.campaignids[i].campaignid === campaign.id) {
                                abortcampaign = false;
                            }
                        }
                        if (abortcampaign) {
                            return;
                        }
                    }
                    
                    // campaignTable.row.add([
                    //         escapeHtml(campaign.name),
                    //         campaign_date,
                    //         "<span class=\"label " + label + "\">" + campaign.status + "</span>",
                    //         "<div class='pull-right'><a class='btn btn-primary' href='" + campaignsahref + campaign.id + "' data-toggle='tooltip' data-placement='right' title='View Results'>\
                    // <i class='fa fa-bar-chart'></i>\
                    // </a>" + deletecampaign
                    //     ]).draw()
                        // Add it to the chart data
                    campaign.y = 0
                    $.each(campaign.results, function(j, result) {
                        if (result.status == "Success") {
                            campaign.y++;
                        }
                    })
                    campaign.y = Math.floor((campaign.y / campaign.results.length) * 100)
                    average += campaign.y
                        // Add the data to the overview chart
                    overview_data.labels.push(campaign_date)
                    overview_data.series[0].push({
                        meta: i,
                        value: campaign.y
                    })
                })
                average = Math.floor(average / campaigns.length);
                average_data.series.push({
                    meta: "Unsuccessful Phishes",
                    value: 100 - average
                })
                average_data.series.push({
                        meta: "Successful Phishes",
                        value: average
                    })
                    // Build the charts
                var average_chart = new Chartist.Pie("#average_chart", average_data, average_opts)
                var overview_chart = new Chartist.Line('#overview_chart', overview_data, overview_opts)
                    // Setup the average chart listeners
                $piechart = $("#average_chart")
                var $pietoolTip = $piechart
                    .append('<div class="chartist-tooltip"></div>')
                    .find('.chartist-tooltip')
                    .hide();

                $piechart.on('mouseenter', '.ct-slice-donut', function() {
                    var $point = $(this)
                    value = $point.attr('ct:value')
                    label = $point.attr('ct:meta')
                    $pietoolTip.html(label + ': ' + value.toString() + "%").show();
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

                // Setup the overview chart listeners
                $chart = $("#overview_chart")
                var $toolTip = $chart
                    .append('<div class="chartist-tooltip"></div>')
                    .find('.chartist-tooltip')
                    .hide();

                $chart.on('mouseenter', '.ct-point', function() {
                    var $point = $(this)
                    value = $point.attr('ct:value') || 0
                    cidx = $point.attr('ct:meta')
                    $toolTip.html(campaigns[cidx].name + '<br>' + "Successes: " + value.toString() + "%").show();
                });

                $chart.on('mouseleave', '.ct-point', function() {
                    $toolTip.hide();
                });
                $chart.on('mousemove', function(event) {
                    $toolTip.css({
                        left: (event.offsetX || event.originalEvent.layerX) - $toolTip.width() / 2 - 10,
                        top: (event.offsetY + 40 || event.originalEvent.layerY) - $toolTip.height() - 40
                    });
                });
                $("#overview_chart").on("click", ".ct-point", function(e) {
                    var $cidx = $(this).attr('ct:meta');
                    window.location.href = "/campaigns/" + campaigns[cidx].id
                });
            } else {
                $("#emptyMessage").show()
            }
        })
        .error(function() {
            errorFlash("Error fetching campaigns")
        })
})
