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
