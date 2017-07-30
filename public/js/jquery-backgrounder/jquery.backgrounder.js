(function ($) {
    $.fn.backgrounder = function (options) {
        var defaults = {element: 'body'};
        var options = $.extend(defaults, options);
        // Get the image we're using
        var img = $(this).children('img').first();
        // Get the image source
        var src = $(img).attr('src');
        // Hide the original element
        $(this).hide();
        // Make parent relative
        w = options.element == 'body' ? $(window).width() : $(options.element).width();
        h = options.element == 'body' ? $(window).height() : $(options.element).height();
        // Create a new div
        $('<div id="backgrounder-container"></div>')
            .css({
                'position': 'absolute',
                'z-index': -100,
                'left': 0,
                'top': 0,
                'overflow': 'hidden',
                'width': w,
                'height': h
            })
            .appendTo($(options.element))
        // Create a new image
        $('<img />')
            .appendTo($('#backgrounder-container'))
            .attr('src', src)
            .css({'position': 'absolute'})
            .hide()
            .on('load', function () {
                resizeBackgrounder(this, options.element);
                $(this).fadeIn();
            })
        // Resize handler
        $(window).resize(function () {
            var newW = options.element == 'body' ? $(window).width() : $(options.element).width();
            var newH = options.element == 'body' ? $(window).height() : $(options.element).height();
            $('#backgrounder-container').css({'width': newW, 'height': newH});
            resizeBackgrounder('#backgrounder-container img:first', options.element);
        })
        // Update function
        function resizeBackgrounder(item, elem) {
            if (elem != 'body') {
                w = $(elem).width();
                h = $(elem).height();
            } else {
                w = $(window).width();
                h = $(window).height();
            }
            var ow = $(item).width();
            var oh = $(item).height();
            if (ow / oh > w / h) { // image aspect ratio is wider than browser window
                var scale = h / oh;
                $(item).attr({'width': ow * scale, 'height': oh * scale});
            } else {
                var scale = w / ow;
                $(item).attr({'width': ow * scale, 'height': oh * scale});
            }
            $(item).css({'left': -(($(item).width() - w) / 2), 'top': -(($(item).height() - h) / 2)});
        }

        // Return
        return this.each(function () {
        });
    };
})(jQuery);
