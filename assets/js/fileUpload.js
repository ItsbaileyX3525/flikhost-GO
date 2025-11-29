var dropZone = document.getElementById('dropZone');
var dropZoneText = document.getElementById('dropZoneText');
var fileName = document.getElementById('fileName');
var replaceText = document.getElementById('replaceText');
var fileUpload = document.getElementById('fileUpload');
var uploadForm = document.getElementById('uploadForm');

var turnstileToken = '';
var turnstileWidgetId = null;
var isLoggedIn = false;

window.turnstileCallback = function(token) {
    turnstileToken = token;
};

function checkSession() {
    fetch('/api/checkSession')
        .then(function(response) {
            return response.json();
        })
        .then(function(data) {
            isLoggedIn = data.loggedIn;
            
            var accountSection = document.getElementById('account-section');
            var createAccText = document.getElementById('createAccText');
            var turnstileContainer = document.getElementById('turnstile-container');
            var usernameDisplay = document.getElementById('username-display');
            
            if (data.loggedIn) {
                if (accountSection) accountSection.style.display = 'block';
                if (usernameDisplay) usernameDisplay.textContent = '@' + data.username;
                if (createAccText) createAccText.style.display = 'none';
                if (turnstileContainer) turnstileContainer.style.display = 'none';
            } else {
                if (accountSection) accountSection.style.display = 'none';
                if (createAccText) createAccText.style.display = 'block';
                if (turnstileContainer) turnstileContainer.style.display = 'block';
                setTimeout(initTurnstile, 100);
            }
        })
        .catch(function(error) {
            var accountSection = document.getElementById('account-section');
            var createAccText = document.getElementById('createAccText');
            if (accountSection) accountSection.style.display = 'none';
            if (createAccText) createAccText.style.display = 'block';
            setTimeout(initTurnstile, 100);
        });
}

function initTurnstile() {
    var turnstileContainer = document.getElementById('turnstile-container');
    
    if (!window.turnstile) {
        setTimeout(initTurnstile, 500);
        return;
    }
    
    if (turnstileContainer && !isLoggedIn) {
        try {
            turnstileWidgetId = turnstile.render('#turnstile-container', {
                sitekey: '0x4AAAAAACDFfOKm7uvwfqiR',
                theme: 'dark',
                callback: window.turnstileCallback
            });
        } catch (e) {
            console.error('Turnstile error:', e);
        }
    }
}

dropZone.addEventListener('dragover', function(event) {
    event.preventDefault();
    event.stopPropagation();
    dropZone.classList.add('hover');
});

dropZone.addEventListener('dragleave', function(event) {
    event.preventDefault();
    event.stopPropagation();
    dropZone.classList.remove('hover');
});

dropZone.addEventListener('drop', function(event) {
    event.preventDefault();
    event.stopPropagation();
    
    var files = event.dataTransfer.files;
    if (files.length > 0) {
        fileUpload.files = files;
        fileUpload.dispatchEvent(new Event('change'));
    }
});

dropZone.addEventListener('click', function() {
    fileUpload.click();
});

fileUpload.addEventListener('change', function(event) {
    var file = event.target.files[0];
    if (!file) return;
    
    dropZoneText.style.display = 'none';
    fileName.textContent = file.name;
    fileName.style.display = 'block';
    dropZone.classList.add('filed');
    dropZone.classList.remove('hover');
    replaceText.style.cssText = 'display:inline-block; text-align:center; width:100%;';
});

uploadForm.addEventListener('submit', function(event) {
    event.preventDefault();
    
    if (!fileUpload.files || fileUpload.files.length === 0) {
        notify('Please select a file!', 'error');
        return;
    }
    
    if (!isLoggedIn && !turnstileToken) {
        notify('Please complete the captcha!', 'error');
        return;
    }
    
    var formData = new FormData();
    formData.append('file', fileUpload.files[0]);
    if (turnstileToken) {
        formData.append('token', turnstileToken);
    }
    
    var submitButton = uploadForm.querySelector('button[type="submit"]');
    var originalText = submitButton.textContent;
    submitButton.disabled = true;
    submitButton.textContent = 'Uploading...';
    
    fetch('/api/uploadFile', {
        method: 'POST',
        body: formData
    })
    .then(function(response) {
        return response.json();
    })
    .then(function(data) {
        if (data.status === 'success') {
            notify(data.message, 'success');
            var directLink = window.location.origin + data.path;
            showLinks(directLink);
            resetForm();
        } else {
            notify(data.message || 'Upload failed', 'error');
        }
    })
    .catch(function(error) {
        notify('An error occurred during upload', 'error');
    })
    .finally(function() {
        submitButton.disabled = false;
        submitButton.textContent = originalText;
    });
});

function showLinks(directLink) {
    var linkContainer = document.getElementById('link-container');
    var directLinkEl = document.getElementById('file-link-direct');
    var copyBtn1 = document.getElementById('copy-button-1');
    
    if (linkContainer) {
        linkContainer.style.display = 'flex';
    }
    
    if (directLinkEl) {
        directLinkEl.textContent = directLink;
    }
    
    if (copyBtn1) {
        copyBtn1.onclick = function() {
            copyText(directLink);
        };
    }
}

function resetForm() {
    dropZoneText.style.display = 'block';
    fileName.style.display = 'none';
    fileName.textContent = '';
    dropZone.classList.remove('filed');
    replaceText.style.display = 'none';
    fileUpload.value = '';
    
    if (window.turnstile && !isLoggedIn && turnstileWidgetId !== null) {
        turnstile.reset(turnstileWidgetId);
        turnstileToken = '';
    }
}

function copyText(text) {
    navigator.clipboard.writeText(text)
        .then(function() {
            notify('Copied to clipboard!', 'success');
        })
        .catch(function() {
            notify('Failed to copy', 'error');
        });
}

document.addEventListener('DOMContentLoaded', checkSession);
