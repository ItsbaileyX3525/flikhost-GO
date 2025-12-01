document.addEventListener('DOMContentLoaded', function() {
    loadUserImages();
});

function loadUserImages() {
    fetch('/api/getUserImages')
        .then(function(response) {
            return response.json();
        })
        .then(function(data) {
            if (data.status === 'success') {
                var usernameDisplay = document.getElementById('usernameDisplay');
                if (usernameDisplay) {
                    usernameDisplay.textContent = '@' + data.username;
                }
                renderImages(data.images);
            } else {
                // Not logged in, redirect to signup
                window.location.href = '/signup';
            }
        })
        .catch(function(error) {
            console.error('Error loading images:', error);
            window.location.href = '/signup';
        });
}

function renderImages(images) {
    var container = document.getElementById('imageContainer');
    if (!container) return;
    
    container.innerHTML = '';
    
    if (!images || images.length === 0) {
        container.innerHTML = '<p class="empty-state">No images uploaded yet :( <br><a href="/">upload one!</a></p>';
        return;
    }
    
    images.forEach(function(image) {
        var postDiv = document.createElement('div');
        postDiv.className = 'postContainer';
        postDiv.id = 'image-' + image.id;
        
        var copy = window.location.host + image.path;
        
        postDiv.innerHTML = 
            '<div class="imageContainer">' +
                '<img src="' + escapeHtml(image.path) + '" alt="' + escapeHtml(image.name) + '">' +
            '</div>' +
            '<div class="controlContainer">' +
                '<button onclick="copyText(\'' + escapeHtml(copy) + '\')">Copy Link</button>' +
                '<button class="delete-btn" onclick="deleteImage(' + image.id + ')">Delete</button>' +
                '<p>Uploaded: ' + escapeHtml(image.uploadDate) + '</p>' +
            '</div>';
        
        container.appendChild(postDiv);
    });
}

function copyText(text) {
    navigator.clipboard.writeText(text)
        .then(function() {
            notify('Copied image!', 'success');
        })
        .catch(function() {
            notify('Failed to copy', 'error');
        });
}

function deleteImage(imageId) {
    var formData = new FormData();
    formData.append('image_id', imageId);
    
    fetch('/api/deleteImage', {
        method: 'POST',
        body: formData
    })
    .then(function(response) {
        return response.json();
    })
    .then(function(result) {
        if (result.success) {
            var imageElement = document.getElementById('image-' + imageId);
            if (imageElement) {
                imageElement.style.display = 'none';
            }
            notify('Image deleted!', 'info');
        } else {
            notify(result.message || 'Failed to delete image', 'error');
        }
    })
    .catch(function(error) {
        console.error('AJAX Error:', error);
        notify('An error occurred, please try again.', 'error');
    });
}

function escapeHtml(text) {
    if (!text) return '';
    var div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
