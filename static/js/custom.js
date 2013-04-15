/**
 * bootstrap-notify.js v1.0.0
 * --
 * Copyright 2012 Nijiko Yonskai <nijikokun@gmail.com>
 * Copyright 2012 Goodybag, Inc.
 * --
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

(function ($) {
	var Notification = function (element, options) {
		// Element collection
		this.$element = $(element);
		this.$note		= $('<div class="alert"></div>');
		this.options	= $.extend(true, {}, $.fn.notify.defaults, options);
		this._link		= null;

		// Setup from options
		if (this.options.transition)
	if (this.options.transition === 'fade')
	this.$note.addClass('in').addClass(this.options.transition);
	else this.$note.addClass(this.options.transition);
		else this.$note.addClass('fade').addClass('in');

		if (this.options.type)
	this.$note.addClass('alert-' + this.options.type);
		else this.$note.addClass('alert-success');

		if (this.options.message)
	if (typeof this.options.message === 'string')
		this.$note.html(this.options.message);
	else if (typeof this.options.message === 'object')
		if (this.options.message.html)
			this.$note.html(this.options.message.html);
		else if (this.options.message.text)
			this.$note.text(this.options.message.text);

		if (this.options.closable)
			this._link = $('<a class="close pull-right">&times;</a>'),
				$(this._link).on('click', $.proxy(Notification.onClose, this)),
					this.$note.prepend(this._link);

		return this;
	};

	Notification.onClose = function () {
		this.options.onClose();
		$(this.$note).remove();
		this.options.onClosed();
	};

	Notification.prototype.show = function () {
		if (this.options.fadeOut.enabled)
			this.$note.delay(this.options.fadeOut.delay || 3000).fadeOut('slow', $.proxy(Notification.onClose, this));

		this.$element.append(this.$note);
		this.$note.alert();
	};

	Notification.prototype.hide = function () {
		if (this.options.fadeOut.enabled)
			this.$note.delay(this.options.fadeOut.delay || 3000).fadeOut('slow', $.proxy(Notification.onClose, this));
		else Notification.onClose.call(this);
	};

	$.fn.notify = function (options) {
		return new Notification(this, options);
	};

	$.fn.notify.defaults = {
		type: 'success',
		closable: true,
		transition: 'fade',
		fadeOut: {
			enabled: true,
			delay: 3000
		},
		message: null,
		onClose: function () {},
		onClosed: function () {}
	}
})(window.jQuery);

//Custom JS code to do some stuff
$(window).load(function(){
	var query_result_template = _.template($("#results_template").html())
	var start;
	var gotall = false;
	var loading = false;
	var step = 100;
	$("#query_form").on("submit", function(e) {
		start = 0;
		gotall = false;
		loading = false;
		e.preventDefault();
		$.ajax({
			dataType: "json",
			url: "/search",
			data: {search:
			$(this).find("input[type=text]").val(),
			regex: $(this).find("input[name=regex]").is(":checked"),
			start: 0, step: step}, 
			success: function(data){
				if (data.length < step) { gotall = true }
				$("#file_list").html(query_result_template({data: data}));
				$("#loading").hide()
				add_dl_listeners();
			},
			beforeSend: function(){
				$("#file_list").html("");
				$("#loading").show();
			}});
	})
	$("#killbtn").on("click", function(e) {
		e.preventDefault();
		if (confirm("Are you sure you want to shut down?")) {
			$.get("/kill")
		}
		return false;
	});
	add_dl_listeners = function(){ $(".dl-link").on("click", function(e) {
		e.preventDefault();
		$.get("/download", {ip: $(this).attr("data-ip"), file: $(this).attr("data-file")});
		$(".top-right").notify({message: {text: "Download started..."}}).show();
	})}
	$(window).scroll(function() {
		if (!loading && !gotall && $(window).scrollTop() >= $(document).height() - $(window).height() - 30) {
			loading = true;
			start += step;
			$.getJSON("/search", {start: start, step: step}, function(data){
				loading = false;
				if (data.length < 100) { gotall = true }
				$("#file_list").append(query_result_template({data: data}));
				add_dl_listeners();
			});
		}
	});
})

