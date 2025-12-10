document.addEventListener('DOMContentLoaded', function() {
    loadUserInfo();
    setupLogout();
});

function loadUserInfo() {
    fetch('/api/getUserInfo')
        .then(function(response) {
            return response.json();
        })
        .then(function(data) {
            if (data.status === 'success') {
                document.getElementById('usernameDisplay').textContent = data.username;
                document.getElementById('infoUsername').value = data.username;
                document.getElementById('infoEmail').value = data.email;
                document.getElementById('infoApiKey').value = data.apiKey || 'Not available';
                document.getElementById('infoCreated').value = data.createdAt;
            } else {
                // Not logged in, redirect to signup
                window.location.href = '/signup';
            }
        })
        .catch(function(error) {
            console.error('Error loading user info:', error);
            window.location.href = '/signup';
        });
}

function setupLogout() {
    var logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', function() {
            fetch('/api/logout', {
                method: 'POST'
            })
            .then(function(response) {
                return response.json();
            })
            .then(function(data) {
                if (data.status === 'success') {
                    notify('Logged out successfully', 'success');
                    setTimeout(function() {
                        window.location.href = '/';
                    }, 1000);
                } else {
                    notify(data.message || 'Logout failed', 'error');
                }
            })
            .catch(function(error) {
                notify('An error occurred', 'error');
            });
        });
    }
}

function copyApiKey() {
    var apiKeyInput = document.getElementById('infoApiKey');
    apiKeyInput.select();
    apiKeyInput.setSelectionRange(0, 99999);
    
    navigator.clipboard.writeText(apiKeyInput.value).then(function() {
        console.log('API Key copied to clipboard!');
    }).catch(function(err) {
        // Fallback for older browsers
        document.execCommand('copy');
        console.log('API Key copied to clipboard!');
    });
}
