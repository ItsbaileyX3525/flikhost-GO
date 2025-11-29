const dropZone = document.getElementById('dropZone');
const preImage = document.getElementById('preImage');
const replaceText = document.getElementById('replaceText');
const imageUpload = document.getElementById('imageUpload');
const uploadForm = document.getElementById('uploadForm');
let turnstileToken = '';
let turnstileWidgetId = null;
let isLoggedIn = false;

window.turnstileCallback = function(token) {
    turnstileToken = token;
};

async function checkSession() {
    try {
        const response = await fetch('/api/checkSession');
        const data = await response.json();
        isLoggedIn = data.loggedIn;
        
        const accountSection = document.getElementById('account-section');
        const createAccText = document.getElementById('createAccText');
        const turnstileContainer = document.getElementById('turnstile-container');
        
        if (data.loggedIn) {
            if (accountSection) accountSection.style.display = 'block';
            if (document.getElementById('username-display')) {
                document.getElementById('username-display').textContent = '@' + data.username;
            }
            if (createAccText) createAccText.style.display = 'none';
            if (turnstileContainer) turnstileContainer.style.display = 'none';
        } else {
            if (accountSection) accountSection.style.display = 'none';
            if (createAccText) createAccText.style.display = 'block';
            if (turnstileContainer) turnstileContainer.style.display = 'block';
            setTimeout(initTurnstile, 100);
        }
    } catch (error) {
        const accountSection = document.getElementById('account-section');
        const createAccText = document.getElementById('createAccText');
        if (accountSection) accountSection.style.display = 'none';
        if (createAccText) createAccText.style.display = 'block';
        setTimeout(initTurnstile, 100);
    }
}

function initTurnstile() {
    const turnstileContainer = document.getElementById('turnstile-container');
    if (window.turnstile && turnstileContainer && !isLoggedIn) {
        try {
            turnstileWidgetId = turnstile.render('#turnstile-container', {
                sitekey: '0x4AAAAAACDFfOKm7uvwfqiR',
                theme: 'dark',
                callback: window.turnstileCallback
            });
        } catch (e) {
            console.error('Turnstile render error:', e);
        }
    } else if (!window.turnstile) {
        setTimeout(initTurnstile, 500);
    }
}

dropZone.addEventListener('dragover', function(event) {
    event.preventDefault();
    event.stopPropagation();
    this.classList.add("hover");
});

dropZone.addEventListener('dragleave', function(event) {
    event.preventDefault();
    event.stopPropagation();
    this.classList.remove("hover");
});

dropZone.addEventListener('drop', function(event) {
    event.preventDefault();
    event.stopPropagation();

    const files = event.dataTransfer.files;
    if (files.length > 0) {
        imageUpload.files = files;
        imageUpload.dispatchEvent(new Event('change'));
    }
});

imageUpload.addEventListener('change', function(event) {
    const files = event.target.files;
    if (files.length > 0) {
        const file = files[0];
        const fileType = file.type;

        if (fileType.startsWith('image/')) {
            const reader = new FileReader();

            reader.onload = function(e) {
                preImage.src = e.target.result;
                preImage.style.display = 'block';
                dropZone.classList.add("filed");
                dropZone.classList.remove("hover");
                replaceText.style.cssText = 'display:inline-block; text-align: center; width: 100%;';
            };

            reader.readAsDataURL(file);
        } else {
            notify('Not an image!', "error");
            imageUpload.value = '';
        }
    }
});

dropZone.addEventListener('click', function() {
    imageUpload.click();
});

uploadForm.addEventListener('submit', async function(event) {
    event.preventDefault();

    if (!imageUpload.files || imageUpload.files.length === 0) {
        notify('Please select an image!', 'error');
        return;
    }

    if (!isLoggedIn && !turnstileToken) {
        notify('Please complete the captcha!', 'error');
        return;
    }

    const formData = new FormData();
    formData.append('image', imageUpload.files[0]);
    if (turnstileToken) {
        formData.append('token', turnstileToken);
    }

    const submitButton = uploadForm.querySelector('button[type="submit"]');
    const originalText = submitButton.textContent;
    submitButton.disabled = true;
    submitButton.textContent = 'Uploading...';

    try {
        const response = await fetch('/api/uploadImage', {
            method: 'POST',
            body: formData
        });

        const data = await response.json();

        if (data.status === 'success') {
            notify(data.message, 'success');
            
            const linkContainer = document.getElementById('link-container');
            if (linkContainer) linkContainer.style.display = 'flex';
            
            const directLink = data.path || '/files/' + imageUpload.files[0].name;
            const wrapperId = data.fileID || 'unknown';
            const wrapperLink = `https://flik.host/image.php?id=${wrapperId}`;
            
            const directLinkEl = document.getElementById('image-link-direct');
            const wrapperLinkEl = document.getElementById('image-link-wrapper');
            const copyBtn1 = document.getElementById('copy-button-1');
            const copyBtn2 = document.getElementById('copy-button-2');
            
            if (directLinkEl) directLinkEl.textContent = directLink;
            if (copyBtn1) copyBtn1.onclick = function() { copyText(directLink); };
            
            if (wrapperLinkEl) wrapperLinkEl.textContent = wrapperLink;
            if (copyBtn2) copyBtn2.onclick = function() { copyText(wrapperLink); };

            preImage.style.display = 'none';
            preImage.src = '';
            dropZone.classList.remove('filed');
            replaceText.style.display = 'none';
            imageUpload.value = '';
            
            if (window.turnstile && !isLoggedIn && turnstileWidgetId !== null) {
                turnstile.reset(turnstileWidgetId);
                turnstileToken = '';
            }
        } else {
            notify(data.message || 'Upload failed', 'error');
        }
    } catch (error) {
        notify('An error occurred during upload', 'error');
    } finally {
        submitButton.disabled = false;
        submitButton.textContent = originalText;
    }
});

function copyText(text) {
    navigator.clipboard.writeText(text)
        .then(() => notify('Copied to clipboard!', 'success'))
        .catch(() => notify('Failed to copy', 'error'));
}

document.addEventListener('DOMContentLoaded', checkSession);
