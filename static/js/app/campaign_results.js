var map = null
var doPoll = true;

// statuses is a helper map to point result statuses to ui classes
var statuses = {
    "Email Sent": {
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
    "Clicked Link": {
        slice: "ct-slice-donut-clicked",
        legend: "ct-legend-clicked",
        label: "label-danger",
        icon: "fa-mouse-pointer",
        point: "ct-point-clicked"
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

var campaign = {}
var bubbles = []

function dismiss() {
    $("#modal\\.flashes").empty()
    $("#modal").modal('hide')
    $("#resultsTable").dataTable().DataTable().clear().draw()
}

// Deletes a campaign after prompting the user
function deleteCampaign() {
    swal({
        title: "Are you sure?",
        text: "This will delete the campaign. This can't be undone!",
        type: "warning",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Delete Campaign",
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        preConfirm: function() {
            return new Promise(function(resolve, reject) {
                api.campaignId.delete(campaign.id)
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
            location.href = '/campaigns'
        })
    })
}

// Completes a campaign after prompting the user
function completeCampaign() {
    swal({
        title: "Are you sure?",
        text: "Phishaway will stop processing events for this campaign",
        type: "warning",
        animation: false,
        showCancelButton: true,
        confirmButtonText: "Complete Campaign",
        confirmButtonColor: "#428bca",
        reverseButtons: true,
        allowOutsideClick: false,
        preConfirm: function() {
            return new Promise(function(resolve, reject) {
                api.campaignId.complete(campaign.id)
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
            'Campaign Completed!',
            'This campaign has been completed!',
            'success'
        );
        $('#complete_button')[0].disabled = true;
        $('#complete_button').text('Completed!')
        doPoll = false;
    })
}

// Exports campaign results as a CSV file
function exportAsCSV(scope) {
    exportHTML = $("#exportButton").html()
    var csvScope = null
    switch (scope) {
        case "results":
            csvScope = campaign.results
            break;
        case "events":
            csvScope = campaign.timeline
            break;
    }
    if (!csvScope) {
        return
    }
    $("#exportButton").html('<i class="fa fa-spinner fa-spin"></i>')
    var csvString = Papa.unparse(csvScope, {})
    var csvData = new Blob([csvString], {
        type: 'text/csv;charset=utf-8;'
    });
    if (navigator.msSaveBlob) {
        navigator.msSaveBlob(csvData, scope + '.csv');
    } else {
        var csvURL = window.URL.createObjectURL(csvData);
        var dlLink = document.createElement('a');
        dlLink.href = csvURL;
        dlLink.setAttribute('download', scope + '.csv');
        document.body.appendChild(dlLink)
        dlLink.click();
        document.body.removeChild(dlLink)
    }
    $("#exportButton").html(exportHTML)
}

// Exports campaign results as a PDF file
function exportAsPDF(scope) {
    exportHTML = $("#exportButton").html()
    var pdfScope = null
    switch (scope) {
        case "results":
            pdfScope = campaign.results
            break;
        case "events":
            pdfScope = campaign.timeline
            break;
    }
    if (!pdfScope) {
        return
    }
    window.location.href = '/pdfreport?' + 'id=' + campaign.id +'&scope=' + scope;
}


function replay(event_idx) {
    request = campaign.timeline[event_idx]
    details = JSON.parse(request.details)
    url = null
    form = $('<form>').attr({
            method: 'POST',
            target: '_blank',
        })
        /* Create a form object and submit it */
    $.each(Object.keys(details.payload), function(i, param) {
            if (param == "rid") {
                return true;
            }
            if (param == "__original_url") {
                url = details.payload[param];
                return true;
            }
            $('<input>').attr({
                name: param,
            }).val(details.payload[param]).appendTo(form);
        })
        /* Ensure we know where to send the user */
        // Prompt for the URL
    swal({
        title: 'Where do you want the credentials submitted to?',
        input: 'text',
        showCancelButton: true,
        inputPlaceholder: "http://example.com/login",
        inputValue: url || "",
        inputValidator: function(value) {
            return new Promise(function(resolve, reject) {
                if (value) {
                    resolve();
                } else {
                    reject('Invalid URL.');
                }
            });
        }
    }).then(function(result) {
        url = result
        submitForm()
    })
    return
    submitForm()

    function submitForm() {
        form.attr({
            action: url
        })
        form.appendTo('body').submit().remove()
    }
}

function renderTimeline(data) {
    var emaillink = data[4];
    emaillink = emaillink.trimRight();
    // remove trailing </a>
    var email = emaillink.slice(0, emaillink.length - "</a>".length);
    var idx = email.lastIndexOf(">");
    if (idx == - 1) {
        // ERROR!
    } else {
        email = email.slice(idx + 1, email.length);
    }
    record = {
        "first_name": data[2],
        "last_name": data[3],
        "email": email,
        "position": data[5],
        "department": data[6]
    }
    var anchorName = encodeURI(record.email);
    results = '<a name="' + anchorName + '"' + 'id = "' + anchorName + '"' + '></a>';
    results += '<div class="timeline col-sm-12 well well-lg">';
    results += '<p style="margin-left:60px">Timeline for ' + escapeHtml(record.first_name) + ' ' + escapeHtml(record.last_name) +
        '</p><span class="subtitle" style="margin-left:60px">Email: ' + escapeHtml(record.email) + '</span>' +
        '<div class="timeline-graph col-sm-6">'
    $.each(campaign.timeline, function(i, event) {
        if (!event.email || event.email == record.email) {
            // Add the event
            results += '<div class="timeline-entry">'
            results +=
                '    <div class="timeline-message">' + escapeHtml(event.message) +
                '    <span class="timeline-date">' + moment(event.time).format('MMMM Do YYYY h:mm a') + '</span>'
            if (event.details) {
                if (event.message == "Submitted Data") {
                    results += '<div class="timeline-replay-button"><button onclick="replay(' + i + ')" class="btn btn-success">'
                    results += '<i class="fa fa-refresh"></i> Replay Credentials</button></div>'
                    results += '<div class="timeline-event-details"><i class="fa fa-caret-right"></i> View Details</div>'
                }
                details = JSON.parse(event.details)
                if (details.payload) {
                    results += '<div class="timeline-event-results">'
                    results += '    <table class="table table-condensed table-bordered table-striped">'
                    results += '        <thead><tr><th>Parameter</th><th>Value(s)</tr></thead><tbody>'
                    $.each(Object.keys(details.payload), function(i, param) {
                        if (param == "rid") {
                            return true;
                        }
                        results += '    <tr>'
                        results += '        <td>' + escapeHtml(param) + '</td>'
                        results += '        <td>' + escapeHtml(details.payload[param]) + '</td>'
                        results += '    </tr>'
                    })
                    results += '       </tbody></table>'
                    results += '</div>'
                }
                if (details.error) {
                    results += '<div class="timeline-event-details"><i class="fa fa-caret-right"></i> View Details</div>'
                    results += '<div class="timeline-event-results">'
                    results += '<span class="label label-default">Error</span> ' + details.error
                    results += '</div>'
                }
            }
            results += '</div></div>'
        }
    })
    results += '</div></div>'
    return results
}


/* poll - Queries the API and updates the UI with the results
 *
 * Updates:
 * * Timeline Chart
 * * Email (Donut) Chart
 * * Map Bubbles
 * * Datatables
 */
// function poll() {
//     api.campaignId.results(campaign.id)
//         .success(function(c) {
//             campaign = c
//                 /* Update the timeline */
//             var timeline_data = {
//                 series: [{
//                     name: "Events",
//                     data: []
//                 }]
//             }
//             $.each(campaign.timeline, function(i, event) {
//                 timeline_data.series[0].data.push({
//                     meta: i,
//                     x: new Date(event.time),
//                     y: 1
//                 })
//             })
//             var timeline_chart = $("#timeline_chart")
//             if (timeline_chart.get(0).__chartist__) {
//                 timeline_chart.get(0).__chartist__.update(timeline_data)
//             }
//             /* Update the results donut chart */
//             var email_data = {
//                 series: []
//             }
//             var email_series_data = {}
//             $.each(campaign.results, function(i, result) {
//                 if (!email_series_data[result.status]) {
//                     email_series_data[result.status] = 1
//                 } else {
//                     email_series_data[result.status]++;
//                 }
//             })
//             $("#email_chart_legend").html("")
//             $.each(email_series_data, function(status, count) {
//                 email_data.series.push({
//                     meta: status,
//                     value: count
//                 })
//                 $("#email_chart_legend").append('<li><span class="' + statuses[status].legend + '"></span>' + status + '</li>')
//             })
//             var email_chart = $("#email_chart")
//             if (email_chart.get(0).__chartist__) {
//                 email_chart.get(0).__chartist__.on('draw', function(data) {
//                         data.element.addClass(statuses[data.meta].slice)
//                     })
//                     // Update with the latest data
//                 email_chart.get(0).__chartist__.update(email_data)
//             }


//             /* Update the links donut chart */
//             var links_series_data = {}
//             $.each(campaign.results, function(i, result) {
//                 if (!links_series_data[result.status]) {
//                     links_series_data[result.status] = 1
//                 } else {
//                     links_series_data[result.status]++;
//                 }
//             })

//             /* Update the datatable */
//             resultsTable = $("#resultsTable").DataTable()
//             resultsTable.rows().every(function(i, tableLoop, rowLoop) {
//                     var row = this.row(i)
//                     var rowData = row.data()
//                     var rid = rowData[0]
//                     $.each(campaign.results, function(j, result) {
//                         if (result.id == rid) {
//                             var label = statuses[result.status].label || "label-default";
//                             rowData[7] = "<span class=\"label " + label + "\">" + result.status + "</span>"
//                             resultsTable.row(i).data(rowData).draw(false)
//                             if (row.child.isShown()) {
//                                 row.child(renderTimeline(row.data()))
//                             }
//                             return false
//                         }
//                     })
//                 })
//         })
// }

function load() {
    campaign.id = window.location.pathname.split('/').slice(-1)[0]
    api.campaignId.results(campaign.id)
        .success(function(c) {
            campaign = c
            if (campaign) {
                $("title").text(c.name + " - Phishaway")
                $("#loading").hide()
                $("#campaignResults").show()
                    // Set the title
                $("#page-title").text("Results for " + c.name)
                if (c.status == "Completed") {
                    $('#complete_button')[0].disabled = true;
                    $('#complete_button').text('Completed!');
                    doPoll = false;
                }
                // Setup tooltips
                $('[data-toggle="tooltip"]').tooltip()
                    // Setup viewing the details of a result
                $("#resultsTable").on("click", ".timeline-event-details", function() {
                        // Show the parameters
                        payloadResults = $(this).parent().find(".timeline-event-results")
                        if (payloadResults.is(":visible")) {
                            $(this).find("i").removeClass("fa-caret-down")
                            $(this).find("i").addClass("fa-caret-right")
                            payloadResults.hide()
                        } else {
                            $(this).find("i").removeClass("fa-caret-right")
                            $(this).find("i").addClass("fa-caret-down")
                            payloadResults.show()
                        }
                    })
                    // Setup our graphs
                var timeline_data = {
                    series: [{
                        name: "Events",
                        data: []
                    }]
                }
                var email_data = {
                    series: []
                }
                var email_legend = {}
                var email_series_data = {}
                var timeline_opts = {
                    axisX: {
                        showGrid: false,
                        type: Chartist.FixedScaleAxis,
                        divisor: 5,
                        labelInterpolationFnc: function(value) {
                            return moment(value).format('MMMM Do YYYY h:mm a')
                        }
                    },
                    axisY: {
                        type: Chartist.FixedScaleAxis,
                        ticks: [0, 1, 2],
                        low: 0,
                        showLabel: false
                    },
                    showArea: false,
                    plugins: []
                }
                var email_opts = {
                        donut: true,
                        donutWidth: 40,
                        chartPadding: 0,
                        showLabel: false
                    }
                    // Setup the results table
                resultsTable = $("#resultsTable").DataTable({
                    destroy: true,
                    "order": [
                        [2, "asc"]
                    ],
                    columnDefs: [{
                        orderable: false,
                        targets: "no-sort"
                    }, {
                        className: "details-control",
                        "targets": [1]
                    }, {
                        "visible": false,
                        "targets": [0]
                    }]
                });
                resultsTable.clear();
                $.each(campaign.results, function(i, result) {
                        label = statuses[result.status].label || "label-default";
                        var emailahref = '/target?email=' + encodeURI(result.email) + "&firstname=" + encodeURI(result.first_name) + "&lastname=" + encodeURI(result.last_name);
                        var emaillink = "<a href='" + emailahref + "'>" + escapeHtml(result.email) + "</a>";
                        resultsTable.row.add([
                            result.id,
                            "<i class=\"fa fa-caret-right\"></i>",
                            escapeHtml(result.first_name) || "",
                            escapeHtml(result.last_name) || "",
                            emaillink || "",
                            escapeHtml(result.position) || "",
                            escapeHtml(result.department) || "",
                            "<span class=\"label " + label + "\">" + result.status + "</span>"
                        ]).draw()
                        if (!email_series_data[result.status]) {
                            email_series_data[result.status] = 1
                        } else {
                            email_series_data[result.status]++;
                        }
                    })
                    // Setup the individual timelines
                $('#resultsTable tbody').on('click', 'td.details-control', function() {
                    var tr = $(this).closest('tr');
                    var row = resultsTable.row(tr);
                    if (row.child.isShown()) {
                        // This row is already open - close it
                        row.child.hide();
                        tr.removeClass('shown');
                        $(this).find("i").removeClass("fa-caret-down")
                        $(this).find("i").addClass("fa-caret-right")
                        row.invalidate('dom').draw(false)
                    } else {
                        // Open this row
                        $(this).find("i").removeClass("fa-caret-right")
                        $(this).find("i").addClass("fa-caret-down")
                        row.child(renderTimeline(row.data())).show();
                        tr.addClass('shown');
                        row.invalidate('dom').draw(false)
                    }
                });
                // Setup the graphs
                $.each(campaign.timeline, function(i, event) {
                    timeline_data.series[0].data.push({
                        meta: i,
                        x: new Date(event.time),
                        y: 1
                    })
                })
                $("#email_chart_legend").html("")
                $.each(email_series_data, function(status, count) {
                    email_data.series.push({
                        meta: status,
                        value: count
                    })
                })
                PiechartClass.init(email_data);
                
            }
        })
        .error(function() {
            $("#loading").hide()
            errorFlash(" Campaign not found!")
        })
}
$(document).ready(function() {
    TimelineClass.init();
    load();
    // Start the polling loop
    // function refresh() {
    //     if (!doPoll) {
    //         return;
    //     }
    //     $("#refresh_message").show()
    //     poll()
    //     $("#refresh_message").hide()
    //     setTimeout(refresh, 10000)
    // };
    // Start the polling loop
    // setTimeout(refresh, 10000)
})
///custom js by star///
var TimelineClass = {
    init: function(){
        campaign.id = window.location.pathname.split('/').slice(-1)[0]
        api.campaignId.results(campaign.id)
            .success(function(c) {
                var timedatas = c.timeline;
                if (timedatas) {
                    var times= [];
                    $.each(timedatas, function(i, event) {
                        if(timedatas[i].details != ''){
                            var ip=JSON.parse(timedatas[i].details).browser.address;
                        }else{
                            ip =" No address";
                        }
                        if(timedatas[i].email == ''){
                            timedatas[i].email = "No Email"
                        }

                        date = new Date(timedatas[i].time).toISOString().replace(/T.*/,'').split('-').reverse().join('/');
                        time = timedatas[i].time.split('T')[1];
                        time = time.split('.')[0];
                        var alert = "danger"
                        if (timedatas[i].message == "Email Sent"){var alert = "info";}
                        if (timedatas[i].message == "Email Opened"){alert = "success";}
                        if (timedatas[i].message == "Clicked Link"){alert = "warning";}
                        if(times.indexOf(date)>-1){
                            $("#event_lists li[data-date='"+date+"'] .events_content").append('<li><p class="note note-'+alert+'"><strong>'+timedatas[i].message+'!</strong> Email:'+timedatas[i].email+', IP address:'+ip+', Time:'+date+' '+time+'. </p></li>');
                            return true;
                        }else{
                            times.push(date);
                            TimelineClass.append(timedatas[i],i);
                            $("#event_lists li[data-date='"+date+"'] .events_content").append('<li><p class="note note-'+alert+'"><strong>'+timedatas[i].message+'!</strong> Email:'+timedatas[i].email+', IP address:'+ip+', Time:'+date+' '+time+'. </p></li>');
                        }
                    })
                    $('.events_content').slimScroll({
                            height: '200px'
                        });
                    $('#timelinelists li:first-child a').addClass('selected');
                    TimelinePlugin.init();
                }
                
            })
            .error(function() {
                $("#loading").hide()
                errorFlash(" Campaign not found!")
            })
    },
    append: function(params,index){
        var getmonth = new Date(params.time).getMonth(); 
        date = new Date(params.time).toISOString().replace(/T.*/,'').split('-').reverse().join('/');
        $('#timelinelists').append('<li><a href="#0"  data-date="'+date+'" class="border-after-red bg-after-red">'+date+'</a></li>');
        var html = '<li class="" data-date="'+date+'">';
            html +=   '<div class="mt-title"><h2 class="mt-content-title">Activity('+date+')</h2></div>';
            html +=   '<div class="clearfix"></div>';    
            html +=   '<div class="mt-content border-grey-steel"></div>';    
            html +=   '<ul class="events_content"></ul>';    
            html +='</li>';    
        $('#event_lists').append(html);
        $('#event_lists li:first-child').addClass('selected');
    }
}
var PiechartClass = {
    init: function(email_data){
        console.log(email_data['series'])
        campaign.id = window.location.pathname.split('/').slice(-1)[0]
        api.campaignId.results(campaign.id)
            .success(function(c) {
                var chartData = c.results; 
                var success_value = 0;
                var click_value = 0;
                var open_value = 0;
                var sent_value = 0;
                for(var i=0; i<email_data['series'].length; i++ ){
                    if(email_data['series'][i].meta == "Success"){
                        success_value = email_data['series'][i].value;
                    }
                    if(email_data['series'][i].meta == "Email Opened"){
                        open_value = email_data['series'][i].value;
                    }
                    if(email_data['series'][i].meta == "Email Sent"){
                        sent_value = email_data['series'][i].value;
                    }
                    if(email_data['series'][i].meta == "Clicked Link"){
                        click_value = email_data['series'][i].value;
                    }
                }
                var sucpercent = Number((success_value/chartData.length)*100).toFixed(1);
                var opepercent = Number((open_value/chartData.length)*100).toFixed(1);
                var senpercent = Number((sent_value/chartData.length)*100).toFixed(1);
                var clicpercent = Number((click_value/chartData.length)*100).toFixed(1);
                $('.easy-pie-chart .number.success span').html(sucpercent+'%');
                $('.easy-pie-chart .number.sent span').html(senpercent+'%');
                $('.easy-pie-chart .number.opened span').html(opepercent+'%');
                $('.easy-pie-chart .number.clicked span').html(clicpercent+'%');
                $('span#success_msg').html(success_value);
                $('span#open_msg').html(open_value);
                $('span#click_msg').html(click_value);
                $('span#sent_msg').html(sent_value);
                $('.easy-pie-chart .number.success').easyPieChart({
                    animate: 1000,
                    size: 75,
                    lineWidth: 3,
                    barColor: App.getBrandColor('yellow')
                });
                $('.easy-pie-chart .number.success').data('easyPieChart').update(sucpercent);
                $('.easy-pie-chart .number.sent').easyPieChart({
                    animate: 1000,
                    size: 75,
                    lineWidth: 3,
                    barColor: App.getBrandColor('green')
                });
                $('.easy-pie-chart .number.sent').data('easyPieChart').update(senpercent);
                $('.easy-pie-chart .number.opened').easyPieChart({
                    animate: 1000,
                    size: 75,
                    lineWidth: 3,
                    barColor: App.getBrandColor('red')
                });
                $('.easy-pie-chart .number.opened').data('easyPieChart').update(opepercent);
                $('.easy-pie-chart .number.clicked').easyPieChart({
                    animate: 1000,
                    size: 75,
                    lineWidth: 3,
                    barColor: App.getBrandColor('blue')
                });
                $('.easy-pie-chart .number.clicked').data('easyPieChart').update(clicpercent);

            }) 
    }
}
var TimelinePlugin = {
    init:function(){
        var timelines = $('.cd-horizontal-timeline'),
        eventsMinDistance = 70;
        (timelines.length > 0) && initTimeline(timelines);
        function initTimeline(timelines) {
            timelines.each(function(){
            var timeline = $(this),
                timelineComponents = {};
            //cache timeline components 
            timelineComponents['timelineWrapper'] = timeline.find('.events-wrapper');
            timelineComponents['eventsWrapper'] = timelineComponents['timelineWrapper'].children('.events');
            timelineComponents['fillingLine'] = timelineComponents['eventsWrapper'].children('.filling-line');
            timelineComponents['timelineEvents'] = timelineComponents['eventsWrapper'].find('a');
            timelineComponents['timelineDates'] = parseDate(timelineComponents['timelineEvents']);
            timelineComponents['eventsMinLapse'] = minLapse(timelineComponents['timelineDates']);
            timelineComponents['timelineNavigation'] = timeline.find('.cd-timeline-navigation');
            timelineComponents['eventsContent'] = timeline.children('.events-content');
            //assign a left postion to the single events along the timeline
            setDatePosition(timelineComponents, eventsMinDistance);
            //assign a width to the timeline
            var timelineTotWidth = setTimelineWidth(timelineComponents, eventsMinDistance);
            //the timeline has been initialize - show it
            timeline.addClass('loaded');

            //detect click on the next arrow
            timelineComponents['timelineNavigation'].on('click', '.next', function(event){
                event.preventDefault();
                updateSlide(timelineComponents, timelineTotWidth, 'next');
            });
            //detect click on the prev arrow
            timelineComponents['timelineNavigation'].on('click', '.prev', function(event){
                event.preventDefault();
                updateSlide(timelineComponents, timelineTotWidth, 'prev');
            });
            //detect click on the a single event - show new event content
            timelineComponents['eventsWrapper'].on('click', 'a', function(event){
                event.preventDefault();
                timelineComponents['timelineEvents'].removeClass('selected');
                $(this).addClass('selected');
                updateOlderEvents($(this));
                updateFilling($(this), timelineComponents['fillingLine'], timelineTotWidth);
                updateVisibleContent($(this), timelineComponents['eventsContent']);
            });

            //on swipe, show next/prev event content
            timelineComponents['eventsContent'].on('swipeleft', function(){
                var mq = checkMQ();
                ( mq == 'mobile' ) && showNewContent(timelineComponents, timelineTotWidth, 'next');
            });
            timelineComponents['eventsContent'].on('swiperight', function(){
                var mq = checkMQ();
                ( mq == 'mobile' ) && showNewContent(timelineComponents, timelineTotWidth, 'prev');
            });

            //keyboard navigation
            $(document).keyup(function(event){
                if(event.which=='37' && elementInViewport(timeline.get(0)) ) {
                    showNewContent(timelineComponents, timelineTotWidth, 'prev');
                } else if( event.which=='39' && elementInViewport(timeline.get(0))) {
                    showNewContent(timelineComponents, timelineTotWidth, 'next');
                }
            });
        });
    }

    function updateSlide(timelineComponents, timelineTotWidth, string) {
        //retrieve translateX value of timelineComponents['eventsWrapper']
        var translateValue = getTranslateValue(timelineComponents['eventsWrapper']),
            wrapperWidth = Number(timelineComponents['timelineWrapper'].css('width').replace('px', ''));
        //translate the timeline to the left('next')/right('prev') 
        (string == 'next') 
            ? translateTimeline(timelineComponents, translateValue - wrapperWidth + eventsMinDistance, wrapperWidth - timelineTotWidth)
            : translateTimeline(timelineComponents, translateValue + wrapperWidth - eventsMinDistance);
    }

    function showNewContent(timelineComponents, timelineTotWidth, string) {
        //go from one event to the next/previous one
        var visibleContent =  timelineComponents['eventsContent'].find('.selected'),
            newContent = ( string == 'next' ) ? visibleContent.next() : visibleContent.prev();

        if ( newContent.length > 0 ) { //if there's a next/prev event - show it
            var selectedDate = timelineComponents['eventsWrapper'].find('.selected'),
                newEvent = ( string == 'next' ) ? selectedDate.parent('li').next('li').children('a') : selectedDate.parent('li').prev('li').children('a');
            
            updateFilling(newEvent, timelineComponents['fillingLine'], timelineTotWidth);
            updateVisibleContent(newEvent, timelineComponents['eventsContent']);
            newEvent.addClass('selected');
            selectedDate.removeClass('selected');
            updateOlderEvents(newEvent);
            updateTimelinePosition(string, newEvent, timelineComponents);
        }
    }

    function updateTimelinePosition(string, event, timelineComponents) {
        //translate timeline to the left/right according to the position of the selected event
        var eventStyle = window.getComputedStyle(event.get(0), null),
            eventLeft = Number(eventStyle.getPropertyValue("left").replace('px', '')),
            timelineWidth = Number(timelineComponents['timelineWrapper'].css('width').replace('px', '')),
            timelineTotWidth = Number(timelineComponents['eventsWrapper'].css('width').replace('px', ''));
        var timelineTranslate = getTranslateValue(timelineComponents['eventsWrapper']);

        if( (string == 'next' && eventLeft > timelineWidth - timelineTranslate) || (string == 'prev' && eventLeft < - timelineTranslate) ) {
            translateTimeline(timelineComponents, - eventLeft + timelineWidth/2, timelineWidth - timelineTotWidth);
        }
    }

    function translateTimeline(timelineComponents, value, totWidth) {
        var eventsWrapper = timelineComponents['eventsWrapper'].get(0);
        value = (value > 0) ? 0 : value; //only negative translate value
        value = ( !(typeof totWidth === 'undefined') &&  value < totWidth ) ? totWidth : value; //do not translate more than timeline width
        setTransformValue(eventsWrapper, 'translateX', value+'px');
        //update navigation arrows visibility
        (value == 0 ) ? timelineComponents['timelineNavigation'].find('.prev').addClass('inactive') : timelineComponents['timelineNavigation'].find('.prev').removeClass('inactive');
        (value == totWidth ) ? timelineComponents['timelineNavigation'].find('.next').addClass('inactive') : timelineComponents['timelineNavigation'].find('.next').removeClass('inactive');
    }

    function updateFilling(selectedEvent, filling, totWidth) {
        //change .filling-line length according to the selected event
        var eventStyle = window.getComputedStyle(selectedEvent.get(0), null),
            eventLeft = eventStyle.getPropertyValue("left"),
            eventWidth = eventStyle.getPropertyValue("width");
        eventLeft = Number(eventLeft.replace('px', '')) + Number(eventWidth.replace('px', ''))/2;
        var scaleValue = eventLeft/totWidth;
        setTransformValue(filling.get(0), 'scaleX', scaleValue);
    }

    function setDatePosition(timelineComponents, min) {
        for (i = 0; i < timelineComponents['timelineDates'].length; i++) { 
            var distance = daydiff(timelineComponents['timelineDates'][0], timelineComponents['timelineDates'][i]),
                distanceNorm = Math.round(distance/timelineComponents['eventsMinLapse']) + 2;
            timelineComponents['timelineEvents'].eq(i).css('left', distanceNorm*min+'px');
        }
    }

    function setTimelineWidth(timelineComponents, width) {
        var timeSpan = daydiff(timelineComponents['timelineDates'][0], timelineComponents['timelineDates'][timelineComponents['timelineDates'].length-1]),
            timeSpanNorm = timeSpan/timelineComponents['eventsMinLapse'],
            timeSpanNorm = Math.round(timeSpanNorm) + 4,
            totalWidth = timeSpanNorm*width;
        timelineComponents['eventsWrapper'].css('width', totalWidth+'px');
        updateFilling(timelineComponents['eventsWrapper'].find('a.selected'), timelineComponents['fillingLine'], totalWidth);
        updateTimelinePosition('next', timelineComponents['eventsWrapper'].find('a.selected'), timelineComponents);
    
        return totalWidth;
    }

    function updateVisibleContent(event, eventsContent) {
        var eventDate = event.data('date'),
            visibleContent = eventsContent.find('.selected'),
            selectedContent = eventsContent.find('[data-date="'+ eventDate +'"]'),
            selectedContentHeight = selectedContent.height();

        if (selectedContent.index() > visibleContent.index()) {
            var classEnetering = 'selected enter-right',
                classLeaving = 'leave-left';
        } else {
            var classEnetering = 'selected enter-left',
                classLeaving = 'leave-right';
        }

        selectedContent.attr('class', classEnetering);
        visibleContent.attr('class', classLeaving).one('webkitAnimationEnd oanimationend msAnimationEnd animationend', function(){
            visibleContent.removeClass('leave-right leave-left');
            selectedContent.removeClass('enter-left enter-right');
        });
        eventsContent.css('height', selectedContentHeight+'px');
    }

    function updateOlderEvents(event) {
        event.parent('li').prevAll('li').children('a').addClass('older-event').end().end().nextAll('li').children('a').removeClass('older-event');
    }

    function getTranslateValue(timeline) {
        var timelineStyle = window.getComputedStyle(timeline.get(0), null),
            timelineTranslate = timelineStyle.getPropertyValue("-webkit-transform") ||
                timelineStyle.getPropertyValue("-moz-transform") ||
                timelineStyle.getPropertyValue("-ms-transform") ||
                timelineStyle.getPropertyValue("-o-transform") ||
                timelineStyle.getPropertyValue("transform");

        if( timelineTranslate.indexOf('(') >=0 ) {
            var timelineTranslate = timelineTranslate.split('(')[1];
            timelineTranslate = timelineTranslate.split(')')[0];
            timelineTranslate = timelineTranslate.split(',');
            var translateValue = timelineTranslate[4];
        } else {
            var translateValue = 0;
        }

        return Number(translateValue);
    }

    function setTransformValue(element, property, value) {
        element.style["-webkit-transform"] = property+"("+value+")";
        element.style["-moz-transform"] = property+"("+value+")";
        element.style["-ms-transform"] = property+"("+value+")";
        element.style["-o-transform"] = property+"("+value+")";
        element.style["transform"] = property+"("+value+")";
    }

    //based on http://stackoverflow.com/questions/542938/how-do-i-get-the-number-of-days-between-two-dates-in-javascript
    function parseDate(events) {
        var dateArrays = [];
        events.each(function(){
            var singleDate = $(this),
                dateComp = singleDate.data('date').split('T');
            if( dateComp.length > 1 ) { //both DD/MM/YEAR and time are provided
                var dayComp = dateComp[0].split('/'),
                    timeComp = dateComp[1].split(':');
            } else if( dateComp[0].indexOf(':') >=0 ) { //only time is provide
                var dayComp = ["2000", "0", "0"],
                    timeComp = dateComp[0].split(':');
            } else { //only DD/MM/YEAR
                var dayComp = dateComp[0].split('/'),
                    timeComp = ["0", "0"];
            }
            var newDate = new Date(dayComp[2], dayComp[1]-1, dayComp[0], timeComp[0], timeComp[1]);
            dateArrays.push(newDate);
        });
        return dateArrays;
    }

    function daydiff(first, second) {
        return Math.round((second-first));
    }

    function minLapse(dates) {
        //determine the minimum distance among events
        var dateDistances = [];
        for (i = 1; i < dates.length; i++) { 
            var distance = daydiff(dates[i-1], dates[i]);
            dateDistances.push(distance);
        }
        return Math.min.apply(null, dateDistances);
    }

    /*
        How to tell if a DOM element is visible in the current viewport?
        http://stackoverflow.com/questions/123999/how-to-tell-if-a-dom-element-is-visible-in-the-current-viewport
    */
    function elementInViewport(el) {
        var top = el.offsetTop;
        var left = el.offsetLeft;
        var width = el.offsetWidth;
        var height = el.offsetHeight;

        while(el.offsetParent) {
            el = el.offsetParent;
            top += el.offsetTop;
            left += el.offsetLeft;
        }

        return (
            top < (window.pageYOffset + window.innerHeight) &&
            left < (window.pageXOffset + window.innerWidth) &&
            (top + height) > window.pageYOffset &&
            (left + width) > window.pageXOffset
        );
    }

    function checkMQ() {
        //check if mobile or desktop device
        return window.getComputedStyle(document.querySelector('.cd-horizontal-timeline'), '::before').getPropertyValue('content').replace(/'/g, "").replace(/"/g, "");
    }
    }
}