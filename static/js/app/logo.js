'use strict';
$(document).ready(function() {
    // Handle file uploads
    $("#logoupload").fileupload({
        dataType: "json",
        add: function(e, data) {
            var acceptFileTypes = /(png)$/i;
            var filename = data.originalFiles[0]['name']
            if (filename && !acceptFileTypes.test(filename.split(".").pop())) {
                errorFlash("Unsupported file extension (use .png)")
                return false;
            }
            data.submit()
                .success(function (result, textStatus, jqXHR) {
                    successFlash(result.message);
                })
                .error(function (jqXHR, textStatus, errorThrown) {
                    errorFlash(result.message);
                })
                .complete(function (result, textStatus, jqXHR) {
                });
        },
        done: function(e, data) {
        }
    });
});
