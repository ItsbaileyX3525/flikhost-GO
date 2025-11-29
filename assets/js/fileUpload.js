var form = document.getElementById('upload-form');
var turnstileToken = '';
var widgetID = null;

form.addEventListener('submit', function(event) {
    event.preventDefault();
    
    if (!turnstileToken) {
        console.log('Waiting for turnstile verification');
        return;
    }
    
    var input = document.getElementById('imageUpload');
    var url = '/api/uploadImage';
    if (input && input.getAttribute('accept') !== 'image/png') {
        url = '/api/uploadFile';
    }
    
    var formData = new FormData(form);
    formData.append('token', turnstileToken);
    
    fetch(url, {
        method: 'POST',
        body: formData
    })
    .then(function(response) {
        return response.json();
    })
    .then(function(data) {
        console.log(data.message);
    })
    .finally(function() {
        turnstileToken = '';
        turnstile.reset(widgetID);
    });
});

window.onloadTurnstileCallback = function() {
    widgetID = turnstile.render('#turnstile-container', {
        sitekey: '0x4AAAAAACDFfOKm7uvwfqiR',
        callback: function(token) {
            turnstileToken = token;
        }
    });
};
