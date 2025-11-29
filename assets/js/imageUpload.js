async function checkSession() {
    try {
        const response = await fetch('/api/checkSession');
        const data = await response.json();
        
        if (data.loggedIn) {
            document.getElementById('account-section').style.display = 'block';
            document.getElementById('username-display').textContent = '@' + data.username;
            document.getElementById('createAccText').style.display = 'none';
            document.getElementById('turnstile-container').style.display = 'none';
        } else {
            document.getElementById('account-section').style.display = 'none';
            document.getElementById('createAccText').style.display = 'block';
            if (window.turnstile) {
                turnstile.render('#turnstile-container', {
                    sitekey: '0x4AAAAAABCj8wt8gwneL-OU',
                    theme: 'dark',
                    callback: turnstileCallback
                });
            }
        }
    } catch (error) {
        console.error('Error checking session:', error);
        document.getElementById('account-section').style.display = 'none';
        document.getElementById('createAccText').style.display = 'block';
    }
}

document.addEventListener('DOMContentLoaded', checkSession);