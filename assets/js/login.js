var loginForm = document.getElementById('loginForm');

// Check if already logged in
document.addEventListener('DOMContentLoaded', function() {
    fetch('/api/checkSession')
        .then(function(response) {
            return response.json();
        })
        .then(function(data) {
            if (data.loggedIn) {
                window.location.href = '/home';
            }
        })
        .catch(function(error) {
            // Not logged in, stay on page
        });
});

loginForm.addEventListener('submit', function(event) {
    event.preventDefault();

    var username = document.getElementById('username').value;
    var password = document.getElementById('password').value;

    if (!username || !password) {
        notify('Please fill in all fields', 'error');
        return;
    }

    var submitButton = loginForm.querySelector('button[type="submit"]');
    var originalText = submitButton.textContent;
    submitButton.disabled = true;
    submitButton.textContent = 'Logging in...';

    fetch('/api/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            username: username,
            password: password
        })
    })
    .then(function(response) {
        return response.json();
    })
    .then(function(data) {
        if (data.status === 'success') {
            notify(data.message, 'success');
            setTimeout(function() {
                window.location.href = '/home';
            }, 1000);
        } else {
            notify('Login failed: ' + data.message, 'error');
            submitButton.disabled = false;
            submitButton.textContent = originalText;
        }
    })
    .catch(function(error) {
        console.error('Error:', error);
        notify('An error occurred, please try again.', 'error');
        submitButton.disabled = false;
        submitButton.textContent = originalText;
    });
});
