var ricky = document.getElementById('ricky');
if (ricky) {
    ricky.onclick = function() {
        if (ricky.getAttribute('data-isfired') == 0) {
            ricky.setAttribute('data-isfired', 1);
            var container = document.getElementById('video-container');
            var iframe = document.createElement('iframe');
            iframe.src = 'https://www.youtube.com/embed/dQw4w9WgXcQ?autoplay=1';
            iframe.width = '560';
            iframe.height = '315';
            iframe.frameBorder = '0';
            iframe.allow = 'accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture';
            iframe.allowFullscreen = true;
            container.appendChild(iframe);
        }

        ricky.innerHTML='lololololol try to click again';
        ricky.download='file.txt';
        ricky.href = ricky.getAttribute('data-link');
        var roller = ricky.getAttribute('data-ricky');
    };
} else {
    console.log('Element with ID "ricky" not found');
}