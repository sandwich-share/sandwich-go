//Custom JS code to do some stuff
$(window).load(function(){
	var query_result_template = _.template($("#search_results_template").html())
	var peer_list_template = _.template($("#peer_list_template").html())
	var peer_results_template = _.template($("#peer_results_template").html())
	var peer
	var start;
	var gotall = false;
	var loading = false;
	var step = 100;
	var ip;
	var port;
	var path;
	//Initializes the version number
	$.get("/version", {}, function(data) {
		$("#version").html(data);
	});
	//Initializes the settings
	$.getJSON("/settings", {}, function(data) {
		$("#localport").val(data["LocalServerPort"]);
		$("#dirname").val(data["SandwichDirName"]);
		if (data["DontOpenBrowserOnStart"]) {
			$("#openbrowser").removeAttr("checked");
		} else {
			$("#openbrowser").attr("checked", "checked");
		}
	});
	update_peers = function(){
		$.getJSON("/peers", {}, function(data) {
			data = _.sortBy(data, function(x){return x["IP"]});
			$("#peers").html(peer_list_template({data: data}));
		})
	}
	update_peers();
	setInterval(update_peers, 10000);
	$(".peer-link").on("click", function(e) {
		e.preventDefault();
		$("#content").html($("#peer_template").html());
		$("#peer_folder_list").html("");
		$("#loading").show();
		get_peer_folders(this);
	})
	//Saves the settings
	$("#save_settings").on("click", function() {
		$.post("/settings", {localport: $("#localport").val(),
			dirname: $("#dirname").val(),
			openbrowser: $("#openbrowser").is(":checked")});
		$("#settingsModal").modal('hide');
	});
	get_peer_folders = function(clicked) {
		path = $(clicked).attr("data-folder")
		ip = $(clicked).attr("data-ip")
		port = $(clicked).attr("data-port")
		start = 0;
		gotall = false;
		loading = false;
		$.getJSON("/peer", {peer: ip, path: path, start: start, step: step}, function(data) {
			if (data.length < 100) { gotall = true }
			$("#peer_folder_list").html(peer_results_template({data: data, ip: ip, port: port, path: path}));
			$("#loading").hide();
		});
	}
	$(document).on("click", ".folder-link", function(e) {
		e.preventDefault();
		get_peer_folders(this);
	});
	$("#query_form").on("submit", function(e) {
		start = 0;
		gotall = false;
		loading = false;
		e.preventDefault();
		$("#content").html($("#search_template").html());
		$("#file_list").html("");
		$("#loading").show();
		$.getJSON("/search", {search: $(this).find("input[type=text]").val(),
			regex: $(this).find("input[name=regex]").is(":checked"),
			start: start, step: step}, function(data) {
				if (data.length < step) { gotall = true }
				$("#file_list").html(query_result_template({data: data}));
				$("#loading").hide()
			});
	});
	$("#killbtn").on("click", function(e) {
		e.preventDefault();
		if (confirm("Are you sure you want to shut down?")) {
			$.get("/kill")
		}
		return false;
	});
	$(document).on("click", ".dl-link", function(e) {
		e.preventDefault();
		$.get("/download", {ip: $(this).attr("data-ip"), file: $(this).attr("data-file"), type: $(this).attr("data-type")});
		$(".top-right").notify({message: {text: "Download started..."}}).show();
	});
	$(window).scroll(function() {
		if (!loading && !gotall && $(window).scrollTop() >= $(document).height() - $(window).height() - 30) {
			loading = true;
			start += step;
			if ($("#file_list").length) {
				$.getJSON("/search", {start: start, step: step}, function(data){
					loading = false;
					if (data.length < 100) { gotall = true }
					$("#file_list").append(query_result_template({data: data}));
				});
			} else if ($("#peer_folder_list").length) {
				$.getJSON("/peer", {start: start, step: step, peer: ip, path: path}, function(data){
					loading = false;
					if (data.length < 100) { gotAll = true }
					$("#peer_folder_list").append(peer_results_template({data: data, ip: ip, port: port}));
				});
			}
		}
	});
})

