$(document).ready(function() {
    const host_ip = "192.168.12.250"
    const cp_url = "http://"+host_ip+":8081"
    const pep_url = "http://"+host_ip+":8080"

    $.ajax({
        dataType: 'json',
        url: pep_url + "/apps",
        type: 'GET'
    }).then(function(data) {

        $.each(data, function(i, app_data) {
            var col = $('<tr />');
            col.append($('<td />').text(app_data.id));
            col.append($('<td />').text("Not Implemented"));
            col.append($('<td />').text("Not Implemented"));
            col.append($('<td />').text("Not Implemented"))
            var servicePort = ""
            $.each(app_data.appInfo.metaInfo.servicePorts, function(j, port) {
                servicePort += port + " ";
            });
            col.append($('<td />').text(app_data.vNICIPAddress));
            col.append($('<td><input type="button" value="STOP" class="btn btn-warning" id="btn-stop-'+app_data.id+'" /></td>'));
            $('#app-table-body').append(col)

            $(document).on("click", "#btn-stop-"+app_data.id, function(){
                var button = $(this);
                button.attr("disabled", true);

                console.log(app_data.pid+" has been clicked")
                $.ajax({
                    type:"POST",
                    url: pep_url + "/app/"+app_data.id+"/stop",
                    contentType: "application/json",
                    success: function(json_response) {
                        console.log('success')
                    },
                    error: function(json_response) {
                        console.log('fail '+json_response)
                    },
                    complete: function() {
                        button.attr("disabled", false);
                    }
                });
                document.location.reload();
            })
        })
    });

    $.ajax({
        dataType: 'json',
        url: cp_url + "/cap/delegated",
        type: 'GET'
    }).then(function(data){
        $.each(data, function(i, cap){
            var col = $('<tr />');
            col.append($('<td />').text(cap.capabilityID));
            col.append($('<td />').text(cap.assignerID));
            col.append($('<td />').text(cap.assigneeID));
            col.append($('<td />').text(cap.capabilityName));
            col.append($('<td />').text(cap.capabilityValue));
            col.append($('<td />').text(cap.authorizeCapabilityID));
            col.append($('<td />').text(cap.grantPolicy.requesterAttribute));
            col.append($('<td />').text(cap.grantPolicy.grantCondition));
            col.append($('<td />').text(cap.grantPolicy.grantValue));
            $('#cap-delegated-list-table-body').append(col)
        })
    });

    $.ajax({
        dataType: 'json',
        url: cp_url + "/capReq/pending",
        type: 'GET'
    }).then(function (data) {
        $.each(data, function (i, cap) {
            var col = $('<tr />');
            col.append($('<td />').text(cap.request.requestID));
            col.append($('<td />').text(cap.request.requesterID));
            col.append($('<td />').text(cap.request.requesteeID));
            col.append($('<td />').text(cap.request.requestCapability));
            col.append($('<td />').text(cap.request.requestCapabilityValue));
            col.append($('<td />'));
            $('#cap-req-list-table-body').append(col);
            $.each(cap.pendingCapabilities, function (j, pendingCap){
                var pendingCapCol = $('<tr />');
                var btn_id = Math.floor( Math.random() * 1000 ) ;
                pendingCapCol.append($('<td />').text(pendingCap.capabilityID));
                pendingCapCol.append($('<td />').text(pendingCap.assignerID));
                pendingCapCol.append($('<td />').text(pendingCap.assigneeID));
                pendingCapCol.append($('<td />').text(pendingCap.capabilityName));
                pendingCapCol.append($('<td />').text(pendingCap.capabilityValue));
                pendingCapCol.append($('<td><input type="button" value="GRANT" class="btn btn-success" id="btn-cap-grant-'+btn_id+'" /></td>'));
                $('#cap-req-list-table-body').append(pendingCapCol)

                $(document).on("click", "#btn-cap-grant-"+btn_id, function () {
                    var button = $(this);
                    button.attr("disabled", true);
                    $.ajax({
                        type: "POST",
                        dataType: "json",
                        url: cp_url + "/capReq/" + cap.request.requestID + "/grant/" + pendingCap.capabilityID,
                        contentType: "application/json",
                        success: function (json_response) {
                            console.log('success')
                        },
                        error: function (json_response) {
                            console.log('fail ' + json_response)
                        },
                        complete: function () {
                            button.attr("disabled", false);
                        }
                    });
                    document.location.reload();
                })
            })
        })
    });

    $.ajax({
        dataType: 'json',
        url: cp_url + "/cap/granted",
        type: 'GET'
    }).then(function (data) {
        $.each(data, function (i, cap) {
            var col = $('<tr />');
            col.append($('<td />').text(cap.capabilityID));
            col.append($('<td />').text(cap.grantCondition));
            col.append($('<td />').text(cap.assignerID));
            col.append($('<td />').text(cap.assigneeID));
            col.append($('<td />').text(cap.capabilityName));
            col.append($('<td />').text(cap.capabilityValue));
            col.append($('<td />').text(cap.authorizeCapabilityID));
            col.append($('<td />').text(cap.grantPolicy.requesterAttribute));
            col.append($('<td />').text(cap.grantPolicy.grantCondition));
            col.append($('<td />').text(cap.grantPolicy.grantValue));
            $('#cap-granted-list-table-body').append(col)
        })
    });
});